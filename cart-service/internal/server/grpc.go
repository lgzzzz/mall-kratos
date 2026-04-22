package server

import (
	v1 "cart-service/api/cart/v1"
	"cart-service/internal/conf"
	"cart-service/internal/middleware"
	"cart-service/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(c *conf.Config, cart *service.CartService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			middleware.ServerAuth(c.Auth.JwtSecret),
		),
	}
	if c.Server.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Server.Grpc.Addr))
	}
	if c.Server.Grpc.Timeout > 0 {
		opts = append(opts, grpc.Timeout(c.Server.Grpc.Timeout))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterCartServiceServer(srv, cart)
	return srv
}
