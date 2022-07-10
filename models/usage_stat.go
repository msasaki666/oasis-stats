package models

import (
	"time"

	"gorm.io/gorm"
)

type UsageStat struct {
	gorm.Model
	Female    int       `gorm:"not null"`
	Male      int       `gorm:"not null"`
	ScrapedAt time.Time `gorm:"not null"`
}
