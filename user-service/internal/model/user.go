package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64          `gorm:"primarykey"`
	Username  string         `gorm:"size:64;uniqueIndex;not null"`
	Password  string         `gorm:"size:255;not null"`
	Nickname  string         `gorm:"size:64"`
	Email     string         `gorm:"size:128"`
	Mobile    string         `gorm:"size:20;uniqueIndex"`
	Avatar    string         `gorm:"size:500"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Address struct {
	ID        int64          `gorm:"primarykey"`
	UserID    int64          `gorm:"index;not null"`
	Name      string         `gorm:"size:64;not null"`
	Mobile    string         `gorm:"size:20;not null"`
	Province  string         `gorm:"size:64"`
	City      string         `gorm:"size:64"`
	District  string         `gorm:"size:64"`
	Detail    string         `gorm:"size:255"`
	IsDefault bool           `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
