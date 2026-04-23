package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"

	"gateway/internal/conf"
)

type AuthInfo struct {
	UserID   int64
	Username string
	Role     string
}

type authKey struct{}

func ServerAuth(secret string, whitelist []string) middleware.Middleware {
	whitelistMap := make(map[string]bool, len(whitelist))
	for _, w := range whitelist {
		whitelistMap[w] = true
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Extract HTTP method and path from context
			method, path := extractHTTPInfo(ctx)
			key := method + ":" + path

			// Check whitelist
			if whitelistMap[key] {
				return handler(ctx, req)
			}

			// Extract token
			tokenStr := extractToken(ctx)
			if tokenStr == "" {
				return nil, conf.ErrUnauthorized
			}

			// Parse and validate JWT
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return nil, conf.ErrUnauthorized
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, conf.ErrUnauthorized
			}

			info := &AuthInfo{}
			if uid, ok := claims["user_id"].(float64); ok {
				info.UserID = int64(uid)
			}
			if username, ok := claims["username"].(string); ok {
				info.Username = username
			}
			if role, ok := claims["role"].(string); ok {
				info.Role = role
			}

			if info.UserID == 0 {
				return nil, conf.ErrUnauthorized
			}

			// Store auth info in context
			ctx = context.WithValue(ctx, authKey{}, info)

			// Inject into gRPC metadata for downstream services
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			md.Set("x-user-id", fmt.Sprintf("%d", info.UserID))
			md.Set("x-username", info.Username)
			md.Set("x-role", info.Role)
			ctx = metadata.NewOutgoingContext(ctx, md)

			return handler(ctx, req)
		}
	}
}

func GetAuthInfo(ctx context.Context) (*AuthInfo, bool) {
	info, ok := ctx.Value(authKey{}).(*AuthInfo)
	return info, ok
}

func extractToken(ctx context.Context) string {
	// Try HTTP header first
	if req, ok := http.RequestFromServerContext(ctx); ok {
		auth := req.Header.Get("Authorization")
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

func extractHTTPInfo(ctx context.Context) (method, path string) {
	// Kratos stores HTTP transport in context
	if req, ok := http.RequestFromServerContext(ctx); ok {
		method = req.Method
		path = req.URL.Path
	}
	return
}
