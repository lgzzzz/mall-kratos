package model

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	ID             int64          `gorm:"primarykey"`
	Name           string         `gorm:"size:100;not null"`
	Type           int32          `gorm:"not null"`
	Threshold      int64          `gorm:"not null"`
	Discount       int64          `gorm:"not null"`
	StartTime      time.Time      `gorm:"not null"`
	EndTime        time.Time      `gorm:"not null"`
	TotalCount     int64          `gorm:"default:-1"` // -1: unlimited
	ReceivedCount  int64          `gorm:"default:0"`
	PerUserLimit   int32          `gorm:"default:0"` // 0: unlimited
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type UserCoupon struct {
	ID        int64          `gorm:"primarykey"`
	UserID    int64          `gorm:"index;not null"`
	CouponID  int64          `gorm:"index;not null"`
	Status    int32          `gorm:"default:0"` // 0: 未使用, 1: 已使用, 2: 已过期
	UsedTime  time.Time
	OrderID   string         `gorm:"size:64"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
