package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	reasonCouponNotFound   = "COUPON_NOT_FOUND"
	reasonCouponExhausted  = "COUPON_EXHAUSTED"
	reasonCouponLimit      = "COUPON_PER_USER_LIMIT"
	reasonCouponExpired    = "COUPON_EXPIRED"
	reasonUnauthorized     = "UNAUTHORIZED"
)

var (
	ErrCouponNotFound  = errors.NotFound(reasonCouponNotFound, "coupon not found")
	ErrCouponExhausted = errors.BadRequest(reasonCouponExhausted, "coupon exhausted")
	ErrCouponLimit     = errors.BadRequest(reasonCouponLimit, "user coupon limit exceeded")
	ErrCouponExpired   = errors.BadRequest(reasonCouponExpired, "coupon expired")
	ErrUnauthorized    = errors.Unauthorized(reasonUnauthorized, "unauthorized")
)
