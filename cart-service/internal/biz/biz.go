package biz

import (
	"context"

	"github.com/google/wire"
)

// Product represents a product from product-service
type Product struct {
	ID     int64
	Name   string
	Price  int64
	Status int32
}

// ProductRepo is the interface for product service client
type ProductRepo interface {
	GetProduct(ctx context.Context, id int64) (*Product, error)
}

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewCartUseCase)
