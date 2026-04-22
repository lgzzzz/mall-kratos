package biz

import (
	"context"
	"time"

	"promotion-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

type Coupon struct {
	ID           int64
	Name         string
	Type         int32 // 1: 满减, 2: 折扣
	Threshold    int64
	Discount     int64
	StartTime    time.Time
	EndTime      time.Time
	TotalCount   int64 // -1: unlimited
	ReceivedCount int64
	PerUserLimit int32 // 0: unlimited
}

type UserCoupon struct {
	ID        int64
	UserID    int64
	CouponID  int64
	Status    int32 // 0: 未使用, 1: 已使用, 2: 已过期
	UsedTime  time.Time
	OrderID   string
}

type PromotionRepo interface {
	CreateCoupon(ctx context.Context, c *Coupon) (*Coupon, error)
	ListCoupons(ctx context.Context, pageNum, pageSize int32) ([]*Coupon, int32, error)
	GetCoupon(ctx context.Context, id int64) (*Coupon, error)
	GrantCoupon(ctx context.Context, userID, couponID int64) error
	UseCoupon(ctx context.Context, userID, couponID int64, orderID string) error
	GetUserCoupon(ctx context.Context, userID, couponID int64) (*UserCoupon, error)
	GetUserCouponCount(ctx context.Context, userID, couponID int64) (int64, error)
	UpdateReceivedCount(ctx context.Context, couponID int64) error
}

type PromotionUseCase struct {
	repo PromotionRepo
	log  *log.Helper
}

func NewPromotionUseCase(repo PromotionRepo, logger log.Logger) *PromotionUseCase {
	return &PromotionUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *PromotionUseCase) CreateCoupon(ctx context.Context, c *Coupon) (*Coupon, error) {
	return uc.repo.CreateCoupon(ctx, c)
}

func (uc *PromotionUseCase) ListCoupons(ctx context.Context, pageNum, pageSize int32) ([]*Coupon, int32, error) {
	return uc.repo.ListCoupons(ctx, pageNum, pageSize)
}

func (uc *PromotionUseCase) GrantCoupon(ctx context.Context, userID, couponID int64) error {
	c, err := uc.repo.GetCoupon(ctx, couponID)
	if err != nil {
		return conf.ErrCouponNotFound
	}
	// Check time validity
	now := time.Now()
	if now.Before(c.StartTime) || now.After(c.EndTime) {
		return conf.ErrCouponExpired
	}
	// Check total count (-1 means unlimited)
	if c.TotalCount >= 0 && c.ReceivedCount >= c.TotalCount {
		return conf.ErrCouponExhausted
	}
	// Check per user limit (0 means unlimited)
	if c.PerUserLimit > 0 {
		count, err := uc.repo.GetUserCouponCount(ctx, userID, couponID)
		if err == nil && count >= int64(c.PerUserLimit) {
			return conf.ErrCouponLimit
		}
	}
	return uc.repo.GrantCoupon(ctx, userID, couponID)
}

func (uc *PromotionUseCase) UseCoupon(ctx context.Context, userID, couponID int64, orderID string) error {
	return uc.repo.UseCoupon(ctx, userID, couponID, orderID)
}

func (uc *PromotionUseCase) CalculateDiscount(ctx context.Context, userID, totalAmount int64, couponIDs []int64) (int64, int64, error) {
	var totalDiscount int64
	for _, cid := range couponIDs {
		c, err := uc.repo.GetCoupon(ctx, cid)
		if err != nil {
			continue
		}
		// 验证用户是否有这张券
		ucp, err := uc.repo.GetUserCoupon(ctx, userID, cid)
		if err != nil || ucp.Status != 0 {
			continue
		}
		// 验证时间
		now := time.Now()
		if now.Before(c.StartTime) || now.After(c.EndTime) {
			continue
		}
		// 验证门槛
		if totalAmount < c.Threshold {
			continue
		}
		// 计算优惠
		if c.Type == 1 { // 满减
			totalDiscount += c.Discount
		} else if c.Type == 2 { // 折扣 (假设 Discount 为 80 表示 8 折)
			totalDiscount += totalAmount * (100 - c.Discount) / 100
		}
	}
	finalAmount := totalAmount - totalDiscount
	if finalAmount < 0 {
		finalAmount = 0
	}
	return totalDiscount, finalAmount, nil
}
