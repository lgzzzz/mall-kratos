//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"cart-service/internal/biz"
	"cart-service/internal/conf"
	"cart-service/internal/data"
	"cart-service/internal/server"
	"cart-service/internal/service"
)

func wireApp(*conf.Config, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
		newApp,
	))
}
