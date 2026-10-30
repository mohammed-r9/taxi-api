package network

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
)

// BaseResponse is the common envelope returned by successful API operations.
type BaseResponse[T any] struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Data    T      `json:"data"`
}

// ErrorResponse is the common envelope returned by failed API operations.
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

func configureHumaErrors() {
	huma.NewError = newHumaError
	huma.NewErrorWithContext = func(_ huma.Context, status int, msg string, errs ...error) huma.StatusError {
		return newHumaError(status, msg, errs...)
	}
}

func newHumaError(status int, msg string, errs ...error) huma.StatusError {
	for _, err := range errs {
		if response, ok := err.(*ErrorResponse); ok {
			return response
		}
	}

	code := "INVALID_REQUEST_BODY"
	if status >= http.StatusInternalServerError {
		code = "internal_error"
	}

	details := make([]string, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			details = append(details, err.Error())
		}
	}

	var detail any
	if len(details) > 0 {
		detail = details
	}

	return &ErrorResponse{
		Success:    false,
		Status:     status,
		Code:       code,
		Message:    msg,
		Detail:     detail,
		httpStatus: status,
	}
}

func toHumaError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(huma.StatusError); ok {
		return err
	}

	status, info := mapError(err)
	var detail any
	if len(info.detail) > 0 {
		detail = info.detail
	}
	return &ErrorResponse{
		Success:    false,
		Status:     status,
		Code:       info.code,
		Message:    info.message,
		Detail:     detail,
		httpStatus: status,
		cause:      err,
	}
}

func WriteSuccess[T any](ctx *echo.Context, status int, data T, msg string) error {
	return ctx.JSON(status, BaseResponse[T]{
		Status:  status,
		Message: msg,
		Data:    data,
		Success: true,
	})
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
