package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	reasonPaymentNotFound = "PAYMENT_NOT_FOUND"
	reasonAlreadyPaid    = "PAYMENT_ALREADY_PAID"
	reasonInvalidStatus  = "INVALID_PAYMENT_STATUS"
	reasonUnauthorized   = "UNAUTHORIZED"
)

var (
	ErrPaymentNotFound = errors.NotFound(reasonPaymentNotFound, "payment not found")
	ErrAlreadyPaid     = errors.BadRequest(reasonAlreadyPaid, "payment already paid")
	ErrInvalidStatus   = errors.BadRequest(reasonInvalidStatus, "invalid payment status")
	ErrUnauthorized    = errors.Unauthorized(reasonUnauthorized, "unauthorized")
)
