package apperrors

import "net/http"

const (
	MissingEnvVar envError = iota

	// new errors should be added above this line
	envErrorCount
)

var envErrorTable = [envErrorCount]errorInfo{
	MissingEnvVar: {
		message: "missing environment variable",
		code:    "MISSING_ENV_VAR",
		status:  http.StatusInternalServerError,
	},
}

var unhandledEnvError = errorInfo{
	message: "unhandled env error",
	code:    "UNHANDLED_ENV_ERROR",
	status:  http.StatusInternalServerError,
}

func (e envError) Error() string {
	return e.info().message
}

func (e envError) Code() string {
	return e.info().code
}

func (e envError) Status() int {
	return e.info().status
}

func (e envError) info() errorInfo {
	if int(e) < 0 || int(e) >= len(envErrorTable) {
		return unhandledEnvError
	}

	return envErrorTable[e]
}
