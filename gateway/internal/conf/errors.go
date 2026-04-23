package conf

import "github.com/go-kratos/kratos/v2/errors"

// Gateway error codes use 10000 range to avoid collision with service-specific codes.
const (
	ErrCodeSuccess            = 0
	ErrCodeInternal           = 10000
	ErrCodeUnauthorized       = 10001
	ErrCodeForbidden          = 10002
	ErrCodeNotFound           = 10003
	ErrCodeBadRequest         = 10004
	ErrCodeRateLimitExceeded  = 10005
	ErrCodeServiceUnavailable = 10006
)

const (
	reasonUnauthorized       = "UNAUTHORIZED"
	reasonForbidden          = "FORBIDDEN"
	reasonNotFound           = "NOT_FOUND"
	reasonBadRequest         = "BAD_REQUEST"
	reasonRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	reasonServiceUnavailable = "SERVICE_UNAVAILABLE"
	reasonInternal           = "INTERNAL"
)

var (
	ErrUnauthorized       = errors.Unauthorized(reasonUnauthorized, "unauthorized")
	ErrForbidden          = errors.Forbidden(reasonForbidden, "forbidden")
	ErrNotFound           = errors.NotFound(reasonNotFound, "not found")
	ErrBadRequest         = errors.BadRequest(reasonBadRequest, "bad request")
	ErrRateLimitExceeded  = errors.New(429, reasonRateLimitExceeded, "rate limit exceeded")
	ErrServiceUnavailable = errors.New(503, reasonServiceUnavailable, "service unavailable")
	ErrInternal           = errors.InternalServer(reasonInternal, "internal server error")
)

var reasonToCode = map[string]int32{
	reasonUnauthorized:       ErrCodeUnauthorized,
	reasonForbidden:          ErrCodeForbidden,
	reasonNotFound:           ErrCodeNotFound,
	reasonBadRequest:         ErrCodeBadRequest,
	reasonRateLimitExceeded:  ErrCodeRateLimitExceeded,
	reasonServiceUnavailable: ErrCodeServiceUnavailable,
	reasonInternal:           ErrCodeInternal,
}

func ErrorCodeFromReason(reason string) int32 {
	if code, ok := reasonToCode[reason]; ok {
		return code
	}
	return ErrCodeInternal
}
