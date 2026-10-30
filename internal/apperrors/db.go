package apperrors

import "net/http"

const (
	UniqueViolation databaseError = iota
	NoRowsAffected
	CacheMiss

	// new errors should be added above this line
	dbErrorCount
)

var databaseErrorTable = [dbErrorCount]errorInfo{
	UniqueViolation: {
		message: "unique violation",
		code:    "UNIQUE_VIOLATION",
		status:  http.StatusConflict,
	},
	NoRowsAffected: {
		message: "no rows affected",
		code:    "NO_ROWS_AFFECTED",
		status:  http.StatusNotFound,
	},
	CacheMiss: {
		message: "cache miss",
		code:    "CACHE_MISS",
		status:  http.StatusNotFound,
	},
}

var unhandledDatabaseError = errorInfo{
	message: "unhandled database error",
	code:    "UNHANDLED_DATABASE_ERROR",
	status:  http.StatusInternalServerError,
}

func (e databaseError) Error() string {
	return e.info().message
}

func (e databaseError) Code() string {
	return e.info().code
}

func (e databaseError) Status() int {
	return e.info().status
}

func (e databaseError) info() errorInfo {
	if int(e) < 0 || int(e) >= len(databaseErrorTable) {
		return unhandledDatabaseError
	}

	return databaseErrorTable[e]
}
