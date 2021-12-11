package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	pg "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Git interface {
	LogReverse(commit string, n int) ([]LogEntry, error)
	RevList(commit string) ([]string, error)
	CatFile(commit string, path string) ([]byte, error)
}

type DB interface {
	GetCommit(givenCommit string) (ancestor string, height int, present bool, err error)
	InsertCommit(commit string, height int, ancestor string) error
	GetBlob(hop string, path string) (id int, found bool, err error)
	UpdateBlobHops(id int, status StatusAD, hop string) error
	InsertBlob(blob Blob) (id int, err error)
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

type Instants struct {
	name       string
	start      time.Time
	nameToTask map[string]*Task
}

type Task struct {
	Duration time.Duration
	Count    int
}

func NewInstants() Instants {
	return Instants{
		name:       "<start>",
		start:      time.Now(),
		nameToTask: map[string]*Task{"<start>": &Task{Duration: 0, Count: 0}},
	}
}

func (instants *Instants) Start(name string) {
	now := time.Now()

	if _, ok := instants.nameToTask[instants.name]; !ok {
		instants.nameToTask[instants.name] = &Task{Duration: 0, Count: 0}
	}
	instants.nameToTask[instants.name].Duration += now.Sub(instants.start)
	instants.nameToTask[instants.name].Count += 1

	instants.name = name
	instants.start = now
}

func (instants *Instants) Reset() {
	instants.name = "<start>"
	instants.start = time.Now()
	instants.nameToTask = map[string]*Task{"<start>": &Task{Duration: 0, Count: 0}}
}

func (instants Instants) Print() {
	instants.Start(instants.name)

	ms := func(d time.Duration) string {
		return fmt.Sprintf("%dms", int(d.Seconds()*1000))
	}

	var total time.Duration = 0
	totalCount := 0
	for _, task := range instants.nameToTask {
		total += task.Duration
		totalCount += task.Count
	}
	fmt.Printf("Instants (%s total):\n", ms(total))

	type kv struct {
		Key   string
		Value *Task
	}

	var kvs []kv
	for k, v := range instants.nameToTask {
		kvs = append(kvs, kv{k, v})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value.Duration > kvs[j].Value.Duration
	})

	for _, kv := range kvs {
		fmt.Printf("  [%6s] x%d %s\n", ms(kv.Value.Duration), kv.Value.Count, kv.Key)
	}
}

var INSTANTS = NewInstants()

func Index(git Git, db DB, parse ParseSymbolsFunc, givenCommit string) error {
	tipCommit := NULL
	tipHeight := 0
	missingCount := 0
	INSTANTS.Start("RevList")
	revs, err := git.RevList(givenCommit)
	INSTANTS.Start("idle")
	if err != nil {
		return err
	}
	for _, commit := range revs {
		INSTANTS.Start("GetCommit")
		_, height, present, err := db.GetCommit(commit)
		INSTANTS.Start("idle")
		if err != nil {
			return err
		} else if present {
			tipCommit = commit
			tipHeight = height
			break
		}
		missingCount += 1
	}

	pathToBlobIdCache := map[string]int{}

	INSTANTS.Start("LogReverse")
	entries, err := git.LogReverse(givenCommit, missingCount)
	INSTANTS.Start("idle")
	if err != nil {
		return err
	}
	start := time.Now()
	last := time.Now()
	for entryIndex, entry := range entries {
		if time.Since(last) > time.Second {
			fmt.Printf("Index: height %d (%d/s)\n", tipHeight+1, entryIndex/int(time.Since(start).Seconds()))
			last = time.Now()
		}
		hops, err := getHops(db, tipCommit)
		if err != nil {
			return err
		}

		r := ruler(tipHeight + 1)
		if r >= len(hops) {
			return fmt.Errorf("ruler(%d) = %d is out of range of len(hops) = %d", tipHeight+1, r, len(hops))
		}

		INSTANTS.Start("AppendHop (added)")
		err = db.AppendHop(hops[0:r], AddedAD, entry.Commit)
		INSTANTS.Start("idle")
		if err != nil {
			return err
		}
		INSTANTS.Start("AppendHop (deleted)")
		err = db.AppendHop(hops[0:r], DeletedAD, entry.Commit)
		INSTANTS.Start("idle")
		if err != nil {
			return err
		}

		// Could delete redundant hops here.

		for _, pathStatus := range entry.PathStatuses {
			if pathStatus.Status == DeletedAMD || pathStatus.Status == ModifiedAMD {
				id := 0

				ok := false
				if id, ok = pathToBlobIdCache[pathStatus.Path]; !ok {
					found := false
					for _, hop := range hops {
						INSTANTS.Start("GetBlob")
						id, found, err = db.GetBlob(hop, pathStatus.Path)
						INSTANTS.Start("idle")
						if err != nil {
							return err
						}
						if found {
							break
						}
					}
					if !found {
						return fmt.Errorf("could not find blob for path %s", pathStatus.Path)
					}
				}

				INSTANTS.Start("UpdateBlobHops")
				db.UpdateBlobHops(id, DeletedAD, entry.Commit)
				INSTANTS.Start("idle")
			}

			if pathStatus.Status == AddedAMD || pathStatus.Status == ModifiedAMD {
				INSTANTS.Start("CatFile")
				contents, err := git.CatFile(entry.Commit, pathStatus.Path)
				INSTANTS.Start("idle")
				if err != nil {
					return err
				}
				INSTANTS.Start("parse")
				symbols, err := parse(pathStatus.Path, contents)
				INSTANTS.Start("idle")
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
				INSTANTS.Start("InsertBlob")
				id, err := db.InsertBlob(blob)
				INSTANTS.Start("idle")
				if err != nil {
					return err
				}
				pathToBlobIdCache[pathStatus.Path] = id
			}
		}

		tipCommit = entry.Commit
		tipHeight += 1

		if tipCommit == hops[r] {
			fmt.Println(tipCommit, r, hops, tipHeight)
		}
		INSTANTS.Start("InsertCommit")
		db.InsertCommit(tipCommit, tipHeight, hops[r])
		INSTANTS.Start("idle")
	}

	return nil
}

func getHops(db DB, commit string) ([]string, error) {
	current := commit
	spine := []string{current}

	for {
		INSTANTS.Start("GetCommit")
		ancestor, _, present, err := db.GetCommit(current)
		INSTANTS.Start("idle")
		if err != nil {
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

type SubprocessGit struct {
	repo          string
	catFileCmd    *exec.Cmd
	catFileStdin  io.WriteCloser
	catFileStdout bufio.Reader
}

func NewSubprocessGit(repo string) (*SubprocessGit, error) {
	cmd := exec.Command("git", "cat-file", "--batch")
	cmd.Dir = "/Users/chrismwendt/" + repo

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &SubprocessGit{
		repo:          repo,
		catFileCmd:    cmd,
		catFileStdin:  stdin,
		catFileStdout: *bufio.NewReader(stdout),
	}, nil
}

func (git SubprocessGit) Close() error {
	err := git.catFileStdin.Close()
	if err != nil {
		return err
	}
	return git.catFileCmd.Wait()
}

func (git SubprocessGit) LogReverse(givenCommit string, n int) (logEntries []LogEntry, returnError error) {
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
	log.Dir = "/Users/chrismwendt/" + git.repo
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
					return nil, fmt.Errorf("read %d bytes, expected 99", read)
				} else if err != nil {
					return nil, err
				}

				// Read the path
				path, err := reader.ReadBytes(0)
				if err != nil {
					return nil, err
				}
				path = path[:len(path)-1] // Drop the trailing NULL byte

				// Inspect the status
				var status StatusAMD
				statusByte := buf[97]
				switch statusByte {
				case 'A':
					status = AddedAMD
				case 'M':
					status = ModifiedAMD
				case 'D':
					status = DeletedAMD
				case 'T':
					// Type changed. Check if it changed from a file to a submodule or vice versa,
					// treating submodules as empty.

					isSubmodule := func(mode string) bool {
						// Submodules are mode "160000". https://stackoverflow.com/questions/737673/how-to-read-the-mode-field-of-git-ls-trees-output#comment3519596_737877
						return mode == "160000"
					}

					oldMode := string(buf[1:7])
					newMode := string(buf[8:14])

					if isSubmodule(oldMode) && !isSubmodule(newMode) {
						// It changed from a submodule to a file, so consider it added.
						status = AddedAMD
						break
					}

					if !isSubmodule(oldMode) && isSubmodule(newMode) {
						// It changed from a file to a submodule, so consider it deleted.
						status = DeletedAMD
						break
					}

					// Otherwise, it remained the same, so ignore the type change.
				case 'C':
					// Copied
					return nil, fmt.Errorf("unexpected status 'C' given --no-renames was specified")
				case 'R':
					// Renamed
					return nil, fmt.Errorf("unexpected status 'R' given --no-renames was specified")
				case 'X':
					return nil, fmt.Errorf("unexpected status 'X' indicates a bug in git")
				default:
					fmt.Printf("LogReverse repo %q commit %q path %q: unrecognized diff status %q, skipping\n", git.repo, commit, path, string(statusByte))
					continue
				}

				pathStatuses = append(pathStatuses, PathStatus{Path: string(path), Status: status})
			}

			logEntries = append(logEntries, LogEntry{Commit: commit, PathStatuses: pathStatuses})
		}
	}

	return logEntries, nil
}

func (git SubprocessGit) RevList(givenCommit string) (commits []string, returnError error) {
	revList := exec.Command("git", "rev-list", givenCommit)
	revList.Dir = "/Users/chrismwendt/" + git.repo
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

func (git SubprocessGit) CatFile(commit string, path string) ([]byte, error) {
	_, err := git.catFileStdin.Write([]byte(fmt.Sprintf("%s:%s\n", commit, path)))
	if err != nil {
		return nil, err
	}

	line, err := git.catFileStdout.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = line[:len(line)-1] // Drop the trailing newline
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected cat-file output: %q", line)
	}
	size, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, err
	}

	fileContents, err := io.ReadAll(io.LimitReader(&git.catFileStdout, size))
	if err != nil {
		return nil, err
	}

	discarded, err := git.catFileStdout.Discard(1) // Discard the trailing newline
	if err != nil {
		return nil, err
	}
	if discarded != 1 {
		return nil, fmt.Errorf("expected to discard 1 byte, but discarded %d", discarded)
	}

	return fileContents, nil
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
			path    TEXT          NOT NULL,
			added   VARCHAR(40)[] NOT NULL,
			deleted VARCHAR(40)[] NOT NULL,
			symbols TEXT[]		  NOT NULL
		)`)
	if err != nil {
		return nil, fmt.Errorf("creating rockskip_blobs: %s", err)
	}

	_, err = db.Exec("CREATE INDEX rockskip_ancestry_commit_id ON rockskip_ancestry (commit_id)")
	if err != nil {
		return nil, fmt.Errorf("creating index rockskip_ancestry_commit_id: %s", err)
	}

	_, err = db.Exec("CREATE INDEX rockskip_blobs_path ON rockskip_blobs(path)")
	if err != nil {
		return nil, fmt.Errorf("creating index rockskip_blobs_path: %s", err)
	}

	_, err = db.Exec("CREATE INDEX rockskip_blobs_added ON rockskip_blobs USING GIN (added)")
	if err != nil {
		return nil, fmt.Errorf("creating index rockskip_blobs_added: %s", err)
	}

	_, err = db.Exec("CREATE INDEX rockskip_blobs_deleted ON rockskip_blobs USING GIN (deleted)")
	if err != nil {
		return nil, fmt.Errorf("creating index rockskip_blobs_deleted: %s", err)
	}

	fmt.Println("TODO add more indexes")
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

func (db PostgresDB) InsertBlob(blob Blob) (id int, err error) {
	lastInsertId := 0
	err = db.db.QueryRow(`
		INSERT INTO rockskip_blobs (commit, path, added, deleted, symbols)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, blob.commit, blob.path, pg.Array(blob.added), pg.Array(blob.deleted), pg.Array(blob.symbols)).Scan(&lastInsertId)
	return lastInsertId, errors.Wrap(err, "InsertBlob")
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

func (db PostgresDB) Close() error {
	return db.db.Close()
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
