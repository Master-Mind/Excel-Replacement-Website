package models

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
	Amount     float64
	Unit       string
}

type Nutrient struct {
	ID         int64
	Name       string
	DVUnit     string
	DailyValue uint
}

type Ingredient struct {
	ID        int64
	FoodID    int64
	FoodToUse Food
	AmountG   float64 // Amount in grams
	RecipeID  int64   // Foreign key to Recipe
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
	HeightCM             uint
	WeightKG             uint
	BodyFatPercent       uint
	TargetBodyFatPercent uint
} // Add other fields as necessary
