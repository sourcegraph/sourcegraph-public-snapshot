package main

import "fmt"

type Git interface {
	LogReverse(repo string, commit string, n int) ([]LogEntry, error)
	RevList(repo string, commit string) ([]string, error)
}

type DB interface {
	GetCommit(givenCommit string) (commit string, height int, present bool, err error)
	InsertCommit(commit string, height int, ancestor string) error
	GetBlob(hop string, status StatusAD, path string) (id int, found bool, err error)
	UpdateBlobHops(id int, status StatusAD, hop string) error
	InsertBlob(blob Blob) error
	AppendHop(hops []string, status StatusAD, hop string) error
	Search(hops []string) ([]Blob, error)
}

type LogEntry struct {
	Commit       string
	PathStatuses []PathStatus
}

type PathStatus struct {
	Path   string
	Status StatusAMD
}

type CommitStatus struct {
	Commit string
	Status StatusAMD
}

type Blob struct {
	path    string
	added   []string
	deleted []string
}

type StatusAMD int

const (
	AddedAMD    StatusAMD = 0
	ModifiedAMD StatusAMD = 1
	DeletedAMD  StatusAMD = 2
)

type StatusAD int

const (
	AddedAD   StatusAD = 0
	DeletedAD StatusAD = 1
)

const NULL = "0000000000000000000000000000000000000000"

func Index(git Git, db DB, repo string, givenCommit string) error {
	tipCommit := NULL
	tipHeight := 0
	missingCount := 0
	revs, err := git.RevList(repo, givenCommit)
	if err != nil {
		return err
	}
	for _, commit := range revs {
		if commit, height, present, err := db.GetCommit(commit); err != nil {
			return err
		} else if present {
			tipCommit = commit
			tipHeight = height
			break
		}
		missingCount += 1
	}

	entries, err := git.LogReverse(repo, givenCommit, missingCount)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		hops, err := getHops(db, tipCommit)
		if err != nil {
			return err
		}

		r := ruler(tipHeight + 1)
		if r >= len(hops) {
			return fmt.Errorf("ruler(%d) = %d is out of range of len(hops) = %d", tipHeight+1, r, len(hops))
		}

		db.AppendHop(hops[0:r], AddedAD, entry.Commit)
		db.AppendHop(hops[0:r], DeletedAD, entry.Commit)

		// Could delete redundant hops here, but skipping instead because it would make the DB interface
		// more complex. Besides, it doesn't change the asymptotic space complexity.

		for _, pathStatus := range entry.PathStatuses {
			if pathStatus.Status == DeletedAMD || pathStatus.Status == ModifiedAMD {
				for _, hop := range hops {
					// TODO time this with some kind of Instants type
					if id, found, err := db.GetBlob(hop, AddedAD, pathStatus.Path); err != nil {
						return err
					} else if found {
						db.UpdateBlobHops(id, DeletedAD, entry.Commit)
					}
				}
			}

			if pathStatus.Status == AddedAMD || pathStatus.Status == ModifiedAMD {
				if err := db.InsertBlob(Blob{path: pathStatus.Path, added: []string{entry.Commit}, deleted: []string{}}); err != nil {
					return err
				}
			}
		}

		tipCommit = entry.Commit
		tipHeight += 1

		db.InsertCommit(tipCommit, tipHeight, hops[r])
	}

	return nil
}

func getHops(db DB, givenCommit string) ([]string, error) {
	current := givenCommit
	spine := []string{current}

	for {
		if commit, _, present, err := db.GetCommit(current); err != nil {
			return nil, err
		} else if !present {
			break
		} else {
			if current == NULL {
				break
			}
			current = commit
			spine = append(spine, current)
		}
	}

	return spine, nil
}

// Ruler sequence
//
// input : 0, 1, 2, 3, 4, 5, 6, 7, 8, ...
// output: 0, 0, 1, 0, 2, 0, 1, 0, 3, ...
//
// https://oeis.org/A007814
func ruler(n int) int {
	if n == 0 {
		return 0
	}
	if n%2 != 0 {
		return 0
	}
	return 1 + ruler(n/2)
}

func Search(db DB, commit string) ([]Blob, error) {
	hops, err := getHops(db, commit)
	if err != nil {
		return nil, err
	}

	return db.Search(hops)
}
