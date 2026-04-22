//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"promotion-service/internal/biz"
	"promotion-service/internal/conf"
	"promotion-service/internal/data"
	"promotion-service/internal/server"
	"promotion-service/internal/service"
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
