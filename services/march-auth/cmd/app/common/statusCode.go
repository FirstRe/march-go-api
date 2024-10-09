package common

import (
	"go-graphql/cmd/app/graph/types"
	"net/http"
)

func StatusResponse(code int, message string) *types.Status {
	if message == "" {
		message = http.StatusText(code)
	}
	status := types.Status{
		Code:    &code,
		Message: &message,
	}
	return &status
}
