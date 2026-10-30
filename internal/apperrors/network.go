package apperrors

import "net/http"

const (
	MissingUserIDInContext networkError = iota
	InvalidRequestBody
	FileTooLarge
	MissingFile
	InvalidMime

	// new errors should be added above this line
	networkErrorCount
)

var networkErrorTable = [networkErrorCount]errorInfo{
	MissingUserIDInContext: {
		message: "missing user id in context",
		code:    "MISSING_USER_ID_IN_CONTEXT",
		status:  http.StatusUnauthorized,
	},
	InvalidRequestBody: {
		message: "invalid request body",
		code:    "INVALID_REQUEST_BODY",
		status:  http.StatusBadRequest,
	},
	FileTooLarge: {
		message: "file too large",
		code:    "FILE_TOO_LARGE",
		status:  http.StatusRequestEntityTooLarge,
	},
	MissingFile: {
		message: "missing file",
		code:    "MISSING_FILE",
		status:  http.StatusBadRequest,
	},
	InvalidMime: {
		message: "invalid file type",
		code:    "INVALID_FILE_TYPE",
		status:  http.StatusUnprocessableEntity,
	},
}

var unhandledNetworkError = errorInfo{
	message: "unhandled network error",
	code:    "UNHANDLED_NETWORK_ERROR",
	status:  http.StatusInternalServerError,
}

func (e networkError) Error() string {
	return e.info().message
}

func (e networkError) Code() string {
	return e.info().code
}

func (e networkError) Status() int {
	return e.info().status
}

func (e networkError) info() errorInfo {
	if int(e) < 0 || int(e) >= len(networkErrorTable) {
		return unhandledNetworkError
	}

	return networkErrorTable[e]
}
