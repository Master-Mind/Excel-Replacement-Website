package data_loaders

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Master-Mind/Excel-Replacement-Website/models"
	"gorm.io/gorm"
)

func TransformNutritionData(nutdb *gorm.DB) error {
	file, err := os.ReadFile(os.Getenv("NUTRITION_DATA_FILE"))

	if err != nil {
		fmt.Printf("Error reading nutrition data file: %v\n", err)
		return err
	}

	var nutritionData map[string]interface{}

	if err := json.Unmarshal(file, &nutritionData); err != nil {
		fmt.Printf("Error unmarshaling nutrition data: %v\n", err)
		return err
	}

	foundationFoodsArray := nutritionData["FoundationFoods"].([]interface{})

	if foundationFoodsArray == nil {
		fmt.Printf("Couldn't find foundation foods array\n")
		return errors.New("couldn't find foundation foods array")
	}

	foods := make([]models.Food, 0)
	nutrientMap := make(map[string]models.Nutrient)
	nutrients := make([]models.Nutrient, 0)

	err = nutdb.Find(&nutrients).Error

	if err != nil {
		fmt.Printf("Error finding nutrients: %v\n", err)
		return err
	}

	for _, nutrient := range nutrients {
		nutrientMap[nutrient.Name] = nutrient
	}

	for _, item := range foundationFoodsArray {
		foodItem := item.(map[string]interface{})

		if foodItem == nil {
			fmt.Printf("Couldn't find food\n")
			return errors.New("couldn't find food")
		}

		var newFood models.Food

		newFood.Description = foodItem["description"].(string)

		foodNutrients := foodItem["foodNutrients"].([]interface{})

		if foodNutrients == nil {
			fmt.Printf("Couldn't find food nutrients\n")
			return errors.New("couldn't find food nutrients")
		}

		for _, nutrient := range foodNutrients {
			nutrientItem := nutrient.(map[string]interface{})
			nutrientInfo := nutrientItem["nutrient"].(map[string]interface{})

			nutrientName := nutrientInfo["name"].(string)
			nutrientUnit := nutrientInfo["unitName"].(string)

			//fmt.Printf("Processing nutrient: %s (%s)\n", nutrientName, nutrientUnit)

			if nutrientItem["amount"] == nil {
				//fmt.Printf("Nutrient amount is nil for %s\n", nutrientName)
				continue
			}

			nutrientAmount := nutrientItem["amount"].(float64)

			nut, hasNut := nutrientMap[nutrientName]

			// Energy has two entries, one for kcal and one for kJ. We only want kcal.
			if nutrientName == "Energy" && nutrientUnit != "kcal" {
				continue
			}

			if !hasNut {
				nut = models.Nutrient{
					Name:   nutrientName,
					DVUnit: nutrientUnit,
				}
				if err := nutdb.Where("name = ? AND dv_unit = ?", nutrientName, nutrientUnit).FirstOrCreate(&nut).Error; err != nil {
					fmt.Printf("Error creating/finding nutrient: %v\n", err)
					continue
				}
				nutrientMap[nutrientName] = nut
			}

			amountNum := float64(nutrientAmount)

			if amountNum < 0.01 {
				continue
			}

			newNutrient := models.FoodNutrient{
				FoodToUse:  newFood,
				NutrientID: nut.ID,
				Unit:       nutrientUnit,
				Amount:     amountNum,
			}

			newFood.Nutrients = append(newFood.Nutrients, newNutrient)
		}

		foods = append(foods, newFood)
	}

	fmt.Printf("Found %d foods\n", len(foods))

	// Gorm can't do the create right (it keeps trying to run an insert with all the nutrients, causing a crash)
	// Create the foods with a raw SQL query
	const batchSize = 1000

	for i := 0; i < len(foods); i += batchSize {
		end := max(i+batchSize, len(foods))

		querryStr := "INSERT INTO foods (description) VALUES "

		for j := i; j < end; j++ {
			querryStr += fmt.Sprintf("('%s')", foods[j].Description)

			if j < end-1 {
				querryStr += ", "
			}
		}
		querryStr += " RETURNING id ON CONFLICT (description) DO NOTHING;"

		if err := nutdb.Exec(querryStr).Error; err != nil {
			fmt.Printf("Error inserting foods: %v\n", err)
			return err
		}
	}

	return nil
}
