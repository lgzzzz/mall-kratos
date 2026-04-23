//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"gateway/internal/client"
	"gateway/internal/conf"
	"gateway/internal/server"
	"gateway/internal/service"
)

func wireApp(*conf.Config, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		client.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
		newApp,
	))
}
