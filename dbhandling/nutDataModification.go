package dbhandling

import (
	"fmt"
	"net/http"

	"github.com/Master-Mind/Excel-Replacement-Website/templs"
)

func DietPageHandler(w http.ResponseWriter, r *http.Request) {
	comp := templs.Diet(NutritionDB != nil)

	if err := comp.Render(r.Context(), w); err != nil {
		fmt.Printf("Error rendering diet page: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
