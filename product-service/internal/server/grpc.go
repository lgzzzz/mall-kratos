package server

import (
	v1 "product-service/api/product/v1"
	"product-service/internal/conf"
	"product-service/internal/middleware"
	"product-service/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Config, product *service.ProductService, logger log.Logger) *grpc.Server {
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
	v1.RegisterProductServiceServer(srv, product)
	return srv
}
