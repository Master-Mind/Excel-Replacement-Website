package models

import (
	"time"

	"gorm.io/gorm"
)

type Run struct {
	gorm.Model
	Date      time.Time `gorm:"not null"`
	Distance  float64   `gorm:"not null"`
	Minutes   int       `gorm:"not null"`
	Elevation int
}

type Shoe struct {
	gorm.Model
	Name          string    `gorm:"not null"`
	MinMilage     int       `gorm:"not null"`
	MaxMilage     int       `gorm:"not null"`
	DatePurchased time.Time `gorm:"not null"`
	DateRetired   time.Time
}
