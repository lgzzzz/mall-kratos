package data

import (
	"context"

	"product-service/internal/biz"
	"product-service/internal/model"

	"github.com/go-kratos/kratos/v2/log"
)

type productRepo struct {
	data *Data
	log  *log.Helper
}

type categoryRepo struct {
	data *Data
	log  *log.Helper
}

func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *productRepo) CreateProduct(ctx context.Context, p *biz.Product) (*biz.Product, error) {
	po := &model.Product{
		Name:        p.Name,
		Description: p.Description,
		Content:     p.Content,
		ImageURL:    p.ImageURL,
		CategoryID:  p.CategoryID,
		Price:       p.Price,
		Status:      p.Status,
	}
	result := r.data.db.WithContext(ctx).Create(po)
	if result.Error != nil {
		return nil, result.Error
	}
	p.ID = po.ID
	p.CreatedAt = po.CreatedAt
	p.UpdatedAt = po.UpdatedAt
	return p, nil
}

func (r *productRepo) UpdateProduct(ctx context.Context, p *biz.Product) (*biz.Product, error) {
	po := &model.Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Content:     p.Content,
		ImageURL:    p.ImageURL,
		CategoryID:  p.CategoryID,
		Price:       p.Price,
		Status:      p.Status,
	}
	result := r.data.db.WithContext(ctx).Save(po)
	if result.Error != nil {
		return nil, result.Error
	}
	p.UpdatedAt = po.UpdatedAt
	return p, nil
}

func (r *productRepo) DeleteProduct(ctx context.Context, id int64) error {
	return r.data.db.WithContext(ctx).Delete(&model.Product{}, id).Error
}

func (r *productRepo) GetProduct(ctx context.Context, id int64) (*biz.Product, error) {
	var po model.Product
	if err := r.data.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return &biz.Product{
		ID:          po.ID,
		Name:        po.Name,
		Description: po.Description,
		Content:     po.Content,
		ImageURL:    po.ImageURL,
		CategoryID:  po.CategoryID,
		Price:       po.Price,
		Status:      po.Status,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}, nil
}

func (r *productRepo) ListProducts(ctx context.Context, pageNum, pageSize int32, categoryID int64, keyword string) ([]*biz.Product, int32, error) {
	var pos []model.Product
	db := r.data.db.WithContext(ctx)
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	
	var count int64
	db.Model(&model.Product{}).Count(&count)
	
	err := db.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).Find(&pos).Error
	if err != nil {
		return nil, 0, err
	}
	
	var res []*biz.Product
	for _, po := range pos {
		res = append(res, &biz.Product{
			ID:          po.ID,
			Name:        po.Name,
			Description: po.Description,
			Content:     po.Content,
			ImageURL:    po.ImageURL,
			CategoryID:  po.CategoryID,
			Price:       po.Price,
			Status:      po.Status,
			CreatedAt:   po.CreatedAt,
			UpdatedAt:   po.UpdatedAt,
		})
	}
	return res, int32(count), nil
}

func (r *categoryRepo) CreateCategory(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	po := &model.Category{
		Name:     c.Name,
		ParentID: c.ParentID,
		Level:    c.Level,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	c.ID = po.ID
	c.CreatedAt = po.CreatedAt
	c.UpdatedAt = po.UpdatedAt
	return c, nil
}

func (r *categoryRepo) UpdateCategory(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	po := &model.Category{
		ID:       c.ID,
		Name:     c.Name,
		ParentID: c.ParentID,
		Level:    c.Level,
	}
	if err := r.data.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	c.UpdatedAt = po.UpdatedAt
	return c, nil
}

func (r *categoryRepo) DeleteCategory(ctx context.Context, id int64) error {
	return r.data.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}

func (r *categoryRepo) GetCategory(ctx context.Context, id int64) (*biz.Category, error) {
	var po model.Category
	if err := r.data.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return &biz.Category{
		ID:        po.ID,
		Name:      po.Name,
		ParentID:  po.ParentID,
		Level:     po.Level,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}, nil
}

func (r *categoryRepo) ListCategories(ctx context.Context, parentID int64) ([]*biz.Category, error) {
	var pos []model.Category
	if err := r.data.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&pos).Error; err != nil {
		return nil, err
	}
	var res []*biz.Category
	for _, po := range pos {
		res = append(res, &biz.Category{
			ID:        po.ID,
			Name:      po.Name,
			ParentID:  po.ParentID,
			Level:     po.Level,
			CreatedAt: po.CreatedAt,
			UpdatedAt: po.UpdatedAt,
		})
	}
	return res, nil
}
