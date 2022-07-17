package models

import (
	"time"

	"gorm.io/gorm"
)

type UsageStat struct {
	// json tag設定したかったので
	ID        uint           `gorm:"primarykey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Female    int            `gorm:"not null"`
	Male      int            `gorm:"not null"`
	ScrapedAt time.Time      `gorm:"not null"`
	Weekday   int
}
