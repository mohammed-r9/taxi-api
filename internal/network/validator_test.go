package network

import (
	"errors"
	"testing"
)

type exampleRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=3"`
}

func TestRequestValidator(t *testing.T) {
	v := NewRequestValidator()

	if err := v.Validate(exampleRequest{Email: "a@b.com", Name: "abc"}); err != nil {
		t.Fatalf("valid request returned error: %v", err)
	}

	err := v.Validate(exampleRequest{Email: "not-an-email", Name: "ab"})
	var validationErr *RequestValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want RequestValidationError", err)
	}
	if len(validationErr.Messages) != 2 {
		t.Fatalf("messages = %v, want 2 entries", validationErr.Messages)
	}
}
