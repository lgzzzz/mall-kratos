package data

import (
	"cart-service/internal/conf"

	etcdregistry "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/go-kratos/kratos/v2/registry"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewCartRepo, NewProductRepo, NewDiscovery, ProvideDataConfig, ProvideProductEndpoint)

// ProvideDataConfig extracts *conf.Data from *conf.Config for Wire injection.
func ProvideDataConfig(cfg *conf.Config) *conf.Data {
	return &cfg.Data
}

// ProvideProductEndpoint extracts the product service endpoint string for Wire injection.
func ProvideProductEndpoint(cfg *conf.Config) string {
	return cfg.GrpcClients.ProductService
}

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	return &Data{db: db}, cleanup, nil
}

func NewDiscovery(cfg *conf.Config) registry.Discovery {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: cfg.Registry.Endpoints,
	})
	if err != nil {
		panic(err)
	}
	return etcdregistry.New(etcdClient)
}
