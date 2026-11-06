package users

import (
	"net/http"
	"time"

	"github.com/mohammed-r9/taxi-api/internal/network"
)

const mfaRateLimitWindow = 15 * time.Minute

func Register(api *network.Api, deps Deps) {
	h := New(deps)
	mw := deps.Middleware

	api.Handle(http.MethodPost, "/api/v1/users/register", network.Route{
		ID:      "registerUser",
		Summary: "Register a user",
		Tag:     "Users",
		Status:  http.StatusCreated,
	}, h.Register)

	api.Handle(http.MethodPost, "/api/v1/users/login", network.Route{
		ID:      "loginUser",
		Summary: "Log in",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.Login)

	api.HandleAuthed(http.MethodGet, "/api/v1/users/me", network.Route{
		ID:      "getCurrentUser",
		Summary: "Get the authenticated user",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.Me, mw.Auth())

	api.HandleAuthed(http.MethodPost, "/api/v1/users/mfa/setup", network.Route{
		ID:      "setupMFA",
		Summary: "Start MFA setup",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.SetupMFA,
		mw.Auth(),
		mw.RateLimit(network.Policy{
			Attempts: 5,
			Window:   mfaRateLimitWindow,
		}, network.RateLimitByUserID),
	)

	api.HandleAuthed(http.MethodPost, "/api/v1/users/mfa/enable", network.Route{
		ID:      "enableMFA",
		Summary: "Enable MFA with a TOTP code",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.EnableMFA,
		mw.Auth(),
		mw.RateLimit(network.Policy{
			Attempts: 5,
			Window:   mfaRateLimitWindow,
		}, network.RateLimitByUserID),
	)

	api.Handle(http.MethodPost, "/api/v1/users/mfa/verify", network.Route{
		ID:      "verifyMFA",
		Summary: "Verify an MFA challenge",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.VerifyMFA,
		mw.RateLimit(network.Policy{
			Attempts: 10,
			Window:   mfaRateLimitWindow,
		}, network.RateLimitByIP),
	)

	api.Handle(http.MethodPost, "/api/v1/users/refresh", network.Route{
		ID:      "refreshToken",
		Summary: "Refresh access token",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.RefreshToken)
}
