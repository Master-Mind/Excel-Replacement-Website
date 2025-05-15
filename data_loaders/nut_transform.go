package data_loaders

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Master-Mind/Excel-Replacement-Website/models"
)

func TransformNutritionData(nutdb *sql.DB) error {
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

	foods := make([]string, 0)
	nutrientMap := make(map[string]models.Nutrient)

	transaction, err := nutdb.Begin()

	if err != nil {
		fmt.Printf("Error starting transaction: %v\n", err)
		return err
	}

	foodInsertStmt, err := transaction.Prepare("INSERT INTO foods (id, description) VALUES (?, ?);")

	if err != nil {
		fmt.Printf("Error preparing food insert statement: %v\n", err)
		return err
	}

	foodNutrientID := 0

	foodNutInsertStmt, err := transaction.Prepare("INSERT INTO food_nutrients (id, food_id, nutrient_id, amount, unit) VALUES (?, ?, ?, ?, ?);")

	if err != nil {
		fmt.Printf("Error preparing food nutrient insert statement: %v\n", err)
		return err
	}

	nutrientID := 0
	nutrientInsertStmt, err := transaction.Prepare("INSERT INTO nutrients (id, name, dv_unit) VALUES (?, ?, ?);")

	if err != nil {
		fmt.Printf("Error preparing nutrient insert statement: %v\n", err)
		return err
	}

	defer foodInsertStmt.Close()
	defer foodNutInsertStmt.Close()

	for foodID, item := range foundationFoodsArray {
		foodItem := item.(map[string]interface{})

		if foodItem == nil {
			fmt.Printf("Couldn't find food\n")
			return errors.New("couldn't find food")
		}

		newFoodDesc := foodItem["description"].(string)
		_, err := foodInsertStmt.Exec(foodID, newFoodDesc)

		if err != nil {
			fmt.Printf("Error creating food: %v\n", err)
			continue
		}

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

				_, err := nutrientInsertStmt.Exec(nutrientID, nutrientName, nutrientUnit)

				if err != nil {
					fmt.Printf("Error creating/finding nutrient: %v\n", err)
					continue
				}

				nutrientID++

				nut.ID = int64(nutrientID)

				nutrientMap[nutrientName] = nut
			}

			amountNum := float64(nutrientAmount)

			if amountNum < 0.01 {
				continue
			}

			_, err := foodNutInsertStmt.Exec(foodNutrientID, foodID, nut.ID, amountNum, nutrientUnit)

			foodNutrientID++

			if err != nil {
				fmt.Printf("Error creating food nutrient: %v\n", err)
				continue
			}
		}

		foods = append(foods, newFoodDesc)
	}

	err = transaction.Commit()

	if err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		return err
	}

	fmt.Printf("Found %d foods\n", len(foods))

	return nil
}
