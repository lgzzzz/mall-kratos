package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	kratosmiddleware "github.com/go-kratos/kratos/v2/middleware"
)

func ResponseError() kratosmiddleware.Middleware {
	return func(handler kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			reply, err := handler(ctx, req)
			if err != nil {
				e := errors.FromError(err)
				httpCode := grpcToHTTPCode(e.Code)
				return nil, errors.New(httpCode, e.Reason, e.Message)
			}
			return reply, err
		}
	}
}

func grpcToHTTPCode(grpcCode int32) int {
	switch grpcCode {
	case 0:
		return 200
	case 3:
		return 400
	case 16:
		return 401
	case 7:
		return 403
	case 5:
		return 404
	case 6:
		return 409
	default:
		return 500
	}
}
