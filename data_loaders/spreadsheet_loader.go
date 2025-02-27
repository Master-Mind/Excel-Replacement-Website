package data_loaders

import (
	"encoding/csv"
	"fmt"
	"os"
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

func LoadWeightsSpreadsheet(filePath string, startYear int) ([]CSVWorkout, error) {
	// Open the spreadsheet file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

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
				} else if !strings.HasPrefix(records[row][col], "Column") {
					liftTypes = append(liftTypes, records[row][col])
				}
			}

			for i := 0; i < len(colStarts); i++ {
			}

			workouts = append(workouts, curWorkout)
		}
	}

	return workouts, nil
}
