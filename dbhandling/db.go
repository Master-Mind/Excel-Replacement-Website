package dbhandling

import (
	"database/sql"
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
var NutritionDB *sql.DB //seperate because there's much, MUCH more data since it's pulling from the USDA database
var NutdbInitted = false

func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(os.Getenv("DBSTR")), &gorm.Config{})

	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return err
	}

	err = DB.AutoMigrate(&models.Run{}, &models.Workout{}, &models.Set{},
		&models.SetType{}, &models.Shoe{})

	if err != nil {
		fmt.Printf("Error migrating database: %v\n", err)
		return err
	}

	NutritionDB, err = sql.Open("sqlite3", os.Getenv("NUTDBSTR"))

	if err != nil {
		fmt.Printf("Error opening nutrition database: %v\n", err)
		return err
	}

	initStatement :=
		`CREATE TABLE IF NOT EXISTS foods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS nutrients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		dv_unit TEXT NOT NULL,
		daily_value REAL
	);
	CREATE TABLE IF NOT EXISTS food_nutrients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		food_id INTEGER NOT NULL,
		nutrient_id INTEGER NOT NULL,
		amount REAL NOT NULL,
		unit TEXT NOT NULL,
		FOREIGN KEY (food_id) REFERENCES foods(id),
		FOREIGN KEY (nutrient_id) REFERENCES nutrients(id)
	);
	CREATE TABLE IF NOT EXISTS recipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS ingredients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		food_id INTEGER NOT NULL,
		recipe_id INTEGER NOT NULL,
		amount_g REAL NOT NULL,
		FOREIGN KEY (food_id) REFERENCES foods(id),
		FOREIGN KEY (recipe_id) REFERENCES recipes(id)
	);
	CREATE TABLE IF NOT EXISTS person (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		age INTEGER,
		is_male INTEGER,
		weight_kg REAL,
		height_cm REAL,
		body_fat_percent REAL,
		target_body_fat_percent REAL
		);
	CREATE TABLE IF NOT EXISTS diet_days (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
		);
	CREATE TABLE IF NOT EXISTS diet_day_recipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		diet_day_id INTEGER NOT NULL,
		recipe_id INTEGER NOT NULL,
		FOREIGN KEY (diet_day_id) REFERENCES diet_days(id),
		FOREIGN KEY (recipe_id) REFERENCES recipes(id)
		);
	CREATE TABLE IF NOT EXISTS diet_weeks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
		);
	CREATE TABLE IF NOT EXISTS diet_week_days (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		diet_week_id INTEGER NOT NULL,
		diet_day_id INTEGER NOT NULL,
		FOREIGN KEY (diet_day_id) REFERENCES diet_days(id),
		FOREIGN KEY (diet_week_id) REFERENCES diet_weeks(id)
		);
	CREATE TABLE IF NOT EXISTS excercises (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		METS INTEGER NOT NULL
		);
	CREATE TABLE IF NOT EXISTS diet_day_exercises (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		diet_day_id INTEGER NOT NULL,
		exercise_id INTEGER NOT NULL,
		duration REAL NOT NULL,
		FOREIGN KEY (diet_day_id) REFERENCES diet_days(id),
		FOREIGN KEY (exercise_id) REFERENCES excercises(id)
		);`

	_, err = NutritionDB.Exec(initStatement)

	if err != nil {
		fmt.Printf("Error initializing nutrition database: %v\n", err)
		return err
	}

	rows, err := NutritionDB.Query("SELECT COUNT (*) FROM foods;")

	if err != nil {
		NutdbInitted = false
	} else {
		defer rows.Close()
		rows.Next()
		var count int
		err = rows.Scan(&count)
		NutdbInitted = err == nil && count > 0

		if !NutdbInitted {
			fmt.Printf("%v\n", err)
		}
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

const workoutPageLimit = 10

func WorkoutPage(w http.ResponseWriter, r *http.Request) {
	var setTypes []models.SetType
	err := DB.Find(&setTypes).Error
	if HandleError(w, r, "Error finding set types in db: %v", err) {
		return
	}

	fmt.Printf("Found %d set Types\n", len(setTypes))

	// Get the workout data from the database
	var workouts []models.Workout
	startDate := r.URL.Query().Get("date")

	err = DB.Preload("Sets.SetType").Order("Date desc").Where("Date < ?", startDate).Limit(workoutPageLimit).Find(&workouts).Error
	if HandleError(w, r, "Error finding workouts in db: %v", err) {
		return
	}

	fmt.Printf("Found %d workouts\n", len(workouts))

	comp := templs.WorkoutPage(workouts, setTypes)
	comp.Render(r.Context(), w)
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

	err = DB.Preload("Sets.SetType").Order("Date desc").Limit(workoutPageLimit).Find(&workouts).Error
	if HandleError(w, r, "Error finding workouts in db: %v", err) {
		return
	}

	fmt.Printf("Found %d workouts\n", len(workouts))

	comp := templs.LiftDisplay(workouts, setTypes)
	comp.Render(r.Context(), w)
}

// Limit for the number of runs to fetch.
// Infinite scroll breaks the shoe milage calculation, so set the limit to "infinity" for now.
const limit = 10000

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
