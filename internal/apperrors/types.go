package apperrors

type (
	authError     uint8
	networkError  uint8
	databaseError uint8
	envError      uint8
)

type AppError interface {
	error
	Code() string
	Status() int
}

type errorInfo struct {
	message string
	code    string
	status  int
	detail  string
}
