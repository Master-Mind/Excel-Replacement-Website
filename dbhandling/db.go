package dbhandling

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Master-Mind/Excel-Replacement-Website/data_loaders"
	"github.com/Master-Mind/Excel-Replacement-Website/models"
	"github.com/Master-Mind/Excel-Replacement-Website/templs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(os.Getenv("DBSTR")), &gorm.Config{})

	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return err
	}

	err = DB.AutoMigrate(&models.Run{}, &models.Workout{}, &models.Set{}, &models.SetType{})
	if err != nil {
		fmt.Printf("Error migrating database: %v\n", err)
		return err
	}

	return nil
}

func HandleError(w http.ResponseWriter, r *http.Request, fmtstr string, err error) bool {
	if err != nil {
		fmt.Printf(fmtstr, err)
		w.WriteHeader(http.StatusInternalServerError)

		comp := templs.Error(err)
		comp.Render(r.Context(), w)
		return true
	}

	return false
}

func TransformData(w http.ResponseWriter, r *http.Request) {
	file, fileheader, err := r.FormFile("file")

	if HandleError(w, r, "Error getting file: %v", err) {
		return
	}

	defer file.Close()

	if strings.Contains(strings.ToLower(fileheader.Filename), "run") {
		data, err := data_loaders.LoadRunsSpreadsheet(file, 2021)

		if HandleError(w, r, "Error loading data: %v", err) {
			return
		}

		if r.FormValue("is-preview") != "on" {
			fmt.Printf("Adding %d runs to db\n", len(data))
			err = data_loaders.Add_RunDataToDB(DB, data)

			if HandleError(w, r, "Error adding data to db: %v", err) {
				return
			}
		} else {
			fmt.Printf("Previewing %d runs\n", len(data))
		}

		comp := templs.RunCSVDisplay(data)
		comp.Render(r.Context(), w)
	} else {
		data, err := data_loaders.LoadWeightsSpreadsheet(file, 2022)

		if HandleError(w, r, "Error loading data: %v", err) {
			return
		}

		if r.FormValue("is-preview") != "on" {
			fmt.Printf("Adding %d workouts to db\n", len(data))
			err = data_loaders.Add_WeightDataToDB(DB, data)

			if HandleError(w, r, "Error adding data to db: %v", err) {
				return
			}
		} else {
			fmt.Printf("Previewing %d workouts\n", len(data))
		}

		comp := templs.LiftCSVDisplay(data)
		comp.Render(r.Context(), w)
	}
}

func WorkoutHandler(w http.ResponseWriter, r *http.Request) {
	var setTypes []models.SetType
	err := DB.Find(&setTypes).Error
	if HandleError(w, r, "Error finding set types in db: %v", err) {
		return
	}

	fmt.Printf("Found %d set Types\n", len(setTypes))

	// Get the workout data from the database
	var workouts []models.Workout

	err = DB.Preload("Sets.SetType").Order("Date desc").Find(&workouts).Error
	if HandleError(w, r, "Error finding workouts in db: %v", err) {
		return
	}

	fmt.Printf("Found %d workouts\n", len(workouts))

	comp := templs.LiftDisplay(workouts, setTypes)
	comp.Render(r.Context(), w)
}

func RunHandler(w http.ResponseWriter, r *http.Request) {
	// Get the workout data from the database
	var runs []models.Run

	err := DB.Order("Date desc").Find(&runs).Error
	if HandleError(w, r, "Error finding workouts in db: %v", err) {
		return
	}

	fmt.Printf("Found %d workouts\n", len(runs))

	comp := templs.RunDisplay(runs)
	comp.Render(r.Context(), w)
}
