package data_loaders

import (
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CSVSet struct {
	Intensity int
	Reps      int
	SetType   string
}

type CSVWorkout struct {
	WorkoutDate time.Time
	Sets        []CSVSet
}

type CSVRun struct {
	RunDate  time.Time
	Distance float64
	Minutes  int
}

func LoadWeightsSpreadsheet(file multipart.File, startYear int) ([]CSVWorkout, error) {
	// Parse the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	dateMatch, err := regexp.Compile(`\d{1,2}-\w{3}`)

	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	//parse my janky ass lifting spreadsheet
	csvMonth2GoMonth := map[string]time.Month{
		"Jan": 1,
		"Feb": 2,
		"Mar": 3,
		"Apr": 4,
		"May": 5,
		"Jun": 6,
		"Jul": 7,
		"Aug": 8,
		"Sep": 9,
		"Oct": 10,
		"Nov": 11,
		"Dec": 12,
	}

	curMonth := time.January

	workouts := make([]CSVWorkout, 0)

	//scan for dates
	for row := range records {
		if dateMatch.MatchString(records[row][0]) {
			var curWorkout CSVWorkout
			daymonth := strings.Split(records[row][0], "-")
			day, err := strconv.Atoi(daymonth[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse day: %w", err)
			}

			month := csvMonth2GoMonth[daymonth[1]]

			if month < curMonth {
				startYear++
			}

			curMonth = month

			curWorkout.WorkoutDate = time.Date(startYear, month, day, 0, 0, 0, 0, time.UTC)

			row++
			//scan each table
			colStarts := make([]int, 0)
			colEnds := make([]int, 0)
			liftTypes := make([]string, 0)

			//scan table headers
			for col := 0; col < len(records[row]); col++ {
				if records[row][col] == "Intensity" {
					colStarts = append(colStarts, col)
				} else if records[row][col] == "Adj total" {
					colEnds = append(colEnds, col)
				} else if records[row][col] != "" && !strings.HasPrefix(records[row][col], "Column") {
					liftTypes = append(liftTypes, records[row][col])
				}
			}

			row++

			for ; row < len(records) && records[row][0] != ""; row++ {
				for i := 0; i < len(colStarts); i++ {
					intensity, err := strconv.Atoi(records[row][colStarts[i]])

					if err != nil {
						continue
					}

					for j := colStarts[i] + 1; j < colEnds[i]; j++ {
						if records[row][j] == "" {
							continue
						}

						reps, err := strconv.Atoi(records[row][j])

						if err != nil {
							return nil, fmt.Errorf("failed to parse reps at (%d,%d): %w", row, j, err)
						}

						curWorkout.Sets = append(curWorkout.Sets, CSVSet{
							Intensity: intensity,
							Reps:      reps,
							SetType:   liftTypes[i],
						})
					}
				}
			}

			workouts = append(workouts, curWorkout)
		}
	}

	return workouts, nil
}

func LoadRunsSpreadsheet(file multipart.File, startYear int) ([]CSVRun, error) {
	// Parse the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	//parse my janky ass lifting spreadsheet
	csvMonth2GoMonth := map[string]time.Month{
		"Jan": 1,
		"Feb": 2,
		"Mar": 3,
		"Apr": 4,
		"May": 5,
		"Jun": 6,
		"Jul": 7,
		"Aug": 8,
		"Sep": 9,
		"Oct": 10,
		"Nov": 11,
		"Dec": 12,
	}

	curMonth := time.January

	runs := make([]CSVRun, 0)

	for row := 1; row < len(records); row++ {
		var curRun CSVRun
		daymonth := strings.Split(records[row][0], "-")
		day, err := strconv.Atoi(daymonth[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse day: %w", err)
		}

		month := csvMonth2GoMonth[daymonth[1]]

		if month < curMonth {
			startYear++
		}

		curMonth = month

		curRun.RunDate = time.Date(startYear, month, day, 0, 0, 0, 0, time.UTC)

		curRun.Distance, err = strconv.ParseFloat(records[row][1], 64)

		if err != nil {
			fmt.Printf("Failed to parse distance at %d", row)
		}

		curRun.Minutes, err = strconv.Atoi(records[row][2])

		if err != nil {
			fmt.Printf("Failed to parse minutes at %d", row)
		}

		runs = append(runs, curRun)
	}

	return runs, nil
}
