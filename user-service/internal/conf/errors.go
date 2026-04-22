package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	reasonUserExists      = "USER_ALREADY_EXISTS"
	reasonInvalidPassword = "INVALID_PASSWORD"
	reasonLoginLocked     = "LOGIN_LOCKED"
	reasonAddressNotFound = "ADDRESS_NOT_FOUND"
)

var (
	ErrUserExists      = errors.BadRequest(reasonUserExists, "user already exists")
	ErrInvalidPassword = errors.Unauthorized(reasonInvalidPassword, "invalid username or password")
	ErrLoginLocked     = errors.New(429, reasonLoginLocked, "account locked, try again later")
	ErrAddressNotFound = errors.NotFound(reasonAddressNotFound, "address not found")
)
