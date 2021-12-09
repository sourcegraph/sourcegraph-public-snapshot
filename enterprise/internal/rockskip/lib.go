package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

type Git interface {
	LogReverse(repo string, commit string, n int) ([]LogEntry, error)
	RevList(repo string, commit string) ([]string, error)
}

type DB interface {
	GetCommit(givenCommit string) (commit string, height int, present bool, err error)
	InsertCommit(commit string, height int, ancestor string) error
	GetBlob(hop string, path string) (id int, found bool, err error)
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

func invertAMD(status StatusAMD) StatusAMD {
	switch status {
	case AddedAMD:
		return DeletedAMD
	case ModifiedAMD:
		return DeletedAMD
	case DeletedAMD:
		return AddedAMD
	default:
		fmt.Println("invertAMD: invalid status", status)
		return DeletedAMD
	}
}

func statusAMDToString(status StatusAMD) string {
	switch status {
	case AddedAMD:
		return "A"
	case ModifiedAMD:
		return "M"
	case DeletedAMD:
		return "D"
	default:
		fmt.Println("statusAMDToString: invalid status", status)
		return "?"
	}
}

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
					if id, found, err := db.GetBlob(hop, pathStatus.Path); err != nil {
						return err
					} else if found {
						db.UpdateBlobHops(id, DeletedAD, entry.Commit)
						break
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

		if tipCommit == hops[r] {
			fmt.Println(tipCommit, r, hops, tipHeight)
		}
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

type SubprocessGit struct{}

func NewSubprocessGit() SubprocessGit {
	return SubprocessGit{}
}

func (git SubprocessGit) LogReverse(repo string, givenCommit string, n int) (logEntries []LogEntry, returnError error) {
	log := exec.Command("git",
		"log",
		"--pretty=%H %P",
		"--raw",
		"-z",
		"-m",
		// --no-abbrev speeds up git log a lot
		"--no-abbrev",
		"--no-renames",
		"--first-parent",
		"--reverse",
		"--ignore-submodules",
		fmt.Sprintf("-%d", n),
		givenCommit,
	)
	log.Dir = "/Users/chrismwendt/" + repo
	output, err := log.StdoutPipe()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(output)

	err = log.Start()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = log.Wait()
		if err != nil {
			returnError = err
		}
	}()

	var buf []byte

	for {
		// abc... ... NULL '\n'?

		// Read the commit
		commitBytes, err := reader.Peek(40)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		commit := string(commitBytes)

		// Skip past the NULL byte
		_, err = reader.ReadBytes(0)
		if err != nil {
			return nil, err
		}

		// A '\n' indicates a list of paths and their statuses is next
		buf, err = reader.Peek(1)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if buf[0] == '\n' {
			// A list of paths and their statuses is next

			// Skip the '\n'
			discarded, err := reader.Discard(1)
			if discarded != 1 {
				return nil, fmt.Errorf("discarded %d bytes, expected 1", discarded)
			} else if err != nil {
				return nil, err
			}

			pathStatuses := []PathStatus{}
			for {
				// :100644 100644 abc... def... M NULL file.txt NULL
				// ^ 0                          ^ 97   ^ 99

				// A ':' indicates a path and its status is next
				buf, err = reader.Peek(1)
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
				if buf[0] != ':' {
					break
				}

				// Read the status from index 97 and skip to the path at index 99
				buf = make([]byte, 99)
				read, err := reader.Read(buf)
				if read != 99 {
					return nil, fmt.Errorf("read %d bytes, expected 99", read)
				} else if err != nil {
					return nil, err
				}
				var status StatusAMD
				switch buf[97] {
				case 'A':
					status = AddedAMD
				case 'M':
					status = ModifiedAMD
				case 'D':
					status = DeletedAMD
				default:
					// Ignore other statuses (e.g. 'T' for type changed)
					continue
				}

				// Read the path
				path, err := reader.ReadBytes(0)
				if err != nil {
					return nil, err
				}
				path = path[:len(path)-1] // Drop the trailing NULL byte

				pathStatuses = append(pathStatuses, PathStatus{Path: string(path), Status: status})
			}

			logEntries = append(logEntries, LogEntry{Commit: commit, PathStatuses: pathStatuses})
		}
	}

	return logEntries, nil
}

func (git SubprocessGit) RevList(repo string, givenCommit string) (commits []string, returnError error) {
	revList := exec.Command("git", "rev-list", givenCommit)
	revList.Dir = "/Users/chrismwendt/" + repo
	output, err := revList.StdoutPipe()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(output)

	err = revList.Start()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = revList.Wait()
		if err != nil {
			returnError = err
		}
	}()

	for {
		commit, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}
