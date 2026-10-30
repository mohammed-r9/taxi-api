package network

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type RequestValidationError struct {
	Messages []string
}

func (e *RequestValidationError) Error() string {
	return strings.Join(e.Messages, "; ")
}

type RequestValidator struct {
	validator *validator.Validate
}

func NewRequestValidator() *RequestValidator {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")

		switch name {
		case "-":
			return ""
		case "":
			return field.Name
		default:
			return name
		}
	})

	return &RequestValidator{
		validator: v,
	}
}

func (v *RequestValidator) Validate(value any) error {
	err := v.validator.Struct(value)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return err
	}

	messages := make([]string, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		messages = append(messages, validationMessage(fieldErr))
	}

	return &RequestValidationError{
		Messages: messages,
	}
}

func validationMessage(err validator.FieldError) string {
	field := err.Field()
	param := err.Param()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)

	case "min":
		if err.Kind() == reflect.String {
			return fmt.Sprintf(
				"%s must contain at least %s characters",
				field,
				param,
			)
		}

		return fmt.Sprintf("%s must be at least %s", field, param)

	case "max":
		if err.Kind() == reflect.String {
			return fmt.Sprintf(
				"%s must not exceed %s characters",
				field,
				param,
			)
		}

		return fmt.Sprintf("%s must not be greater than %s", field, param)

	case "len":
		return fmt.Sprintf(
			"%s must contain exactly %s characters",
			field,
			param,
		)

	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)

	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)

	case "gte":
		return fmt.Sprintf(
			"%s must be greater than or equal to %s",
			field,
			param,
		)

	case "lte":
		return fmt.Sprintf(
			"%s must be less than or equal to %s",
			field,
			param,
		)

	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)

	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)

	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)

	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)

	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
