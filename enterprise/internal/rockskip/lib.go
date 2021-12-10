package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os/exec"

	pg "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Git interface {
	LogReverse(repo string, commit string, n int) ([]LogEntry, error)
	RevList(repo string, commit string) ([]string, error)
	CatFile(repo string, commit string, path string) ([]byte, error)
}

type DB interface {
	GetCommit(givenCommit string) (ancestor string, height int, present bool, err error)
	InsertCommit(commit string, height int, ancestor string) error
	GetBlob(hop string, path string) (id int, found bool, err error)
	UpdateBlobHops(id int, status StatusAD, hop string) error
	InsertBlob(blob Blob) error
	AppendHop(hops []string, status StatusAD, hop string) error
	Search(hops []string) ([]Blob, error)
}

type ParseSymbolsFunc func(path string, bytes []byte) (symbols []string, err error)

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
	commit  string
	path    string
	added   []string
	deleted []string
	symbols []string
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

func Index(git Git, db DB, parse ParseSymbolsFunc, repo string, givenCommit string) error {
	tipCommit := NULL
	tipHeight := 0
	missingCount := 0
	revs, err := git.RevList(repo, givenCommit)
	if err != nil {
		return err
	}
	for _, commit := range revs {
		if _, height, present, err := db.GetCommit(commit); err != nil {
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

		err = db.AppendHop(hops[0:r], AddedAD, entry.Commit)
		if err != nil {
			return err
		}
		err = db.AppendHop(hops[0:r], DeletedAD, entry.Commit)
		if err != nil {
			return err
		}

		// Could delete redundant hops here.

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
				contents, err := git.CatFile(repo, entry.Commit, pathStatus.Path)
				if err != nil {
					return err
				}
				symbols, err := parse(pathStatus.Path, contents)
				if err != nil {
					return err
				}
				blob := Blob{
					commit:  entry.Commit,
					path:    pathStatus.Path,
					added:   []string{entry.Commit},
					deleted: []string{},
					symbols: symbols,
				}
				if err := db.InsertBlob(blob); err != nil {
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

func getHops(db DB, commit string) ([]string, error) {
	current := commit
	spine := []string{current}

	for {
		if ancestor, _, present, err := db.GetCommit(current); err != nil {
			return nil, err
		} else if !present {
			break
		} else {
			if current == NULL {
				break
			}
			current = ancestor
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
				read, err := io.ReadFull(reader, buf)
				if read != 99 {
					// TODO figure out this 🐛
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

func (git SubprocessGit) CatFile(repo string, commit string, path string) ([]byte, error) {
	cmd := exec.Command("git", "cat-file", "blob", fmt.Sprintf("%s:%s", commit, path))
	cmd.Dir = "/Users/chrismwendt/" + repo
	return cmd.Output()
}

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB() (*PostgresDB, error) {
	db, err := sql.Open("postgres", "postgres://sourcegraph:sourcegraph@localhost:5432/sourcegraph?sslmode=disable")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("DROP TABLE IF EXISTS rockskip_ancestry")
	if err != nil {
		return nil, fmt.Errorf("dropping rockskip_ancestry: %s", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS rockskip_blobs")
	if err != nil {
		return nil, fmt.Errorf("dropping rockskip_blobs: %s", err)
	}

	_, err = db.Exec(`
		CREATE TABLE rockskip_ancestry (
			commit_id   VARCHAR(40) PRIMARY KEY,
			height      INTEGER     NOT NULL,
			ancestor_id VARCHAR(40) NOT NULL
		)`)
	if err != nil {
		return nil, fmt.Errorf("creating rockskip_ancestry: %s", err)
	}

	_, err = db.Exec(`
		CREATE TABLE rockskip_blobs (
			id      SERIAL        PRIMARY KEY,
			commit  VARCHAR(40)   NOT NULL,
			path    BYTEA         NOT NULL,
			added   VARCHAR(40)[] NOT NULL,
			deleted VARCHAR(40)[] NOT NULL,
			symbols TEXT[]		  NOT NULL
		)`)
	if err != nil {
		return nil, fmt.Errorf("creating rockskip_blobs: %s", err)
	}

	fmt.Println("TODO add indexes")
	fmt.Println("TODO use transactions")

	return &PostgresDB{db: db}, nil
}

func (db PostgresDB) GetCommit(givenCommit string) (ancestor string, height int, present bool, err error) {
	err = db.db.QueryRow(`
		SELECT ancestor_id, height
		FROM rockskip_ancestry
		WHERE commit_id = $1
	`, givenCommit).Scan(&ancestor, &height)
	if err == sql.ErrNoRows {
		return "", 0, false, nil
	} else if err != nil {
		return "", 0, false, fmt.Errorf("GetCommit: %s", err)
	}
	return ancestor, height, true, nil
}

func (db PostgresDB) InsertCommit(commit string, height int, ancestor string) error {
	_, err := db.db.Exec(`
		INSERT INTO rockskip_ancestry (commit_id, height, ancestor_id)
		VALUES ($1, $2, $3)
	`, commit, height, ancestor)
	return errors.Wrap(err, "InsertCommit")
}

func (db PostgresDB) GetBlob(hop string, path string) (id int, found bool, err error) {
	err = db.db.QueryRow(`
		SELECT id
		FROM rockskip_blobs
		WHERE path = $1 AND $2 = ANY (added) AND NOT $2 = ANY (deleted)
	`, path, hop).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	} else if err != nil {
		return 0, false, fmt.Errorf("GetBlob: %s", err)
	}
	return id, true, nil
}

func (db PostgresDB) UpdateBlobHops(id int, status StatusAD, hop string) error {
	column := statusADToColumn(status)
	// TODO also try `||` instead of `array_append``
	_, err := db.db.Exec(fmt.Sprintf(`
		UPDATE rockskip_blobs
		SET %s = array_append(%s, $1)
		WHERE id = $2
	`, column, column), hop, id)
	return errors.Wrap(err, "UpdateBlobHops")
}

func (db PostgresDB) InsertBlob(blob Blob) error {
	_, err := db.db.Exec(`
		INSERT INTO rockskip_blobs (commit, path, added, deleted, symbols)
		VALUES ($1, $2, $3, $4, $5)
	`, blob.commit, blob.path, pg.Array(blob.added), pg.Array(blob.deleted), pg.Array(blob.symbols))
	return errors.Wrap(err, "InsertBlob")
}

func (db PostgresDB) AppendHop(hops []string, givenStatus StatusAD, newHop string) error {
	column := statusADToColumn(givenStatus)
	_, err := db.db.Exec(fmt.Sprintf(`
		UPDATE rockskip_blobs
		SET %s = array_append(%s, $1)
		WHERE $2 && %s
	`, column, column, column), newHop, pg.Array(hops))
	return errors.Wrap(err, "AppendHop")
}

func (db PostgresDB) Search(hops []string) ([]Blob, error) {
	rows, err := db.db.Query(`
		SELECT id, commit, path, added, deleted, symbols
		FROM rockskip_blobs
		WHERE $1 && added AND NOT $1 && deleted
	`, pg.Array(hops))
	if err != nil {
		return nil, errors.Wrap(err, "Search")
	}
	defer rows.Close()

	blobs := []Blob{}
	for rows.Next() {
		var id int
		var commit string
		var path string
		var added, deleted []string
		var symbols []string
		err = rows.Scan(&id, &commit, &path, pg.Array(&added), pg.Array(&deleted), pg.Array(&symbols))
		if err != nil {
			return nil, errors.Wrap(err, "Search: Scan")
		}
		blobs = append(blobs, Blob{commit: commit, path: path, added: added, deleted: deleted, symbols: symbols})
	}
	return blobs, nil
}

func (db PostgresDB) PrintInternals() error {
	fmt.Println("Commit ancestry:")
	fmt.Println()

	// print all rows in the rockskip_ancestry table
	rows, err := db.db.Query(`
		SELECT commit_id, height, ancestor_id
		FROM rockskip_ancestry
		ORDER BY height ASC
	`)
	if err != nil {
		return errors.Wrap(err, "PrintInternals")
	}
	defer rows.Close()

	for rows.Next() {
		var commit, ancestor string
		var height int
		err = rows.Scan(&commit, &height, &ancestor)
		if err != nil {
			return errors.Wrap(err, "PrintInternals: Scan")
		}
		fmt.Printf("height %3d commit %s ancestor %s\n", height, commit, ancestor)
	}

	fmt.Println()
	fmt.Println("Blobs:")
	fmt.Println()

	rows, err = db.db.Query(`
		SELECT id, path, added, deleted
		FROM rockskip_blobs
		ORDER BY id ASC
	`)
	if err != nil {
		return errors.Wrap(err, "PrintInternals")
	}

	for rows.Next() {
		var id int
		var path string
		var added, deleted []string
		err = rows.Scan(&id, &path, pg.Array(&added), pg.Array(&deleted))
		if err != nil {
			return errors.Wrap(err, "PrintInternals: Scan")
		}
		fmt.Printf("  id %d path %-10s\n", id, path)
		for _, a := range added {
			fmt.Printf("    + %-40s\n", a)
		}
		fmt.Println()
		for _, d := range deleted {
			fmt.Printf("    - %-40s\n", d)
		}
		fmt.Println()

	}

	fmt.Println()
	return nil
}

func statusADToColumn(status StatusAD) string {
	switch status {
	case AddedAD:
		return "added"
	case DeletedAD:
		return "deleted"
	default:
		fmt.Println("unexpected status StatusAD: ", status)
		return "unknown_status"
	}
}
