package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	reasonCartNotFound     = "CART_ITEM_NOT_FOUND"
	reasonProductNotFound  = "PRODUCT_NOT_FOUND"
	reasonQuantityExceeded = "QUANTITY_EXCEEDS_LIMIT"
	reasonUnauthorized     = "UNAUTHORIZED"
)

var (
	ErrCartNotFound     = errors.NotFound(reasonCartNotFound, "cart item not found")
	ErrProductNotFound  = errors.NotFound(reasonProductNotFound, "product not found")
	ErrQuantityExceeded = errors.BadRequest(reasonQuantityExceeded, "quantity exceeds limit")
	ErrUnauthorized     = errors.Unauthorized(reasonUnauthorized, "unauthorized")
)
