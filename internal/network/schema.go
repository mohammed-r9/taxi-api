package network

import (
	"github.com/labstack/echo/v5"
)

type ResponseOutput[T any] struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Data    T      `json:"data"`
}

type ErrorResponse struct {
	Success    bool   `json:"success"`
	Status     int    `json:"status"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message"`
	Detail     any    `json:"detail,omitempty"`
	httpStatus int
	cause      error
}

func (e *ErrorResponse) Error() string { return e.Message }

func (e *ErrorResponse) GetStatus() int { return e.httpStatus }

func (e *ErrorResponse) Unwrap() error { return e.cause }

func NewSuccessResponse[T any](status int, data T, msg string) *ResponseOutput[T] {
	return &ResponseOutput[T]{
		Status:  status,
		Message: msg,
		Data:    data,
		Success: true,
	}
}

func WriteError(ctx *echo.Context, status int, code, msg string, err ...error) error {
	var detail any
	if len(err) > 0 && err[0] != nil {
		detail = err[0].Error()
	}
	return ctx.JSON(status, ErrorResponse{
		Message: msg,
		Status:  status,
		Code:    code,
		Success: false,
		Detail:  detail,
	})
}

func WriteValidationError(ctx *echo.Context, status int, code, msg string, messages []string) error {
	return ctx.JSON(status, ErrorResponse{
		Message: msg,
		Status:  status,
		Code:    code,
		Success: false,
		Detail:  messages,
	})
}
