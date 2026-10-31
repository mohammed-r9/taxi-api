package users

import (
	"time"

	"github.com/mohammed-r9/taxi-api/internal/network"
)

type UserResponse struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Username     string     `json:"username"`
	RoleID       int64      `json:"role_id"`
	TwoFAEnabled bool       `json:"two_fa_enabled"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
}

type AuthResponse struct {
	AccessToken    string       `json:"access_token,omitempty"`
	RefreshToken   string       `json:"refresh_token,omitempty"`
	TokenType      string       `json:"token_type"`
	MFARequired    bool         `json:"mfa_required"`
	ChallengeToken string       `json:"challenge_token,omitempty"`
	User           UserResponse `json:"user"`
}

type RegisterInput struct {
	Body struct {
		Email    string `json:"email" validate:"required,email" format:"email"`
		Name     string `json:"name" validate:"required,min=2,max=100" minLength:"2" maxLength:"100"`
		Username string `json:"username" validate:"required,min=3,max=32" minLength:"3" maxLength:"32"`
		Password string `json:"password" validate:"required,min=8,max=128" minLength:"8" maxLength:"128"`
		RoleID   int64  `json:"role_id" validate:"required,gt=0" minimum:"1"`
	}
}

type RegisterOutput = network.Output[AuthResponse]

type LoginInput struct {
	Body struct {
		Email    string `json:"email" validate:"required,email" format:"email"`
		Password string `json:"password" validate:"required" minLength:"1"`
	}
}

type LoginOutput = network.Output[AuthResponse]

type MeOutput = network.Output[UserResponse]

type MFASetupOutput = network.Output[MFASetupResponse]

type MFASetupResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otpauth_url"`
}

type MFAEnableInput struct {
	Body struct {
		Code string `json:"code" validate:"required,len=6,numeric" minLength:"6" maxLength:"6" pattern:"^[0-9]{6}$"`
	}
}

type MFAEnableOutput = network.Output[MFAEnableResponse]

type MFAEnableResponse struct {
	Enabled bool `json:"enabled"`
}

type MFAVerifyInput struct {
	Body struct {
		ChallengeToken string `json:"challenge_token" validate:"required" minLength:"1"`
		Code           string `json:"code" validate:"required,len=6,numeric" minLength:"6" maxLength:"6" pattern:"^[0-9]{6}$"`
	}
}

type MFAVerifyOutput = network.Output[AuthResponse]
