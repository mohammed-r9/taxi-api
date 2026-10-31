package users

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mohammed-r9/taxi-api/db"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/mohammed-r9/taxi-api/internal/auth"
	"github.com/mohammed-r9/taxi-api/internal/network"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	Store     *db.Store
	Redis     *redis.Client
	Validator *network.RequestValidator
}

type Handler struct {
	Deps
}

func New(deps Deps) *Handler {
	return &Handler{Deps: deps}
}

func (h *Handler) Register(ctx context.Context, in *RegisterInput) (*RegisterOutput, error) {
	if err := h.Validator.Validate(&in.Body); err != nil {
		return nil, err
	}

	passwordHash, err := auth.HashPassword(in.Body.Password)
	if err != nil {
		return nil, err
	}

	user, err := h.Store.CreateUser(ctx, db.CreateUserParams{
		Email:    strings.ToLower(strings.TrimSpace(in.Body.Email)),
		Name:     strings.TrimSpace(in.Body.Name),
		Username: strings.ToLower(strings.TrimSpace(in.Body.Username)),
		Password: passwordHash,
		RoleID:   in.Body.RoleID,
	})
	if db.IsUniqueViolation(err) {
		return nil, apperrors.UniqueViolation
	}
	if db.IsForeignKeyViolation(err) {
		return nil, apperrors.InvalidRole
	}
	if err != nil {
		return nil, err
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

	return network.NewSuccessResponse(http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		User:         userFromCreateRow(user),
	}, "user registered"), nil
}

func (h *Handler) Login(ctx context.Context, in *LoginInput) (*LoginOutput, error) {
	if err := h.Validator.Validate(&in.Body); err != nil {
		return nil, err
	}

	user, err := h.Store.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(in.Body.Email)))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.InvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	matches, err := auth.ComparePassword(auth.ComparePasswordParams{
		Password:   in.Body.Password,
		StoredHash: user.Password,
	})
	if err != nil {
		return nil, err
	}
	if !matches {
		return nil, apperrors.InvalidCredentials
	}

	response := AuthResponse{
		TokenType: "Bearer",
		User:      userFromEmailRow(user),
	}

	if user.TwoFaEnabled {
		challenge, err := h.createMFAChallenge(user.ID)
		if err != nil {
			return nil, err
		}
		response.MFARequired = true
		response.ChallengeToken = challenge
	} else {
		accessToken, err := auth.GenerateAccessToken(auth.UserAccessData{
			UserID: user.ID,
			RoleID: user.RoleID,
		})
		if err != nil {
			return nil, err
		}
		response.AccessToken = accessToken
		refreshToken, err := h.createRefreshToken(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		response.RefreshToken = refreshToken
	}

	return network.NewSuccessResponse(http.StatusOK, response, "login successful"), nil
}

func (h *Handler) Me(ctx context.Context, _ *struct{}) (*MeOutput, error) {
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

	return network.NewSuccessResponse(http.StatusOK, userFromIDRow(user), "user retrieved"), nil
}

func (h *Handler) RefreshToken(ctx context.Context, i *RefreshInput) (*RefreshOutput, error) {
	if err := h.Validator.Validate(&i.Body); err != nil {
		return nil, err
	}

	token, err := auth.ParseOpaqueToken(i.Body.Token)
	if err != nil {
		return nil, err
	}
	hash := token.Hash()
	session, err := h.Store.GetUserSession(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.InvalidToken
		}
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, apperrors.ExpiredToken
	}

	accessToken, err := auth.GenerateAccessToken(auth.UserAccessData{
		UserID: session.UserID,
		RoleID: session.RoleID,
	})
	if err != nil {
		return nil, err
	}

	return network.NewSuccessResponse(http.StatusOK, RefreshResponse{Token: accessToken}, "token refreshed successfully"), nil
}

func userFromCreateRow(row db.CreateUserRow) UserResponse {
	return UserResponse{
		ID:           row.ID,
		Email:        row.Email,
		Name:         row.Name,
		Username:     row.Username,
		RoleID:       row.RoleID,
		TwoFAEnabled: row.TwoFaEnabled,
		VerifiedAt:   row.VerifiedAt,
		CreatedAt:    row.CreatedAt,
	}
}

func userFromEmailRow(row db.GetUserByEmailRow) UserResponse {
	return UserResponse{
		ID:           row.ID,
		Email:        row.Email,
		Name:         row.Name,
		Username:     row.Username,
		RoleID:       row.RoleID,
		TwoFAEnabled: row.TwoFaEnabled,
		VerifiedAt:   row.VerifiedAt,
		CreatedAt:    row.CreatedAt,
	}
}

func userFromIDRow(row db.GetUserByIDRow) UserResponse {
	return UserResponse{
		ID:           row.ID,
		Email:        row.Email,
		Name:         row.Name,
		Username:     row.Username,
		RoleID:       row.RoleID,
		TwoFAEnabled: row.TwoFaEnabled,
		VerifiedAt:   row.VerifiedAt,
		CreatedAt:    row.CreatedAt,
	}
}
