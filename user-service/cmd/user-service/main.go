package main

import (
	"os"
	"time"

	"user-service/internal/conf"
	"user-service/internal/server"
	"user-service/internal/biz"
	"user-service/internal/data"
	"user-service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func newApp(logger log.Logger, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.Name("user-service"),
		kratos.Version("1.0.0"),
		kratos.Logger(logger),
		kratos.Server(gs),
	)
}

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.name", "user-service",
	)

	cfg := &conf.Config{
		Server: conf.Server{
			Grpc: conf.Grpc{
				Addr: ":9004",
			},
		},
		Data: conf.Data{
			Database: conf.Database{
				Source: "root:root@tcp(127.0.0.1:3306)/mall_user?charset=utf8mb4&parseTime=True&loc=Local",
			},
		},
		Auth: conf.Auth{
			JwtSecret:        "your-jwt-secret-change-in-production",
			TokenExpireHours: 24,
		},
	}

	d, cleanup, err := data.NewData(&cfg.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	uRepo := data.NewUserRepo(d, logger)
	aRepo := data.NewAddressRepo(d, logger)
	tokenGen := biz.NewTokenGenerator(cfg.Auth.JwtSecret, time.Duration(cfg.Auth.TokenExpireHours)*time.Hour)
	uc := biz.NewUserUseCase(uRepo, aRepo, tokenGen)
	svc := service.NewUserService(uc)
	gs := server.NewGRPCServer(cfg, svc, logger)

	app := newApp(logger, gs)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
