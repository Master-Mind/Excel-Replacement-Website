package data_loaders

import (
	"fmt"

	"github.com/Master-Mind/Excel-Replacement-Website/models"
	"gorm.io/gorm"
)

func Add_RunDataToDB(db *gorm.DB, data []CSVRun) error {
	runs := make([]models.Run, len(data))
	for i, run := range data {
		runs[i] = models.Run{
			Date:     run.RunDate,
			Distance: run.Distance,
			Minutes:  run.Minutes,
		}
	}

	if err := db.Create(&runs).Error; err != nil {
		fmt.Printf("Error adding runs to db: %v", err)
	}

	return nil
}

func Add_WeightDataToDB(db *gorm.DB, data []CSVWorkout) error {
	weights := make([]models.Workout, len(data))

	//create set types manually if they don't exist in the database (easier to enter units and whatnow)
	var setTypes []models.SetType
	if err := db.Find(&setTypes).Error; err != nil {
		fmt.Printf("Error finding set types in db: %v", err)
	}

	expectedSetTypes := []models.SetType{
		{Name: "Squat", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Bench Press", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Deadlift", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Shoulder Press", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Pull Up", RepUnit: "reps", IntensityUnit: "%BW"},
		{Name: "Chin Up", RepUnit: "reps", IntensityUnit: "%BW"},
		{Name: "Calf Raise", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Lunge", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Leg raises", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "DB Press", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "In. DB Press", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "De. DB Press", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Bench Press", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Lateral Raise", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Weighted Crunches", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "DB Curl", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Bench DB Curl", RepUnit: "reps", IntensityUnit: "lbs"},
		{Name: "Farmer Carry", RepUnit: "secs", IntensityUnit: "lbs"}}

	if len(setTypes) != len(expectedSetTypes) {
		for _, expectedSetType := range expectedSetTypes {
			var existingSetType models.SetType
			temp := db.Where("name = ?", expectedSetType.Name)
			if err := temp.First(&existingSetType).Error; err != nil {
				// Set type does not exist, create it
				newSetType := models.SetType{
					Name:          expectedSetType.Name,
					RepUnit:       expectedSetType.RepUnit,
					IntensityUnit: expectedSetType.IntensityUnit,
				}
				if err := db.Create(&newSetType).Error; err != nil {
					fmt.Printf("Error creating set type %s: %v", expectedSetType.Name, err)
					return err
				}
			}
		}
	}

	if err := db.Find(&setTypes).Error; err != nil {
		fmt.Printf("Error finding set types in db: %v", err)
	}

	for i, workout := range data {
		weights[i] = models.Workout{
			Date: workout.WorkoutDate,
			Sets: make([]models.Set, len(workout.Sets)),
		}

		for j, set := range workout.Sets {
			weights[i].Sets[j] = models.Set{
				Intensity: set.Intensity,
				Reps:      set.Reps,
			}

			// Check if the set type exists in the database
			for _, setType := range setTypes {
				if setType.Name == set.SetType {
					setType.Sets = append(setType.Sets, weights[i].Sets[j])
				}
			}
		}
	}

	if err := db.Create(&weights).Error; err != nil {
		fmt.Printf("Error adding weights to db: %v", err)
	}

	return nil
}
