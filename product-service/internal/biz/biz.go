package biz

import (
	"context"

	"github.com/google/wire"
)

type InventoryRepo interface {
	CreateInventory(ctx context.Context, skuID int64) error
}

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewProductUseCase, NewCategoryUseCase)

type CategoryUseCase struct {
	repo CategoryRepo
}

func NewCategoryUseCase(repo CategoryRepo) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

func (uc *CategoryUseCase) CreateCategory(ctx context.Context, c *Category) (*Category, error) {
	return uc.repo.CreateCategory(ctx, c)
}

func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, c *Category) (*Category, error) {
	return uc.repo.UpdateCategory(ctx, c)
}

func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id int64) error {
	return uc.repo.DeleteCategory(ctx, id)
}

func (uc *CategoryUseCase) GetCategory(ctx context.Context, id int64) (*Category, error) {
	return uc.repo.GetCategory(ctx, id)
}

func (uc *CategoryUseCase) ListCategories(ctx context.Context, parentID int64) ([]*Category, error) {
	return uc.repo.ListCategories(ctx, parentID)
}
