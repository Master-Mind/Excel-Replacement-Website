package dbhandling

import (
	"fmt"
	"net/http"

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
	comp := templs.Diet(len(foods) > 0)

	if err := comp.Render(r.Context(), w); err != nil {
		fmt.Printf("Error rendering diet page: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
