package internal

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
)

// TestCase links a definition location and an expected set of reference locations.
type TestCase struct {
	Definition Location
	References []Location
}

// ReadTestCaseCSV reads a CSV file and returns a list of test cases.
func ReadTestCaseCSV(dataDir, file string) (testCases []TestCase, err error) {
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
		definition := Location{
			Repo:      row[0],
			Rev:       row[1],
			Path:      row[2],
			Line:      line,
			Character: character,
		}

		references, err := ReadLocationsCSV(filepath.Join(dataDir, row[5]))
		if err != nil {
			return nil, err
		}

		testCases = append(testCases, TestCase{
			Definition: definition,
			References: references,
		})
	}

	return testCases, nil
}
