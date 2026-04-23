package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	kratosmiddleware "github.com/go-kratos/kratos/v2/middleware"
)

// ResponseError converts gRPC errors from downstream services to HTTP error responses.
// The mapped HTTP status code is handled by Kratos' built-in HTTP error encoder,
// which automatically writes the correct status code and JSON error body.
func ResponseError() kratosmiddleware.Middleware {
	return func(handler kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			reply, err := handler(ctx, req)
			if err != nil {
				e := errors.FromError(err)
				httpCode := GrpcToHTTPCode(e.Code)
				return nil, errors.New(httpCode, e.Reason, e.Message)
			}
			return reply, err
		}
	}
}

// GrpcToHTTPCode maps gRPC status codes to HTTP status codes.
// Note: grpcCode is always non-zero since this is only called when err != nil.
func GrpcToHTTPCode(grpcCode int32) int {
	switch grpcCode {
	case 3: // InvalidArgument
		return 400
	case 16: // Unauthenticated
		return 401
	case 7: // PermissionDenied
		return 403
	case 5: // NotFound
		return 404
	case 6: // AlreadyExists
		return 409
	default:
		return 500
	}
}
