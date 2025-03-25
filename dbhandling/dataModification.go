package dbhandling

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Master-Mind/Excel-Replacement-Website/models"
	"github.com/Master-Mind/Excel-Replacement-Website/templs"
)

func AddRun(w http.ResponseWriter, r *http.Request) {
	newrun := models.Run{}
	var err error

	r.ParseForm()

	fmt.Printf("Adding new run with data: %v\n", r.Form)

	newrun.Date, err = time.Parse("2006-01-02", r.Form.Get("date"))

	if HandleError(w, r, "Error parsing date: %v", err) {
		return
	}

	newrun.Distance, err = strconv.ParseFloat(r.Form.Get("distance"), 64)

	if HandleError(w, r, "Error parsing distance: %v", err) {
		return
	}

	newrun.Minutes, err = strconv.Atoi(r.Form.Get("minutes"))

	if HandleError(w, r, "Error parsing minutes: %v", err) {
		return
	}

	if err := DB.Create(&newrun).Error; err != nil {
		HandleError(w, r, "Error adding run to db: %v", err)
		return
	}

	comp := templs.RunRow(newrun)
	comp.Render(r.Context(), w) // Render the component to show the updated list of runs
}

func RemoveRun(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	runID, err := strconv.Atoi(r.Form.Get("id"))

	if HandleError(w, r, "Error parsing run ID: %v", err) {
		return
	}

	if err := DB.Delete(&models.Run{}, runID).Error; err != nil {
		HandleError(w, r, "Error deleting run from db: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK) // Send a 200 OK response to indicate success
}

func NewShoe(w http.ResponseWriter, r *http.Request) {
	newshoe := models.Shoe{}
	var err error

	r.ParseForm()

	fmt.Printf("Adding new shoe with data: %v\n", r.Form)

	newshoe.Name = r.Form.Get("name")

	newshoe.MinMilage, err = strconv.Atoi(r.Form.Get("min-milage"))
	if HandleError(w, r, "Error parsing min milage: %v", err) {
		return
	}

	newshoe.MaxMilage, err = strconv.Atoi(r.Form.Get("max-milage"))
	if HandleError(w, r, "Error parsing max milage: %v", err) {
		return
	}

	newshoe.DatePurchased, err = time.Parse("2006-01-02", r.Form.Get("purchase-date"))
	if HandleError(w, r, "Error parsing date purchased: %v", err) {
		return
	}

	if r.Form.Get("retire-date") != "" {
		newshoe.DateRetired, err = time.Parse("2006-01-02", r.Form.Get("retire-date"))
		if HandleError(w, r, "Error parsing date retired: %v", err) {
			return
		}
	}

	if err := DB.Create(&newshoe).Error; err != nil {
		HandleError(w, r, "Error adding shoe to db: %v", err)
		return
	}

	comp := templs.ShoeRow(newshoe, 0)
	comp.Render(r.Context(), w) // Render the component to show the updated list of shoes
}

func DeleteShoe(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	shoeID, err := strconv.Atoi(r.Form.Get("id"))

	if HandleError(w, r, "Error parsing shoe ID: %v", err) {
		return
	}

	if err := DB.Delete(&models.Shoe{}, shoeID).Error; err != nil {
		HandleError(w, r, "Error deleting shoe from db: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK) // Send a 200 OK response to indicate success
}
