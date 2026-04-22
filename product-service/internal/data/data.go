package data

import (
	"product-service/internal/conf"

	etcdregistry "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewProductRepo, NewCategoryRepo, NewInventoryRepo, NewDiscovery, ProvideDataConfig, ProvideInventoryEndpoint)

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

func NewDiscovery(c *conf.Config) registry.Discovery {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: c.Registry.Endpoints,
	})
	if err != nil {
		panic(err)
	}
	return etcdregistry.New(etcdClient)
}

func ProvideDataConfig(c *conf.Config) *conf.Data {
	return &c.Data
}

func ProvideInventoryEndpoint(c *conf.Config) string {
	return c.GrpcClients.InventoryService
}
