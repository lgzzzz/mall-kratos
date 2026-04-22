package model

import (
	"time"

	"gorm.io/gorm"
)

type Cart struct {
	ID        int64          `gorm:"primarykey"`
	UserID    int64          `gorm:"index;not null"`
	ProductID int64          `gorm:"index;not null"`
	Quantity  int32          `gorm:"not null;default:1"`
	Selected  bool           `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
