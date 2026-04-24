package main

import (
	"context"
	"flag"
	"os"

	"gateway/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lgzzzz/mall-tracing/tracing"
	"go.opentelemetry.io/otel/trace"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	Name     = "gateway"
	Version  string
	confPath string
)

func init() {
	flag.StringVar(&confPath, "conf", "configs/config.yaml", "config path, eg: -conf configs/config.yaml")
	flag.Parse()
}

func newApp(hs *http.Server, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Logger(logger),
		kratos.Server(hs),
	)
}

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.name", Name,
	)
	h := log.NewHelper(logger)

	c := config.New(
		config.WithSource(
			file.NewSource(confPath),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		h.Fatalf("failed to load config: %v", err)
	}

	var bootstrap conf.Bootstrap
	if err := c.Scan(&bootstrap); err != nil {
		h.Fatalf("failed to scan config: %v", err)
	}

	cfg := extractConfig(&bootstrap)

	var tp trace.TracerProvider
	tracer := tracing.NewTracer("gateway")
	if cfg.Tracing != nil && cfg.Tracing.Enabled {
		var err error
		tp, err = tracing.Init(tracing.Config{
			ServiceName:  "gateway",
			Version:     Version,
			OTLPEndpoint: cfg.Tracing.Endpoint,
			SampleRatio:  cfg.Tracing.SampleRatio,
			Insecure:    true,
		})
		if err != nil {
			h.Errorf("failed to init tracer provider: %v", err)
		} else {
			h.Info("Tracer provider initialized")
		}
	}

	app, cleanup, err := wireApp(cfg, logger, tracer)
	if err != nil {
		h.Fatalf("failed to wire app: %v", err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		h.Fatalf("failed to run app: %v", err)
	}

	if tp != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5e9)
		defer cancel()
		if err := tracing.Shutdown(shutdownCtx, tp); err != nil {
			h.Errorf("failed to shutdown tracer provider: %v", err)
		}
	}
}

func extractConfig(b *conf.Bootstrap) *conf.Config {
	return &conf.Config{
		Server:         b.Server,
		Data:           b.Data,
		Registry:       b.Registry,
		Auth:           b.Auth,
		RateLimit:      b.RateLimit,
		CircuitBreaker: b.CircuitBreaker,
		Tracing:        b.Tracing,
	}
}
