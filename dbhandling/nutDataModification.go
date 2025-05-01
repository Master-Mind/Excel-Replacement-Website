package dbhandling

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Master-Mind/Excel-Replacement-Website/models"
	"github.com/Master-Mind/Excel-Replacement-Website/templs"
)

func FoodRecomendationHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if HandleError(w, r, "Error parsing form: %v", err) {
		return
	}
	query := r.FormValue("food-search")

	var recs []models.Food

	if err := NutritionDB.Limit(10).Where("description LIKE ?", "%"+query+"%").Find(&recs).Error; err != nil {
		HandleError(w, r, "Error finding food recommendations: %v", err)
		return
	}

	fmt.Printf("Found %d food recommendations for query '%s'\n", len(recs), query)

	comp := templs.FoodRecList(recs)

	if err := comp.Render(r.Context(), w); err != nil {
		fmt.Printf("Error rendering food recommendations: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func DietPageHandler(w http.ResponseWriter, r *http.Request) {
	var foods []models.Food

	if err := NutritionDB.Limit(1).Find(&foods).Error; err != nil {
		HandleError(w, r, "Error finding food recommendations: %v", err)
		return
	}

	var recipes []models.Recipe

	if err := NutritionDB.Preload("Ingredients.FoodToUse").Find(&recipes).Error; err != nil {
		HandleError(w, r, "Error finding recipes: %v\n", err)
		return
	}

	comp := templs.Diet(recipes, len(foods) > 0)

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

	if err := NutritionDB.Create(&recipe).Error; err != nil {
		HandleError(w, r, "Error creating recipe: %v\n", err)
		return
	}

	fmt.Printf("Created new recipe: %s\n", recipe.Name)

	comp := templs.RecipeDisplay(recipe)
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

	if err := NutritionDB.Delete(&models.Recipe{}, recipeID).Error; err != nil {
		HandleError(w, r, "Error deleting recipe: %v", err)
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

	var recipe models.Recipe
	if err := NutritionDB.First(&recipe, recipeID).Error; err != nil {
		HandleError(w, r, "Error finding recipe: %v", err)
		return
	}

	recipe.Name = newName

	if err := NutritionDB.Save(&recipe).Error; err != nil {
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

	recipeID, err := strconv.ParseUint(r.URL.Query().Get("recipeid"), 10, 32)

	if HandleError(w, r, "Error parsing recipe ID: %v", err) {
		return
	}

	foodName := r.FormValue("food-search")

	var food models.Food

	if foodName == "" {
		http.Error(w, "Food name is required", http.StatusBadRequest)
		return
	}

	err = NutritionDB.Where("description = ?", foodName).First(&food).Error

	if HandleError(w, r, "Error parsing food ID: %v", err) {
		return
	}

	ingredient := models.Ingredient{
		FoodToUse: food,
		FoodID:    uint(food.ID),
		RecipeID:  uint(recipeID),

		AmountG: 100, // Default amount, can be adjusted later
	}

	if err := NutritionDB.Create(&ingredient).Error; err != nil {
		HandleError(w, r, "Error adding ingredient to recipe: %v", err)
		return
	}

	fmt.Printf("Added ingredient with food ID %d to recipe ID %d\n", food.ID, recipeID)

	w.WriteHeader(http.StatusOK) // Send a 200 OK response to indicate success
}
