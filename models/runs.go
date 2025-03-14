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
