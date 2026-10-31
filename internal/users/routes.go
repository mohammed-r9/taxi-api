package users

import (
	"net/http"
	"time"

	"github.com/mohammed-r9/taxi-api/internal/auth"
	"github.com/mohammed-r9/taxi-api/internal/network"
)

const mfaRateLimitWindow = 15 * time.Minute

func Register(api *network.Api, deps Deps) {
	h := New(deps)
	limiter := network.NewRateLimiter(deps.Redis)

	network.Handle(api, http.MethodPost, "/api/v1/users/register", network.Route{
		ID:      "registerUser",
		Summary: "Register a user",
		Tag:     "Users",
		Status:  http.StatusCreated,
	}, h.Register)

	network.Handle(api, http.MethodPost, "/api/v1/users/login", network.Route{
		ID:      "loginUser",
		Summary: "Log in",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.Login)

	network.Handle(api, http.MethodGet, "/api/v1/users/me", network.Route{
		ID:      "getCurrentUser",
		Summary: "Get the authenticated user",
		Tag:     "Users",
		Status:  http.StatusOK,
		Auth:    true,
	}, h.Me, auth.RequireAuth)

	network.Handle(api, http.MethodPost, "/api/v1/users/mfa/setup", network.Route{
		ID:      "setupMFA",
		Summary: "Start MFA setup",
		Tag:     "Users",
		Status:  http.StatusOK,
		Auth:    true,
	}, h.SetupMFA,
		auth.RequireAuth,
		limiter.Middleware(rateLimitPolicy(5), network.RateLimitByUserID),
	)

	network.Handle(api, http.MethodPost, "/api/v1/users/mfa/enable", network.Route{
		ID:      "enableMFA",
		Summary: "Enable MFA with a TOTP code",
		Tag:     "Users",
		Status:  http.StatusOK,
		Auth:    true,
	}, h.EnableMFA,
		auth.RequireAuth,
		limiter.Middleware(rateLimitPolicy(5), network.RateLimitByUserID),
	)

	network.Handle(api, http.MethodPost, "/api/v1/users/mfa/verify", network.Route{
		ID:      "verifyMFA",
		Summary: "Verify an MFA challenge",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.VerifyMFA,
		limiter.Middleware(rateLimitPolicy(10), network.RateLimitByIP),
	)

	network.Handle(api, http.MethodPost, "/api/v1/users/refresh", network.Route{
		ID:      "refreshToken",
		Summary: "Refresh access token",
		Tag:     "Users",
		Status:  http.StatusOK,
	}, h.RefreshToken)
}

func rateLimitPolicy(attempts int64) network.Policy {
	return network.Policy{
		Attempts: attempts,
		Window:   mfaRateLimitWindow,
	}
}
