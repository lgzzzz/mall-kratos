package model

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID            string         `gorm:"primarykey;size:64"`
	OrderID       string         `gorm:"size:64;index;not null"`
	Amount        int64          `gorm:"not null"`
	Currency      string         `gorm:"size:10;default:'CNY'"`
	Status        int32          `gorm:"default:1"`
	TransactionID string         `gorm:"size:128"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
