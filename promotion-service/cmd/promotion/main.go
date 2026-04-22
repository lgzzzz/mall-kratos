package main

import (
	"flag"
	"os"

	"promotion-service/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func init() {
	flag.Parse()
}

func newApp(logger log.Logger, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.Name("promotion-service"),
		kratos.Version("1.0.0"),
		kratos.Logger(logger),
		kratos.Server(gs),
	)
}

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.name", "promotion-service",
	)

	c := config.New(
		config.WithSource(
			file.NewSource("configs/config.yaml"),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var cfg conf.Config
	if err := c.Scan(&cfg); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(&cfg, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
