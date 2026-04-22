package data

import (
	"context"

	"cart-service/internal/biz"

	productV1 "product-service/api/product/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

type productRepo struct {
	client productV1.ProductServiceClient
}

func NewProductRepo(r registry.Discovery, endpoint string, logger log.Logger) biz.ProductRepo {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		panic(err)
	}
	return &productRepo{client: productV1.NewProductServiceClient(conn)}
}

func (r *productRepo) GetProduct(ctx context.Context, id int64) (*biz.Product, error) {
	reply, err := r.client.GetProduct(ctx, &productV1.GetProductRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &biz.Product{
		ID:     reply.Id,
		Name:   reply.Name,
		Price:  reply.Price,
		Status: reply.Status,
	}, nil
}
