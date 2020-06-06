package internal

import (
	"encoding/csv"
	"os"
	"strconv"
)

// Location represents the first character in a source code range.
type Location struct {
	Repo      string
	Rev       string
	Path      string
	Line      int
	Character int
}

// ReadLocationsCSV reads a CSV file and returns a list of locations.
func ReadLocationsCSV(file string) (locations []Location, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		line, _ := strconv.Atoi(row[3])
		character, _ := strconv.Atoi(row[4])

		locations = append(locations, Location{
			Repo:      row[0],
			Rev:       row[1],
			Path:      row[2],
			Line:      line,
			Character: character,
		})
	}

	return locations, nil
}
