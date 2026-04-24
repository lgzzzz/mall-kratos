package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lgzzzz/mall-tracing/middleware"

	"gateway/internal/conf"
	gwMiddleware "gateway/internal/middleware"
	"gateway/internal/service"
	"go.opentelemetry.io/otel/trace"
)

func NewHTTPServer(
	cfg *conf.Config,
	gw *service.GatewayService,
	logger log.Logger,
	tracer trace.Tracer,
) *http.Server {
	var opts = []http.ServerOption{
		http.Address(cfg.Server.Http.Addr),
		http.Timeout(cfg.Server.Http.Timeout.AsDuration()),
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			ratelimit.Server(),
			middleware.ServerMiddleware(tracer),
			gwMiddleware.ResponseError(),
			gwMiddleware.ServerAuth(cfg.Auth.JwtSecret, cfg.Auth.Whitelist),
		),
	}

	srv := http.NewServer(opts...)

	// Register health check route
	srv.Route("").GET("/health", func(ctx http.Context) error {
		return ctx.Result(200, map[string]string{"status": "ok"})
	})

	// Register gateway routes
	gw.RegisterRoutes(srv)

	return srv
}
