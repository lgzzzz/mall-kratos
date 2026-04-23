package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	ErrCodeSuccess          = 0
	ErrCodeInternal         = 50000
	ErrCodeUnauthorized     = 50005
	ErrCodeForbidden        = 50006
	ErrCodeNotFound         = 50007
	ErrCodeBadRequest       = 50008
	ErrCodeRateLimitExceeded = 50009
	ErrCodeServiceUnavailable = 50010
)

var (
	ErrUnauthorized     = errors.Unauthorized("UNAUTHORIZED", "unauthorized")
	ErrForbidden        = errors.Forbidden("FORBIDDEN", "forbidden")
	ErrNotFound         = errors.NotFound("NOT_FOUND", "not found")
	ErrBadRequest       = errors.BadRequest("BAD_REQUEST", "bad request")
	ErrRateLimitExceeded = errors.New(429, "RATE_LIMIT_EXCEEDED", "rate limit exceeded")
	ErrServiceUnavailable = errors.New(503, "SERVICE_UNAVAILABLE", "service unavailable")
	ErrInternal         = errors.InternalServer("INTERNAL", "internal server error")
)

var reasonToCode = map[string]int32{
	"UNAUTHORIZED":        ErrCodeUnauthorized,
	"FORBIDDEN":           ErrCodeForbidden,
	"NOT_FOUND":           ErrCodeNotFound,
	"BAD_REQUEST":         ErrCodeBadRequest,
	"RATE_LIMIT_EXCEEDED": ErrCodeRateLimitExceeded,
	"SERVICE_UNAVAILABLE": ErrCodeServiceUnavailable,
	"INTERNAL":            ErrCodeInternal,
}

func ErrorCodeFromReason(reason string) int32 {
	if code, ok := reasonToCode[reason]; ok {
		return code
	}
	return ErrCodeInternal
}
