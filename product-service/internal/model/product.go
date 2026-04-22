package model

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          int64          `gorm:"primarykey"`
	Name        string         `gorm:"size:255;not null"`
	Description string         `gorm:"type:text"`
	Content     string         `gorm:"type:longtext"`
	ImageURL    string         `gorm:"size:500"`
	CategoryID  int64          `gorm:"index"`
	Price       int64          `gorm:"not null"`
	Status      int32          `gorm:"default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type Category struct {
	ID        int64          `gorm:"primarykey"`
	Name      string         `gorm:"size:100;not null"`
	ParentID  int64          `gorm:"index"`
	Level     int32          `gorm:"default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
