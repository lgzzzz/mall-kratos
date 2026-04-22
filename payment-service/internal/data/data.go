package data

import (
	"payment-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewPaymentRepo, ProvideDataConfig)

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

func ProvideDataConfig(c *conf.Config) *conf.Data {
	return &c.Data
}
