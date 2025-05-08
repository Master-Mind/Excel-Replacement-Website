package models

import (
	"gorm.io/gorm"
)

type Food struct {
	gorm.Model
	Description string         `gorm:"not null"`
	Nutrients   []FoodNutrient `gorm:"foreignKey:FoodID"`
}

type FoodNutrient struct {
	gorm.Model
	FoodID     uint `gorm:"not null"`
	FoodToUse  Food `gorm:"foreignKey:FoodID;references:ID"`
	NutrientID uint `gorm:"not null"`
	Nutrient   Nutrient
	Amount     float64 `gorm:"not null"`
	Unit       string  `gorm:"not null"`
}

type Nutrient struct {
	gorm.Model
	Name       string `gorm:"not null"`
	DVUnit     string `gorm:"not null"`
	DailyValue uint
}

type Ingredient struct {
	gorm.Model
	FoodID    uint    `gorm:"not null"`
	FoodToUse Food    `gorm:"foreignKey:FoodID;references:ID"`
	AmountG   float64 `gorm:"not null"` // Amount in grams
	RecipeID  uint    `gorm:"not null"` // Foreign key to Recipe
}

type Recipe struct {
	gorm.Model
	Name        string       `gorm:"not null"`
	Ingredients []Ingredient `gorm:"foreignKey:RecipeID"`
}

type DietDay struct {
	gorm.Model
	Name  string   `gorm:"not null"`
	Meals []Recipe `gorm:"many2many:diet_day_meals"` // Meals for the day, e.g., breakfast, lunch, dinner
	Foods []Food   `gorm:"many2many:diet_day_foods"` // in addition to meals, in case there are snacks or other foods
}

type Person struct {
	gorm.Model
	Name                 string `gorm:"not null"`
	Age                  uint   `gorm:"not null"`
	IsMale               bool   `gorm:"not null"`
	HeightCM             uint   `gorm:"not null"`
	WeightKG             uint   `gorm:"not null"`
	BodyFatPercent       uint   `gorm:"not null"`
	TargetBodyFatPercent uint
} // Add other fields as necessary
