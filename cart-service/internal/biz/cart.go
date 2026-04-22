package biz

import (
	"context"

	"cart-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

const maxCartQuantity = 99

type Cart struct {
	ID        int64
	UserID    int64
	ProductID int64
	Quantity  int32
	Selected  bool
}

type CartRepo interface {
	AddCart(ctx context.Context, c *Cart) (*Cart, error)
	UpdateCart(ctx context.Context, c *Cart) (*Cart, error)
	DeleteCart(ctx context.Context, id int64) error
	ListCart(ctx context.Context, userID int64) ([]*Cart, error)
	ClearCart(ctx context.Context, userID int64) error
}

type CartUseCase struct {
	repo        CartRepo
	productRepo ProductRepo
	log         *log.Helper
}

func NewCartUseCase(repo CartRepo, productRepo ProductRepo, logger log.Logger) *CartUseCase {
	return &CartUseCase{
		repo:        repo,
		productRepo: productRepo,
		log:         log.NewHelper(logger),
	}
}

func (uc *CartUseCase) AddCart(ctx context.Context, c *Cart) (*Cart, error) {
	if c.Quantity <= 0 || c.Quantity > maxCartQuantity {
		return nil, conf.ErrQuantityExceeded
	}

	product, err := uc.productRepo.GetProduct(ctx, c.ProductID)
	if err != nil {
		return nil, conf.ErrProductNotFound
	}
	if product.Status != 1 {
		return nil, conf.ErrProductNotFound
	}

	return uc.repo.AddCart(ctx, c)
}

func (uc *CartUseCase) UpdateCart(ctx context.Context, c *Cart) (*Cart, error) {
	if c.Quantity < 0 || c.Quantity > maxCartQuantity {
		return nil, conf.ErrQuantityExceeded
	}
	return uc.repo.UpdateCart(ctx, c)
}

func (uc *CartUseCase) DeleteCart(ctx context.Context, id int64) error {
	return uc.repo.DeleteCart(ctx, id)
}

func (uc *CartUseCase) ListCart(ctx context.Context, userID int64) ([]*Cart, error) {
	return uc.repo.ListCart(ctx, userID)
}

func (uc *CartUseCase) ClearCart(ctx context.Context, userID int64) error {
	return uc.repo.ClearCart(ctx, userID)
}
