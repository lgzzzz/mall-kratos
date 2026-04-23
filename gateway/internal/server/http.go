package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"

	"gateway/internal/conf"
	gwMiddleware "gateway/internal/middleware"
	"gateway/internal/service"
)

func NewHTTPServer(
	cfg *conf.Config,
	gw *service.GatewayService,
	logger log.Logger,
) *http.Server {
	var opts = []http.ServerOption{
		http.Address(cfg.Server.Http.Addr),
		http.Timeout(cfg.Server.Http.Timeout.AsDuration()),
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			gwMiddleware.ResponseError(),
			gwMiddleware.ServerAuth(cfg.Auth.JwtSecret, cfg.Auth.Whitelist),
		),
	}

	srv := http.NewServer(opts...)

	// Register health check route
	srv.Route("/").GET("/health", func(ctx http.Context) error {
		return ctx.Result(200, map[string]string{"status": "ok"})
	})

	// Register gateway routes
	gw.RegisterRoutes(srv)

	return srv
}
