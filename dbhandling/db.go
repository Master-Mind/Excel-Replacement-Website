package dbhandling

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Master-Mind/Excel-Replacement-Website/data_loaders"
	"github.com/Master-Mind/Excel-Replacement-Website/models"
	"github.com/Master-Mind/Excel-Replacement-Website/templs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var NutritionDB *gorm.DB //seperate because there's much, MUCH more data since it's pulling from the USDA database

func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(os.Getenv("DBSTR")), &gorm.Config{})

	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return err
	}

	err = DB.AutoMigrate(&models.Run{}, &models.Workout{}, &models.Set{}, &models.SetType{}, &models.Shoe{})

	if err != nil {
		fmt.Printf("Error migrating database: %v\n", err)
		return err
	}

	NutritionDB, err = gorm.Open(sqlite.Open(os.Getenv("NUTDBSTR")), &gorm.Config{})

	if err != nil {
		fmt.Printf("Error opening nutrition database: %v\n", err)
		return err
	}

	err = NutritionDB.AutoMigrate(&models.Food{}, &models.FoodNutrient{}, &models.Nutrient{})

	if err != nil {
		fmt.Printf("Error migrating nutrition database: %v\n", err)
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

	startYear, err := strconv.Atoi(r.FormValue("startyear"))

	if HandleError(w, r, "Error parsing start year: %v", err) {
		return
	}

	if strings.Contains(strings.ToLower(fileheader.Filename), "run") {
		data, err := data_loaders.LoadRunsSpreadsheet(file, startYear)

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
		data, err := data_loaders.LoadWeightsSpreadsheet(file, startYear)

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

const limit = 30 // Set a limit for the number of runs to display

func RunPage(w http.ResponseWriter, r *http.Request) {
	// Get the workout data from the database
	var runs []models.Run
	var err error

	startDate := r.URL.Query().Get("date")

	if startDate == "" {
		fmt.Print("Date not provided, using default start ID\n")
		return
	}

	DB.Order("Date desc").Where("Date < ?", startDate).Limit(limit).Find(&runs)
	if HandleError(w, r, "Error finding workouts in db: %v", err) {
		return
	}

	fmt.Printf("Found %d runs before %s\n", len(runs), startDate)

	comp := templs.RunPage(runs)
	comp.Render(r.Context(), w)
}

func RunHandler(w http.ResponseWriter, r *http.Request) {
	// Get the workout data from the database
	var runs []models.Run

	err := DB.Order("Date desc").Limit(limit).Find(&runs).Error
	if HandleError(w, r, "Error finding workouts in db: %v", err) {
		return
	}

	fmt.Printf("Found %d workouts\n", len(runs))

	var shoes []models.Shoe
	err = DB.Order("date_purchased desc").Find(&shoes).Error

	if HandleError(w, r, "Error finding shoes in db: %v", err) {
		return
	}

	fmt.Printf("Found %d shoes\n", len(shoes))

	comp := templs.RunDisplay(runs, shoes)
	comp.Render(r.Context(), w)
}

func AddNutritionData(w http.ResponseWriter, r *http.Request) {
	err := data_loaders.TransformNutritionData(NutritionDB)

	if HandleError(w, r, "Error transforming nutrition data: %v", err) {
		return
	}
}
