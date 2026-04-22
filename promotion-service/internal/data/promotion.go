package data

import (
	"context"
	"time"

	"promotion-service/internal/biz"
	"promotion-service/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type promotionRepo struct {
	data *Data
	log  *log.Helper
}

func NewPromotionRepo(data *Data, logger log.Logger) biz.PromotionRepo {
	return &promotionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *promotionRepo) CreateCoupon(ctx context.Context, c *biz.Coupon) (*biz.Coupon, error) {
	po := &model.Coupon{
		Name:         c.Name,
		Type:         c.Type,
		Threshold:    c.Threshold,
		Discount:     c.Discount,
		StartTime:    c.StartTime,
		EndTime:      c.EndTime,
		TotalCount:   c.TotalCount,
		PerUserLimit: c.PerUserLimit,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	c.ID = po.ID
	return c, nil
}

func (r *promotionRepo) ListCoupons(ctx context.Context, pageNum, pageSize int32) ([]*biz.Coupon, int32, error) {
	var pos []model.Coupon
	var count int64
	r.data.db.WithContext(ctx).Model(&model.Coupon{}).Count(&count)
	err := r.data.db.WithContext(ctx).Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).Find(&pos).Error
	if err != nil {
		return nil, 0, err
	}
	var res []*biz.Coupon
	for _, po := range pos {
		res = append(res, &biz.Coupon{
			ID:        po.ID,
			Name:      po.Name,
			Type:      po.Type,
			Threshold: po.Threshold,
			Discount:  po.Discount,
			StartTime: po.StartTime,
			EndTime:   po.EndTime,
		})
	}
	return res, int32(count), nil
}

func (r *promotionRepo) GetCoupon(ctx context.Context, id int64) (*biz.Coupon, error) {
	var po model.Coupon
	if err := r.data.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return &biz.Coupon{
		ID:            po.ID,
		Name:          po.Name,
		Type:          po.Type,
		Threshold:     po.Threshold,
		Discount:      po.Discount,
		StartTime:     po.StartTime,
		EndTime:       po.EndTime,
		TotalCount:    po.TotalCount,
		ReceivedCount: po.ReceivedCount,
		PerUserLimit:  po.PerUserLimit,
	}, nil
}

func (r *promotionRepo) GrantCoupon(ctx context.Context, userID, couponID int64) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create user coupon
		po := &model.UserCoupon{
			UserID:   userID,
			CouponID: couponID,
			Status:   0,
		}
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		// Increment received count
		return tx.Model(&model.Coupon{}).
			Where("id = ?", couponID).
			UpdateColumn("received_count", gorm.Expr("received_count + ?", 1)).Error
	})
}

func (r *promotionRepo) UseCoupon(ctx context.Context, userID, couponID int64, orderID string) error {
	return r.data.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("user_id = ? AND coupon_id = ? AND status = 0", userID, couponID).
		Updates(map[string]interface{}{
			"status":    1,
			"used_time": time.Now(),
			"order_id":  orderID,
		}).Error
}

func (r *promotionRepo) GetUserCoupon(ctx context.Context, userID, couponID int64) (*biz.UserCoupon, error) {
	var po model.UserCoupon
	if err := r.data.db.WithContext(ctx).Where("user_id = ? AND coupon_id = ?", userID, couponID).First(&po).Error; err != nil {
		return nil, err
	}
	return &biz.UserCoupon{
		ID:       po.ID,
		UserID:   po.UserID,
		CouponID: po.CouponID,
		Status:   po.Status,
		UsedTime: po.UsedTime,
		OrderID:  po.OrderID,
	}, nil
}

func (r *promotionRepo) GetUserCouponCount(ctx context.Context, userID, couponID int64) (int64, error) {
	var count int64
	err := r.data.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("user_id = ? AND coupon_id = ? AND status = 0", userID, couponID).
		Count(&count).Error
	return count, err
}

func (r *promotionRepo) UpdateReceivedCount(ctx context.Context, couponID int64) error {
	return r.data.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("id = ?", couponID).
		UpdateColumn("received_count", gorm.Expr("received_count + ?", 1)).Error
}
