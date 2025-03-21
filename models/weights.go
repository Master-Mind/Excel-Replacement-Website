package models

import (
	"time"

	"gorm.io/gorm"
)

type Set struct {
	gorm.Model
	Intensity int     `gorm:"not null"`
	Reps      int     `gorm:"not null"`
	WorkoutID uint    `gorm:"not null"`
	SetTypeID uint    `gorm:"not null"`
	SetType   SetType `gorm:"foreignKey:SetTypeID"`
	Workout   Workout `gorm:"foreignKey:WorkoutID"` // Establishing the relationship with Workout
}

type Workout struct {
	gorm.Model
	Date time.Time `gorm:"not null"`
	Sets []Set
}

type SetType struct {
	gorm.Model
	Name          string `gorm:"not null"`
	RepUnit       string
	IntensityUnit string `gorm:"not null"`
	Sets          []Set
}
