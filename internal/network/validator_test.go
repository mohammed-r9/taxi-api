package network

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type exampleRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=3"`
}

func TestRequestValidator(t *testing.T) {
	v := NewRequestValidator()

	require.NoError(t, v.Validate(exampleRequest{Email: "a@b.com", Name: "abc"}))

	err := v.Validate(exampleRequest{Email: "not-an-email", Name: "ab"})
	var validationErr *RequestValidationError
	require.ErrorAs(t, err, &validationErr)
	require.Len(t, validationErr.Messages, 2)
}
