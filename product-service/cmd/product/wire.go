//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"product-service/internal/biz"
	"product-service/internal/conf"
	"product-service/internal/data"
	"product-service/internal/server"
	"product-service/internal/service"
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
