package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	reasonProductNotFound  = "PRODUCT_NOT_FOUND"
	reasonCategoryNotFound = "CATEGORY_NOT_FOUND"
	reasonForbidden        = "FORBIDDEN"
	reasonUnauthorized     = "UNAUTHORIZED"
)

var (
	ErrProductNotFound  = errors.NotFound(reasonProductNotFound, "product not found")
	ErrCategoryNotFound = errors.NotFound(reasonCategoryNotFound, "category not found")
	ErrForbidden        = errors.Forbidden(reasonForbidden, "permission denied")
	ErrUnauthorized     = errors.Unauthorized(reasonUnauthorized, "unauthorized")
)
