package client

import (
	"context"

	cartV1 "cart-service/api/cart/v1"
	invV1 "inventory-service/api/inventory/v1"
	orderV1 "order-service/api/proto/order/v1"
	paymentV1 "payment-service/api/payment/v1"
	productV1 "product-service/api/product/v1"
	promoV1 "promotion-service/api/promotion/v1"
	userV1 "user-service/api/user/v1"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/registry"
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

type Clients struct {
	User      userV1.UserServiceClient
	Product   productV1.ProductServiceClient
	Order     orderV1.OrderServiceClient
	Cart      cartV1.CartServiceClient
	Payment   paymentV1.PaymentServiceClient
	Inventory invV1.InventoryClient
	Promotion promoV1.PromotionServiceClient
}

func NewClients(etcdEndpoints []string) (*Clients, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: etcdEndpoints,
	})
	if err != nil {
		return nil, err
	}

	discovery := etcd.New(etcdClient)

	return &Clients{
		User:      userV1.NewUserServiceClient(newGRPCClient(discovery, "discovery:///user-service")),
		Product:   productV1.NewProductServiceClient(newGRPCClient(discovery, "discovery:///product-service")),
		Order:     orderV1.NewOrderServiceClient(newGRPCClient(discovery, "discovery:///order-service")),
		Cart:      cartV1.NewCartServiceClient(newGRPCClient(discovery, "discovery:///cart-service")),
		Payment:   paymentV1.NewPaymentServiceClient(newGRPCClient(discovery, "discovery:///payment-service")),
		Inventory: invV1.NewInventoryClient(newGRPCClient(discovery, "discovery:///inventory-service")),
		Promotion: promoV1.NewPromotionServiceClient(newGRPCClient(discovery, "discovery:///promotion-service")),
	}, nil
}

func newGRPCClient(r registry.Discovery, endpoint string) *grpc.ClientConn {
	conn, err := kratosgrpc.DialInsecure(
		context.Background(),
		kratosgrpc.WithEndpoint(endpoint),
		kratosgrpc.WithDiscovery(r),
		kratosgrpc.WithMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	return conn
}

var ProviderSet = wire.NewSet(NewClients)
