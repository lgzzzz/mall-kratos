package server

import (
	v1 "payment-service/api/payment/v1"
	"payment-service/internal/conf"
	"payment-service/internal/middleware"
	"payment-service/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(c *conf.Config, payment *service.PaymentService, logger log.Logger) *grpc.Server {
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
	v1.RegisterPaymentServiceServer(srv, payment)
	return srv
}
