package internal

import (
	"encoding/csv"
	"fmt"
	"os"
)

// Repo carries a list of revs to be cloned and indexed.
type Repo struct {
	Owner string
	Name  string
	Revs  []string
}

// ReadReposCSV reads a CSV file and returns a list of repositories.
func ReadReposCSV(file string) (repos []Repo, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	m := map[string]*Repo{}
	for _, row := range rows {
		key := fmt.Sprintf("%s/%s", row[0], row[1])

		if repo, ok := m[key]; ok {
			repo.Revs = append(repo.Revs, row[2])
		} else {
			m[key] = &Repo{
				Owner: row[0],
				Name:  row[1],
				Revs:  []string{row[2]},
			}
		}
	}

	for _, repo := range m {
		repos = append(repos, *repo)
	}

	return repos, nil
}
