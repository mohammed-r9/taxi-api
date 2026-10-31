package health

import (
	"context"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

type CheckOutput struct {
	Body string `contentType:"text/plain"`
}

func (h *Handler) Check(context.Context, *struct{}) (*CheckOutput, error) {
	return &CheckOutput{Body: "status is available"}, nil
}
