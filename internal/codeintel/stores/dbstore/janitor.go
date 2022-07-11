package dbstore

import (
	"database/sql"
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

// ScanSourcedCommits scans triples of repository ids/repository names/commits from the
// return value of `*Store.query`. The output of this function is ordered by repository
// identifier, then by commit.
func ScanSourcedCommits(rows *sql.Rows, queryErr error) (_ []SourcedCommits, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	sourcedCommitsMap := map[int]SourcedCommits{}
	for rows.Next() {
		var repositoryID int
		var repositoryName string
		var commit string
		if err := rows.Scan(&repositoryID, &repositoryName, &commit); err != nil {
			return nil, err
		}

		sourcedCommitsMap[repositoryID] = SourcedCommits{
			RepositoryID:   repositoryID,
			RepositoryName: repositoryName,
			Commits:        append(sourcedCommitsMap[repositoryID].Commits, commit),
		}
	}

	flattened := make([]SourcedCommits, 0, len(sourcedCommitsMap))
	for _, sourcedCommits := range sourcedCommitsMap {
		sort.Strings(sourcedCommits.Commits)
		flattened = append(flattened, sourcedCommits)
	}

	sort.Slice(flattened, func(i, j int) bool {
		return flattened[i].RepositoryID < flattened[j].RepositoryID
	})
	return flattened, nil
}
