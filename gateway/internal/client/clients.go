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

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/middleware/circuitbreaker"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/registry"
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"gateway/internal/conf"
)

type Clients struct {
	User      userV1.UserServiceClient
	Product   productV1.ProductServiceClient
	Order     orderV1.OrderServiceClient
	Cart      cartV1.CartServiceClient
	Payment   paymentV1.PaymentServiceClient
	Inventory invV1.InventoryClient
	Promotion promoV1.PromotionServiceClient

	conns []*grpc.ClientConn
}

func NewClients(cfg *conf.Config) (*Clients, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: cfg.Data.EtcdEndpoints,
	})
	if err != nil {
		return nil, err
	}

	discovery := etcd.New(etcdClient)

	c := &Clients{}

	c.User, err = newUserClient(discovery, "discovery:///user-service", &c.conns)
	if err != nil {
		return nil, err
	}
	c.Product, err = newProductClient(discovery, "discovery:///product-service", &c.conns)
	if err != nil {
		return nil, err
	}
	c.Order, err = newOrderClient(discovery, "discovery:///order-service", &c.conns)
	if err != nil {
		return nil, err
	}
	c.Cart, err = newCartClient(discovery, "discovery:///cart-service", &c.conns)
	if err != nil {
		return nil, err
	}
	c.Payment, err = newPaymentClient(discovery, "discovery:///payment-service", &c.conns)
	if err != nil {
		return nil, err
	}
	c.Inventory, err = newInventoryClient(discovery, "discovery:///inventory-service", &c.conns)
	if err != nil {
		return nil, err
	}
	c.Promotion, err = newPromotionClient(discovery, "discovery:///promotion-service", &c.conns)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Clients) Close() error {
	for _, conn := range c.conns {
		conn.Close()
	}
	return nil
}

func newUserClient(r registry.Discovery, endpoint string, conns *[]*grpc.ClientConn) (userV1.UserServiceClient, error) {
	conn, err := newGRPCClient(r, endpoint, conns)
	if err != nil {
		return nil, err
	}
	return userV1.NewUserServiceClient(conn), nil
}

func newProductClient(r registry.Discovery, endpoint string, conns *[]*grpc.ClientConn) (productV1.ProductServiceClient, error) {
	conn, err := newGRPCClient(r, endpoint, conns)
	if err != nil {
		return nil, err
	}
	return productV1.NewProductServiceClient(conn), nil
}

func newOrderClient(r registry.Discovery, endpoint string, conns *[]*grpc.ClientConn) (orderV1.OrderServiceClient, error) {
	conn, err := newGRPCClient(r, endpoint, conns)
	if err != nil {
		return nil, err
	}
	return orderV1.NewOrderServiceClient(conn), nil
}

func newCartClient(r registry.Discovery, endpoint string, conns *[]*grpc.ClientConn) (cartV1.CartServiceClient, error) {
	conn, err := newGRPCClient(r, endpoint, conns)
	if err != nil {
		return nil, err
	}
	return cartV1.NewCartServiceClient(conn), nil
}

func newPaymentClient(r registry.Discovery, endpoint string, conns *[]*grpc.ClientConn) (paymentV1.PaymentServiceClient, error) {
	conn, err := newGRPCClient(r, endpoint, conns)
	if err != nil {
		return nil, err
	}
	return paymentV1.NewPaymentServiceClient(conn), nil
}

func newInventoryClient(r registry.Discovery, endpoint string, conns *[]*grpc.ClientConn) (invV1.InventoryClient, error) {
	conn, err := newGRPCClient(r, endpoint, conns)
	if err != nil {
		return nil, err
	}
	return invV1.NewInventoryClient(conn), nil
}

func newPromotionClient(r registry.Discovery, endpoint string, conns *[]*grpc.ClientConn) (promoV1.PromotionServiceClient, error) {
	conn, err := newGRPCClient(r, endpoint, conns)
	if err != nil {
		return nil, err
	}
	return promoV1.NewPromotionServiceClient(conn), nil
}

func newGRPCClient(r registry.Discovery, endpoint string, conns *[]*grpc.ClientConn) (*grpc.ClientConn, error) {
	conn, err := kratosgrpc.DialInsecure(
		context.Background(),
		kratosgrpc.WithEndpoint(endpoint),
		kratosgrpc.WithDiscovery(r),
		kratosgrpc.WithMiddleware(
			recovery.Recovery(),
			circuitbreaker.Client(),
		),
	)
	if err != nil {
		return nil, err
	}
	*conns = append(*conns, conn)
	return conn, nil
}

var ProviderSet = wire.NewSet(NewClients)
