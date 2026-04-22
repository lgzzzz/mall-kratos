package biz

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenGenerator struct {
	secret string
	expiry time.Duration
}

func NewTokenGenerator(secret string, expiry time.Duration) *TokenGenerator {
	return &TokenGenerator{secret: secret, expiry: expiry}
}

func (g *TokenGenerator) GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     "user",
		"exp":      time.Now().Add(g.expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secret))
}
