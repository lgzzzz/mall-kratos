package data

import (
	"context"

	"cart-service/internal/biz"
	"cart-service/internal/model"

	"github.com/go-kratos/kratos/v2/log"
)

type cartRepo struct {
	data *Data
	log  *log.Helper
}

func NewCartRepo(data *Data, logger log.Logger) biz.CartRepo {
	return &cartRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *cartRepo) AddCart(ctx context.Context, c *biz.Cart) (*biz.Cart, error) {
	var po model.Cart
	// 检查是否已存在同类商品
	err := r.data.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", c.UserID, c.ProductID).First(&po).Error
	if err == nil {
		// 存在则增加数量
		po.Quantity += c.Quantity
		if err := r.data.db.WithContext(ctx).Save(&po).Error; err != nil {
			return nil, err
		}
	} else {
		// 不存在则创建
		po = model.Cart{
			UserID:    c.UserID,
			ProductID: c.ProductID,
			Quantity:  c.Quantity,
			Selected:  c.Selected,
		}
		if err := r.data.db.WithContext(ctx).Create(&po).Error; err != nil {
			return nil, err
		}
	}
	c.ID = po.ID
	c.Quantity = po.Quantity
	return c, nil
}

func (r *cartRepo) UpdateCart(ctx context.Context, c *biz.Cart) (*biz.Cart, error) {
	po := &model.Cart{ID: c.ID}
	if err := r.data.db.WithContext(ctx).Model(po).Updates(map[string]interface{}{
		"quantity": c.Quantity,
		"selected": c.Selected,
	}).Error; err != nil {
		return nil, err
	}
	// 重新加载
	if err := r.data.db.WithContext(ctx).First(po, c.ID).Error; err != nil {
		return nil, err
	}
	return &biz.Cart{
		ID:        po.ID,
		UserID:    po.UserID,
		ProductID: po.ProductID,
		Quantity:  po.Quantity,
		Selected:  po.Selected,
	}, nil
}

func (r *cartRepo) DeleteCart(ctx context.Context, id int64) error {
	return r.data.db.WithContext(ctx).Delete(&model.Cart{}, id).Error
}

func (r *cartRepo) ListCart(ctx context.Context, userID int64) ([]*biz.Cart, error) {
	var pos []model.Cart
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	var res []*biz.Cart
	for _, po := range pos {
		res = append(res, &biz.Cart{
			ID:        po.ID,
			UserID:    po.UserID,
			ProductID: po.ProductID,
			Quantity:  po.Quantity,
			Selected:  po.Selected,
		})
	}
	return res, nil
}

func (r *cartRepo) ClearCart(ctx context.Context, userID int64) error {
	return r.data.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Cart{}).Error
}
