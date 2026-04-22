package server

import (
	v1 "promotion-service/api/promotion/v1"
	"promotion-service/internal/conf"
	"promotion-service/internal/middleware"
	"promotion-service/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Config, promotion *service.PromotionService, logger log.Logger) *grpc.Server {
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
	v1.RegisterPromotionServiceServer(srv, promotion)
	return srv
}
