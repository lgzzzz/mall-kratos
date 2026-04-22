package data

import (
	"context"

	"product-service/internal/biz"

	inventoryV1 "inventory-service/api/inventory/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

type inventoryRepo struct {
	client inventoryV1.InventoryClient
}

func NewInventoryRepo(r registry.Discovery, endpoint string, logger log.Logger) biz.InventoryRepo {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		panic(err)
	}
	return &inventoryRepo{client: inventoryV1.NewInventoryClient(conn)}
}

func (r *inventoryRepo) CreateInventory(ctx context.Context, skuID int64) error {
	_, err := r.client.CreateInventory(ctx, &inventoryV1.CreateInventoryRequest{
		SkuId:       skuID,
		WarehouseId: 1,
		Stock:       0,
	})
	return err
}
