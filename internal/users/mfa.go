package users

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mohammed-r9/taxi-api/db"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/mohammed-r9/taxi-api/internal/auth"
	"github.com/mohammed-r9/taxi-api/internal/network"
)

func (h *Handler) SetupMFA(ctx context.Context, _ *struct{}) (*MFASetupOutput, error) {
	current, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, apperrors.InvalidToken
	}

	user, err := h.Store.GetUserByID(ctx, current.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.InvalidUserID
	}
	if err != nil {
		return nil, err
	}
	if user.TwoFaEnabled {
		return nil, apperrors.MFAAlreadyEnabled
	}

	key, err := auth.GenerateTOTPSecret(user.Email)
	if err != nil {
		return nil, err
	}

	encryptedSecret, err := auth.EncryptOTPSecret(key.Secret(), auth.CurrentTOTPKeyVersion)
	if err != nil {
		return nil, err
	}

	version := auth.CurrentTOTPKeyVersion
	if err := h.Store.SetUserTOTPSecret(ctx, db.SetUserTOTPSecretParams{
		ID:                     user.ID,
		TwoFaSecret:            &encryptedSecret,
		TwoFaEncryptionVersion: &version,
	}); err != nil {
		return nil, err
	}

	return network.NewSuccessResponse(http.StatusOK, MFASetupResponse{
		Secret:     key.Secret(),
		OTPAuthURL: key.URL(),
	}, "mfa setup initialized"), nil
}

func (h *Handler) EnableMFA(ctx context.Context, in *MFAEnableInput) (*MFAEnableOutput, error) {
	if err := h.Validator.Validate(&in.Body); err != nil {
		return nil, err
	}

	current, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, apperrors.InvalidToken
	}

	user, err := h.Store.GetUserByID(ctx, current.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.InvalidUserID
	}
	if err != nil {
		return nil, err
	}
	if user.TwoFaEnabled {
		return nil, apperrors.MFAAlreadyEnabled
	}

	secret, err := decryptUserSecret(user)
	if err != nil {
		return nil, err
	}
	if !auth.ValidateTOTP(secret, in.Body.Code) {
		return nil, apperrors.InvalidMFACode
	}

	if err := h.Store.EnableUserTOTP(ctx, user.ID); err != nil {
		return nil, err
	}

	return network.NewSuccessResponse(http.StatusOK, MFAEnableResponse{Enabled: true}, "mfa enabled"), nil
}

func (h *Handler) VerifyMFA(ctx context.Context, in *MFAVerifyInput) (*MFAVerifyOutput, error) {
	if err := h.Validator.Validate(&in.Body); err != nil {
		return nil, err
	}

	challenge, err := auth.ValidateMFAChallengeToken(in.Body.ChallengeToken)
	if err != nil {
		return nil, apperrors.InvalidChallengeToken
	}

	key := usedMFAChallengeKey(challenge.TokenID)
	if exists, err := h.Redis.Exists(ctx, key).Result(); err != nil {
		return nil, err
	} else if exists != 0 {
		return nil, apperrors.InvalidChallengeToken
	}

	user, err := h.Store.GetUserByID(ctx, challenge.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.InvalidChallengeToken
	}
	if err != nil {
		return nil, err
	}
	if !user.TwoFaEnabled {
		return nil, apperrors.MFANotConfigured
	}

	secret, err := decryptUserSecret(user)
	if err != nil {
		return nil, err
	}
	if !auth.ValidateTOTP(secret, in.Body.Code) {
		return nil, apperrors.InvalidMFACode
	}

	used, err := h.Redis.SetNX(ctx, key, "1", auth.MFAChallengeTTL).Result()
	if err != nil {
		return nil, err
	}
	if !used {
		return nil, apperrors.InvalidChallengeToken
	}

	accessToken, err := auth.GenerateAccessToken(auth.UserAccessData{
		UserID: user.ID,
		RoleID: user.RoleID,
	})
	if err != nil {
		return nil, err
	}
	refreshToken, err := h.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return network.NewSuccessResponse(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		User:         userFromIDRow(user),
	}, "mfa verified"), nil
}

func (h *Handler) createRefreshToken(ctx context.Context, userID int64) (string, error) {
	token := auth.NewOpaqueToken()
	hash := token.Hash()
	if err := h.Store.CreateSession(ctx, db.CreateSessionParams{
		UserID:       userID,
		RefreshToken: hash[:],
		DeviceName:   "temp device name",
		ExpiresAt:    time.Now().Add(auth.RefreshTokenTTL),
	}); err != nil {
		return "", err
	}
	return token.String(), nil
}

func (h *Handler) createMFAChallenge(userID int64) (string, error) {
	token, _, err := auth.GenerateMFAChallengeToken(userID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func usedMFAChallengeKey(tokenID string) string {
	return "mfa_challenge_used:" + tokenID
}

func decryptUserSecret(user db.GetUserByIDRow) (string, error) {
	if user.TwoFaSecret == nil || user.TwoFaEncryptionVersion == nil {
		return "", apperrors.MFANotConfigured
	}
	return auth.DecryptOTPSecret(*user.TwoFaSecret, *user.TwoFaEncryptionVersion)
}
