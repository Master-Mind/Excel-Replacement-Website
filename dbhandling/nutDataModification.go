package dbhandling

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Master-Mind/Excel-Replacement-Website/models"
	"github.com/Master-Mind/Excel-Replacement-Website/templs"
	"gonum.org/v1/gonum/unit"
)

const Pound = unit.Mass(0.45359237)
const Ounce = unit.Mass(0.02834952)
const IU = unit.Mass(0.00000001)   // International Unit, used for vitamins and some minerals, technically not valid because it varies by substance, but used here for simplicity
const Calorie = unit.Energy(4.184) // 1 Calorie = 4.184 Joules

func MakeGonumMass(amount float64, unitStr string) (unit.Mass, error) {

	switch unitStr {
	case "g":
		return unit.Mass(amount) * unit.Gram, nil
	case "kg":
		return unit.Mass(amount) * unit.Kilogram, nil
	case "mg":
		return unit.Mass(amount) * unit.Gram * unit.Milli, nil
	case "oz":
		return unit.Mass(amount) * Ounce, nil
	case "lb":
		return unit.Mass(amount) * Pound, nil
	case "µg":
		return unit.Mass(amount) * unit.Gram * unit.Micro, nil
	default:
		return -1, fmt.Errorf("unknown unit: %s", unitStr)
	}
}

func MakeGonumEnergy(amount float64, unitStr string) (unit.Energy, error) {
	switch unitStr {
	case "kcal":
		return unit.Energy(amount) * Calorie * unit.Kilo, nil
	case "cal":
		return unit.Energy(amount) * Calorie, nil
	case "kJ":
		return unit.Energy(amount) * unit.Joule * unit.Kilo, nil
	case "J":
		return unit.Energy(amount) * unit.Joule, nil
	default:
		return -1, fmt.Errorf("unknown energy unit: %s", unitStr)
	}
}

func getNutrients() ([]models.Nutrient, error) {
	var nutrients []models.Nutrient
	rows, err := NutritionDB.Query("SELECT * FROM nutrients")

	if err != nil {
		return nil, fmt.Errorf("error fetching nutrients: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var nutrient models.Nutrient
		var dailyValue sql.NullInt64
		var dvUnit string

		if err := rows.Scan(&nutrient.ID, &nutrient.Name, &dvUnit, &dailyValue); err != nil {
			return nil, fmt.Errorf("error scanning nutrient: %v", err)
		}

		if dailyValue.Valid {
			nutrient.DailyValue, err = MakeGonumMass(float64(dailyValue.Int64), dvUnit) // Convert to gonum unit
			if err != nil {
				//try to convert energy
				nutrient.DVEnergy, err = MakeGonumEnergy(float64(dailyValue.Int64), dvUnit)

				if err != nil {
					//ignore the error, set DailyValue to 0
					fmt.Printf("Error converting daily value for nutrient %s: %v, setting to 0\n", nutrient.Name, err)
					nutrient.DailyValue = 0
				}
			}
		} else {
			nutrient.DailyValue = 0
		}

		nutrients = append(nutrients, nutrient)
	}

	return nutrients, nil
}

func FoodRecomendationHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	err := r.ParseForm()
	if HandleError(w, r, "Error parsing form: %v", err) {
		return
	}
	query := r.FormValue("food-search")

	var recs []string

	rows, err := NutritionDB.Query("SELECT description FROM foods WHERE description LIKE ? LIMIT 10", "%"+query+"%")

	if err != nil {
		HandleError(w, r, "Error finding food recommendations: %v", err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var food string
		if err := rows.Scan(&food); err != nil {
			HandleError(w, r, "Error scanning food recommendation: %v", err)
			return
		}
		recs = append(recs, food)
	}

	fmt.Printf("Found %d food recommendations for query '%s' in %vμs\n", len(recs), query, time.Since(startTime).Microseconds())

	comp := templs.FoodRecList(recs)

	if err := comp.Render(r.Context(), w); err != nil {
		fmt.Printf("Error rendering food recommendations: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DietPageHandler(w http.ResponseWriter, r *http.Request) {
	var recipes []models.Recipe
	var recipeMap = make(map[int64]*models.Recipe) // Map to store recipes by ID for ingredient association

	rows, err := NutritionDB.Query("SELECT * FROM recipes")
	if err != nil {
		HandleError(w, r, "Error fetching recipes: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var recipe models.Recipe
		if err := rows.Scan(&recipe.ID, &recipe.Name); err != nil {
			HandleError(w, r, "Error scanning recipe: %v", err)
			return
		}

		fmt.Printf("Found recipe: %v, %s\n", recipe.ID, recipe.Name)
		recipes = append(recipes, recipe)
		recipeMap[recipe.ID] = &recipes[len(recipes)-1] // Store the recipe in the map
	}

	nutrients, err := getNutrients()
	if err != nil {
		HandleError(w, r, "Error fetching nutrients: %v", err)
		return
	}

	// Fetch ingredients for each recipe
	rows, err = NutritionDB.Query("SELECT id, food_id, recipe_id, amount_g FROM ingredients")
	if err != nil {
		HandleError(w, r, "Error fetching ingredients: %v", err)
		return
	}

	var foodQueryStr = "SELECT id, description FROM foods WHERE id IN ("
	var foundIngredient bool

	for rows.Next() {
		var ingredient models.Ingredient
		var amountG float64

		if err := rows.Scan(&ingredient.ID, &ingredient.FoodID, &ingredient.RecipeID, &amountG); err != nil {
			HandleError(w, r, "Error scanning ingredient: %v", err)
			return
		}

		ingredient.Amount = unit.Mass(amountG) * unit.Gram // Convert amount to gonum unit

		if recipe, exists := recipeMap[ingredient.RecipeID]; exists {
			ingredient.FoodToUse = models.Food{ID: ingredient.FoodID}   // Initialize Food struct
			foodQueryStr += fmt.Sprintf("%d,", ingredient.FoodID)       // Append food ID to the query string
			recipe.Ingredients = append(recipe.Ingredients, ingredient) // Append ingredient to the recipe
			foundIngredient = true
		} else {
			fmt.Printf("Recipe ID %d not found for ingredient with food ID %d\n", ingredient.RecipeID, ingredient.FoodID)
		}
	}

	rows.Close()
	if foundIngredient {
		// Remove the last comma and close the query string
		foodQueryStr = foodQueryStr[:len(foodQueryStr)-1] + ")"
		rows, err = NutritionDB.Query(foodQueryStr)

		if err != nil {
			HandleError(w, r, "Error fetching food descriptions: %v", err)
			return
		}
		defer rows.Close()

		// Map to store food descriptions by ID
		foodDescriptions := make(map[int64]string)

		for rows.Next() {
			var food models.Food
			if err := rows.Scan(&food.ID, &food.Description); err != nil {
				HandleError(w, r, "Error scanning food: %v", err)
				return
			}

			foodDescriptions[food.ID] = food.Description
		}
		// Assign food descriptions to the ingredients
		for _, recipe := range recipes {
			for i, ingredient := range recipe.Ingredients {
				if desc, exists := foodDescriptions[ingredient.FoodID]; exists {
					recipe.Ingredients[i].FoodToUse.Description = desc // Assign the food description

					recipe.Ingredients[i].FoodToUse.Nutrients = make([]models.FoodNutrient, 0) // Initialize the slice
					// Fetch the nutrients for this food
					nutrientRows, err := NutritionDB.Query("SELECT id, nutrient_id, amount, unit FROM food_nutrients WHERE food_id = ?", ingredient.FoodID)
					if err != nil {
						HandleError(w, r, "Error fetching food nutrients: %v", err)
						return
					}
					defer nutrientRows.Close()
					for nutrientRows.Next() {
						var foodNutrient models.FoodNutrient
						var amount float64
						var unitStr string

						if err := nutrientRows.Scan(&foodNutrient.ID, &foodNutrient.NutrientID, &amount, &unitStr); err != nil {
							HandleError(w, r, "Error scanning food nutrient: %v", err)
							return
						}

						foodNutrient.Amount, err = MakeGonumMass(amount, unitStr) // Convert to gonum unit
						if err != nil {
							//ignore the error, set Amount to 0
							fmt.Printf("Error converting amount for nutrient ID %d: %v, setting to 0\n", foodNutrient.NutrientID, err)
							foodNutrient.Amount = 0
						}

						// Find the nutrient by ID
						foodNutrient.Nutrient = nutrients[foodNutrient.NutrientID-1] // Assuming nutrient IDs are sequential

						recipe.Ingredients[i].FoodToUse.Nutrients = append(recipe.Ingredients[i].FoodToUse.Nutrients, foodNutrient)
					}
				} else {
					fmt.Printf("Food ID %d not found in food descriptions\n", ingredient.FoodID)
				}
			}
		}

		//assign nutrients to the food

	}

	comp := templs.Diet(NutdbInitted, recipes, nutrients)

	if err := comp.Render(r.Context(), w); err != nil {
		fmt.Printf("Error rendering diet page: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func AddRecipe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if HandleError(w, r, "Error parsing form: %v\n", err) {
		return
	}

	var recipe models.Recipe
	recipe.Name = r.FormValue("recipe-name")

	_, err = NutritionDB.Exec("INSERT INTO recipes (name) VALUES (?)", recipe.Name)

	if err != nil {
		HandleError(w, r, "Error creating recipe: %v\n", err)
		return
	}

	nutrients, err := getNutrients()
	if err != nil {
		HandleError(w, r, "Error fetching nutrients: %v\n", err)
		return
	}

	fmt.Printf("Created new recipe: %s\n", recipe.Name)

	nutMap := make(map[string]models.Nutrient)

	for _, nut := range nutrients {
		nutMap[nut.Name] = nut
	}

	comp := templs.RecipeDisplay(recipe, nutMap)
	if err := comp.Render(r.Context(), w); err != nil {
		fmt.Printf("Error rendering recipe display: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if HandleError(w, r, "Error parsing form: %v", err) {
		return
	}

	recipeID := r.FormValue("id")

	if recipeID == "" {
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	_, err = NutritionDB.Exec("DELETE FROM recipes WHERE id = ?", recipeID)

	if err != nil {
		HandleError(w, r, "Error deleting recipe: %v", err)
		return
	}

	_, err = NutritionDB.Exec("DELETE FROM ingredients WHERE recipe_id = ?", recipeID)
	if err != nil {
		HandleError(w, r, "Error deleting ingredients: %v", err)
		return
	}

	fmt.Printf("Deleted recipe with ID: %s\n", recipeID)

	w.WriteHeader(http.StatusOK) // Send a 200 OK response to indicate success
}

func UpdateRecipeName(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if HandleError(w, r, "Error parsing form: %v", err) {
		return
	}

	recipeID := r.FormValue("id")
	newName := r.FormValue("recipe-name")

	if recipeID == "" || newName == "" {
		http.Error(w, "Recipe ID and new name are required", http.StatusBadRequest)
		return
	}

	_, err = NutritionDB.Exec("UPDATE recipes SET name = ? WHERE id = ?", newName, recipeID)

	if err != nil {
		HandleError(w, r, "Error updating recipe name: %v", err)
		return
	}

	fmt.Printf("Updated recipe ID %s to new name: %s\n", recipeID, newName)

	w.WriteHeader(http.StatusOK) // Send a 200 OK response to indicate success
}

func AddIngredient(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if HandleError(w, r, "Error parsing form: %v", err) {
		return
	}

	recipeID, err := strconv.ParseInt(r.URL.Query().Get("recipe_id"), 10, 64)

	if HandleError(w, r, "Error parsing recipe ID: %v", err) {
		return
	}

	foodName := r.FormValue("food-search")

	var foodID int64

	if foodName == "" {
		http.Error(w, "Food name is required", http.StatusBadRequest)
		return
	}

	var row = NutritionDB.QueryRow("SELECT id FROM foods WHERE description = ?", foodName)
	err = row.Scan(&foodID)

	if HandleError(w, r, "Error parsing food ID: %v", err) {
		return
	}

	result, err := NutritionDB.Exec("INSERT INTO ingredients (food_id, recipe_id, amount_g) VALUES (?, ?, ?)", foodID, recipeID, 100.0)
	if err != nil {
		HandleError(w, r, "Error adding ingredient: %v", err)
		return
	}

	fmt.Printf("Added ingredient with food ID %d to recipe ID %d\n", foodID, recipeID)

	var ingredient models.Ingredient
	ingredient.FoodID = foodID
	ingredient.RecipeID = recipeID
	ingredient.Amount = 100.0 * unit.Gram // Default amount in grams
	ingredient.ID, err = result.LastInsertId()
	ingredient.FoodToUse = models.Food{ID: foodID, Description: foodName} // Initialize Food struct
	if err != nil {
		HandleError(w, r, "Error getting last insert ID: %v", err)
		return
	}

	comp := templs.IngredientDisplay(ingredient)
	if err := comp.Render(r.Context(), w); err != nil {
		fmt.Printf("Error rendering ingredient display: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if HandleError(w, r, "Error parsing form: %v", err) {
		return
	}

	ingredientID := r.FormValue("id")

	if ingredientID == "" {
		http.Error(w, "Ingredient ID is required", http.StatusBadRequest)
		return
	}

	_, err = NutritionDB.Exec("DELETE FROM ingredients WHERE id = ?", ingredientID)

	if err != nil {
		HandleError(w, r, "Error deleting ingredient: %v", err)
		return
	}

	fmt.Printf("Deleted ingredient with ID: %s\n", ingredientID)

	w.WriteHeader(http.StatusOK) // Send a 200 OK response to indicate success
}

func UpdateIngredientAmount(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if HandleError(w, r, "Error parsing form: %v", err) {
		return
	}

	ingredientID := r.FormValue("id")
	newAmount := r.FormValue("ingredient-amount")

	if ingredientID == "" || newAmount == "" {
		http.Error(w, "Ingredient ID and new amount are required", http.StatusBadRequest)
		return
	}

	_, err = NutritionDB.Exec("UPDATE ingredients SET amount_g = ? WHERE id = ?", newAmount, ingredientID)

	if err != nil {
		HandleError(w, r, "Error updating ingredient amount: %v", err)
		return
	}

	fmt.Printf("Updated ingredient ID %s to new amount: %s\n", ingredientID, newAmount)

	w.WriteHeader(http.StatusOK) // Send a 200 OK response to indicate success
}
