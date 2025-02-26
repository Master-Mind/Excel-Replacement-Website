package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func LoadSpreadsheet(filePath string) ([][]string, error) {
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

	return records, nil
}
