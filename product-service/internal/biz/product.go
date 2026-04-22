package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type Product struct {
	ID          int64
	Name        string
	Description string
	Content     string
	ImageURL    string
	CategoryID  int64
	Price       int64
	Status      int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
	ID        int64
	Name      string
	ParentID  int64
	Level     int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductRepo interface {
	CreateProduct(ctx context.Context, p *Product) (*Product, error)
	UpdateProduct(ctx context.Context, p *Product) (*Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProduct(ctx context.Context, id int64) (*Product, error)
	ListProducts(ctx context.Context, pageNum, pageSize int32, categoryID int64, keyword string) ([]*Product, int32, error)
}

type CategoryRepo interface {
	CreateCategory(ctx context.Context, c *Category) (*Category, error)
	UpdateCategory(ctx context.Context, c *Category) (*Category, error)
	DeleteCategory(ctx context.Context, id int64) error
	GetCategory(ctx context.Context, id int64) (*Category, error)
	ListCategories(ctx context.Context, parentID int64) ([]*Category, error)
}

type ProductUseCase struct {
	productRepo   ProductRepo
	categoryRepo  CategoryRepo
	inventoryRepo InventoryRepo
	log           *log.Helper
}

func NewProductUseCase(productRepo ProductRepo, categoryRepo CategoryRepo, inventoryRepo InventoryRepo, logger log.Logger) *ProductUseCase {
	return &ProductUseCase{
		productRepo:   productRepo,
		categoryRepo:  categoryRepo,
		inventoryRepo: inventoryRepo,
		log:           log.NewHelper(logger),
	}
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	created, err := uc.productRepo.CreateProduct(ctx, p)
	if err != nil {
		return nil, err
	}
	// Initialize inventory for new product
	if err := uc.inventoryRepo.CreateInventory(ctx, created.ID); err != nil {
		uc.log.Warnf("failed to create inventory for product %d: %v", created.ID, err)
	}
	return created, nil
}

func (uc *ProductUseCase) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	return uc.productRepo.UpdateProduct(ctx, p)
}

func (uc *ProductUseCase) DeleteProduct(ctx context.Context, id int64) error {
	return uc.productRepo.DeleteProduct(ctx, id)
}

func (uc *ProductUseCase) GetProduct(ctx context.Context, id int64) (*Product, error) {
	return uc.productRepo.GetProduct(ctx, id)
}

func (uc *ProductUseCase) ListProducts(ctx context.Context, pageNum, pageSize int32, categoryID int64, keyword string) ([]*Product, int32, error) {
	return uc.productRepo.ListProducts(ctx, pageNum, pageSize, categoryID, keyword)
}

func (uc *ProductUseCase) CreateCategory(ctx context.Context, c *Category) (*Category, error) {
	return uc.categoryRepo.CreateCategory(ctx, c)
}

func (uc *ProductUseCase) UpdateCategory(ctx context.Context, c *Category) (*Category, error) {
	return uc.categoryRepo.UpdateCategory(ctx, c)
}

func (uc *ProductUseCase) DeleteCategory(ctx context.Context, id int64) error {
	return uc.categoryRepo.DeleteCategory(ctx, id)
}

func (uc *ProductUseCase) GetCategory(ctx context.Context, id int64) (*Category, error) {
	return uc.categoryRepo.GetCategory(ctx, id)
}

func (uc *ProductUseCase) ListCategories(ctx context.Context, parentID int64) ([]*Category, error) {
	return uc.categoryRepo.ListCategories(ctx, parentID)
}
