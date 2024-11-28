package statusCode

import "march-auth/cmd/app/graph/types"

// ApiResponse is a generic response structure for API responses.
const (
	SuccessCode       = 1000
	ForbiddenCode     = 9001
	DuplicatedCode    = 9002
	BadRequestCode    = 9003
	OnUseCode         = 9004
	NotFoundCode      = 9005
	InternalErrorCode = 9999
)

// StatusMessage holds the default messages
const (
	SuccessMessage       = "Success"
	ForbiddenMessage     = "Forbidden"
	DuplicatedMessage    = "Duplicated entry"
	BadRequestMessage    = "Bad Request"
	OnUseMessage         = "Item is already in use"
	InternalErrorMessage = "Internal Server Error"
)

// Success returns a *types.Status for a successful operation.
func Success(message string) *types.Status {
	if message == "" {
		message = SuccessMessage
	}
	return &types.Status{
		Code:    SuccessCode,
		Message: &message,
	}
}

// Forbidden returns a *types.Status for a forbidden operation.
func Forbidden(message string) *types.Status {
	if message == "" {
		message = ForbiddenMessage
	}
	return &types.Status{
		Code:    ForbiddenCode,
		Message: &message,
	}
}

// Duplicated returns a *types.Status for a duplicated entry.
func Duplicated(message string) *types.Status {
	return &types.Status{
		Code:    DuplicatedCode,
		Message: &message,
	}
}

// BadRequest returns a *types.Status for a bad request.
func BadRequest(message string) *types.Status {
	if message == "" {
		message = BadRequestMessage
	}
	return &types.Status{
		Code:    BadRequestCode,
		Message: &message,
	}
}

// OnUse returns a *types.Status for an in-use item.
func OnUse(message string) *types.Status {
	return &types.Status{
		Code:    OnUseCode,
		Message: &message,
	}
}

func NotFound(message string) *types.Status {
	return &types.Status{
		Code:    NotFoundCode,
		Message: &message,
	}
}

// InternalError returns a *types.Status for an internal server error.
func InternalError(message string) *types.Status {
	if message == "" {
		message = InternalErrorMessage
	}
	return &types.Status{
		Code:    InternalErrorCode,
		Message: &message,
	}
}
