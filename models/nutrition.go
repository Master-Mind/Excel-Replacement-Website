package models

import "gonum.org/v1/gonum/unit"

type Food struct {
	ID          int64
	Description string
	Nutrients   []FoodNutrient
}

type FoodNutrient struct {
	ID         int64
	FoodID     int64
	FoodToUse  Food
	NutrientID int64
	Nutrient   Nutrient
	Amount     unit.Mass
}

type Nutrient struct {
	ID         int64
	Name       string
	DailyValue unit.Mass
	DVEnergy   unit.Energy // Daily value energy in kcal or kJ
}

type Ingredient struct {
	ID        int64
	FoodID    int64
	FoodToUse Food
	Amount    unit.Mass
	RecipeID  int64 // Foreign key to Recipe
}

type Recipe struct {
	ID          int64
	Name        string
	Ingredients []Ingredient
}

type DietDay struct {
	Name  string
	Meals []Recipe // Meals for the day, e.g., breakfast, lunch, dinner
	Foods []Food   // in addition to meals, in case there are snacks or other foods
}

type Person struct {
	Name                 string
	Age                  uint
	IsMale               bool
	Height               unit.Length
	Weight               unit.Length
	BodyFatPercent       float32
	TargetBodyFatPercent float32
} // Add other fields as necessary
