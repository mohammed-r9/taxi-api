package apperrors

import "net/http"

const (
	InvalidCredentials authError = iota
	InvalidToken
	ExpiredToken
	InvalidUserID
	MissingAuthorizationHeader
	MissingRequiredPermissions
	InvalidMFACode
	MFAAlreadyEnabled
	MFANotConfigured
	InvalidChallengeToken

	// new errors should be added above this line
	authErrorCount
)

var authErrorTable = [authErrorCount]errorInfo{
	InvalidCredentials: {
		message: "invalid credentials",
		code:    "INVALID_CREDENTIALS",
		status:  http.StatusUnauthorized,
	},
	InvalidToken: {
		message: "invalid token",
		code:    "INVALID_TOKEN",
		status:  http.StatusUnauthorized,
	},
	ExpiredToken: {
		message: "token expired",
		code:    "EXPIRED_TOKEN",
		status:  http.StatusUnauthorized,
	},
	InvalidUserID: {
		message: "invalid user id",
		code:    "INVALID_USER_ID",
		status:  http.StatusUnauthorized,
	},
	MissingAuthorizationHeader: {
		message: "missing authorization header",
		code:    "MISSING_AUTHORIZATION_HEADER",
		status:  http.StatusUnauthorized,
	},
	MissingRequiredPermissions: {
		message: "user does not have the required permissions",
		code:    "MISSING_REQUIRED_PERMISSIONS",
		status:  http.StatusForbidden,
	},
	InvalidMFACode: {
		message: "invalid mfa code",
		code:    "INVALID_MFA_CODE",
		status:  http.StatusUnauthorized,
	},
	MFAAlreadyEnabled: {
		message: "mfa is already enabled",
		code:    "MFA_ALREADY_ENABLED",
		status:  http.StatusConflict,
	},
	MFANotConfigured: {
		message: "mfa is not configured",
		code:    "MFA_NOT_CONFIGURED",
		status:  http.StatusBadRequest,
	},
	InvalidChallengeToken: {
		message: "invalid or expired challenge token",
		code:    "INVALID_CHALLENGE_TOKEN",
		status:  http.StatusUnauthorized,
	},
}

var unhandledAuthError = errorInfo{
	message: "unhandled auth error",
	code:    "UNHANDLED_AUTH_ERROR",
	status:  http.StatusInternalServerError,
}

func (e authError) Error() string {
	return e.info().message
}

func (e authError) Code() string {
	return e.info().code
}

func (e authError) Status() int {
	return e.info().status
}

func (e authError) info() errorInfo {
	if int(e) < 0 || int(e) >= len(authErrorTable) {
		return unhandledAuthError
	}

	info := authErrorTable[e]

	if info.message == "" {
		return unhandledAuthError
	}

	return info
}
