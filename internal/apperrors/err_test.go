package apperrors

import (
	"net/http"
	"testing"
)

func TestAuthErrTable(t *testing.T) {
	for e := range int(authErrorCount) {
		info := authErrorTable[e]

		if info.code == "" ||
			info.message == "" ||
			info.status == 0 {
			t.Fatalf(
				"invalid entry at %v (%v): %+v",
				e,
				authError(e),
				info,
			)
		}
	}
}

func TestNetworkErrTable(t *testing.T) {
	for e := range int(networkErrorCount) {
		info := networkErrorTable[e]

		if info.code == "" ||
			info.message == "" ||
			info.status == 0 {
			t.Fatalf(
				"invalid entry at %v (%v): %+v",
				e,
				networkError(e),
				info,
			)
		}
	}
}

func TestEnvErrTable(t *testing.T) {
	for e := range int(envErrorCount) {
		info := envErrorTable[e]

		if info.code == "" ||
			info.message == "" ||
			info.status == 0 {
			t.Fatalf(
				"invalid entry at %v (%v): %+v",
				e,
				envError(e),
				info,
			)
		}
	}
}

func TestDatabaseErrTable(t *testing.T) {
	for e := range int(dbErrorCount) {
		info := databaseErrorTable[e]

		if info.code == "" ||
			info.message == "" ||
			info.status == 0 {
			t.Fatalf(
				"invalid entry at %v (%v): %+v",
				e,
				databaseError(e),
				info,
			)
		}
	}
}

func TestOutOfRangeErrorsFallBackTo500(t *testing.T) {
	errs := []AppError{
		authError(authErrorCount),
		networkError(networkErrorCount),
		databaseError(dbErrorCount),
		envError(envErrorCount),
	}

	for _, err := range errs {
		if err.Status() != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %v", err.Status())
		}
	}
}
