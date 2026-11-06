package health

import (
	"context"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

type CheckOutput string

func (h *Handler) Check(context.Context, *struct{}) (*CheckOutput, error) {
	output := CheckOutput("status is available")
	return &output, nil
}
