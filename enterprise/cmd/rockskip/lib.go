package rockskip

import (
	"bufio"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
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

type Symbol struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`
	Kind   string `json:"kind"`
	Line   int    `json:"line"`
}

type Symbols []Symbol

func (symbols Symbols) Value() (driver.Value, error) {
	return json.Marshal(symbols)
}

func (symbols *Symbols) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("scan symbol: expected []byte")
	}

	return json.Unmarshal(bytes, &symbols)
}

type ParseSymbolsFunc func(path string, bytes []byte) (symbols []Symbol, err error)

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
	Commit  string
	Path    string
	Added   []string
	Deleted []string
	Symbols []Symbol
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

type TaskLog struct {
	currentName  string
	currentStart time.Time
	nameToTask   map[string]*Task
}

type Task struct {
	Duration time.Duration
	Count    int
}

func NewTaskLog() TaskLog {
	return TaskLog{
		currentName:  "<start>",
		currentStart: time.Now(),
		nameToTask:   map[string]*Task{"<start>": {Duration: 0, Count: 0}},
	}
}

func (t *TaskLog) Start(name string) {
	now := time.Now()

	if _, ok := t.nameToTask[t.currentName]; !ok {
		t.nameToTask[t.currentName] = &Task{Duration: 0, Count: 0}
	}
	t.nameToTask[t.currentName].Duration += now.Sub(t.currentStart)
	t.nameToTask[t.currentName].Count += 1

	t.currentName = name
	t.currentStart = now
}

func (t *TaskLog) Reset() {
	t.currentName = "<start>"
	t.currentStart = time.Now()
	t.nameToTask = map[string]*Task{"<start>": {Duration: 0, Count: 0}}
}

func (t TaskLog) Print() {
	t.Start(t.currentName)

	ms := func(d time.Duration) string {
		return fmt.Sprintf("%dms", int(d.Seconds()*1000))
	}

	var total time.Duration = 0
	totalCount := 0
	for _, task := range t.nameToTask {
		total += task.Duration
		totalCount += task.Count
	}
	fmt.Printf("Instants (%s total):\n", ms(total))

	type kv struct {
		Key   string
		Value *Task
	}

	var kvs []kv
	for k, v := range t.nameToTask {
		kvs = append(kvs, kv{k, v})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value.Duration > kvs[j].Value.Duration
	})

	for _, kv := range kvs {
		fmt.Printf("  [%6s] %6dx %s\n", ms(kv.Value.Duration), kv.Value.Count, kv.Key)
	}
}

var TASKLOG = NewTaskLog()

func Index(git Git, db *sql.DB, parse ParseSymbolsFunc, givenCommit string) error {
	tipCommit := NULL
	tipHeight := 0
	missingCount := 0
	TASKLOG.Start("RevList")
	revs, err := git.RevList(givenCommit)
	TASKLOG.Start("idle")
	if err != nil {
		return errors.Wrap(err, "RevList")
	}
	for _, commit := range revs {
		TASKLOG.Start("GetCommit")
		_, height, present, err := GetCommit(db, commit)
		TASKLOG.Start("idle")
		if err != nil {
			return errors.Wrap(err, "GetCommit")
		} else if present {
			tipCommit = commit
			tipHeight = height
			break
		}
		missingCount += 1
	}

	pathToBlobIdCache := map[string]int{}

	TASKLOG.Start("LogReverse")
	entries, err := git.LogReverse(givenCommit, missingCount)
	TASKLOG.Start("idle")
	if err != nil {
		return errors.Wrap(err, "LogReverse")
	}
	start := time.Now()
	last := time.Now()
	for entryIndex, entry := range entries {
		tx, err := db.Begin()
		if err != nil {
			return errors.Wrap(err, "begin transaction")
		}
		defer tx.Rollback()

		if time.Since(last) > time.Second {
			fmt.Printf("Index: height %d (%d/s)\n", tipHeight+1, entryIndex/int(time.Since(start).Seconds()))
			last = time.Now()
		}
		hops, err := getHops(tx, tipCommit)
		if err != nil {
			return errors.Wrap(err, "getHops")
		}

		r := ruler(tipHeight + 1)
		if r >= len(hops) {
			return fmt.Errorf("ruler(%d) = %d is out of range of len(hops) = %d", tipHeight+1, r, len(hops))
		}

		TASKLOG.Start("AppendHop (added)")
		err = AppendHop(tx, hops[0:r], AddedAD, entry.Commit)
		TASKLOG.Start("idle")
		if err != nil {
			return errors.Wrap(err, "AppendHop (added)")
		}
		TASKLOG.Start("AppendHop (deleted)")
		err = AppendHop(tx, hops[0:r], DeletedAD, entry.Commit)
		TASKLOG.Start("idle")
		if err != nil {
			return errors.Wrap(err, "AppendHop (deleted)")
		}

		for _, pathStatus := range entry.PathStatuses {
			if pathStatus.Status == DeletedAMD || pathStatus.Status == ModifiedAMD {
				id := 0

				ok := false
				if id, ok = pathToBlobIdCache[pathStatus.Path]; !ok {
					found := false
					for _, hop := range hops {
						TASKLOG.Start("GetBlob")
						id, found, err = GetBlob(tx, hop, pathStatus.Path)
						TASKLOG.Start("idle")
						if err != nil {
							return errors.Wrap(err, "GetBlob")
						}
						if found {
							break
						}
					}
					if !found {
						return fmt.Errorf("could not find blob for path %s", pathStatus.Path)
					}
				}

				TASKLOG.Start("UpdateBlobHops")
				UpdateBlobHops(tx, id, DeletedAD, entry.Commit)
				TASKLOG.Start("idle")
			}

			if pathStatus.Status == AddedAMD || pathStatus.Status == ModifiedAMD {
				TASKLOG.Start("CatFile")
				contents, err := git.CatFile(entry.Commit, pathStatus.Path)
				TASKLOG.Start("idle")
				if err != nil {
					return errors.Wrap(err, "CatFile")
				}
				TASKLOG.Start("parse")
				symbols, err := parse(pathStatus.Path, contents)
				TASKLOG.Start("idle")
				if err != nil {
					return err
				}
				blob := Blob{
					Commit:  entry.Commit,
					Path:    pathStatus.Path,
					Added:   []string{entry.Commit},
					Deleted: []string{},
					Symbols: symbols,
				}
				TASKLOG.Start("InsertBlob")
				id, err := InsertBlob(tx, blob)
				TASKLOG.Start("idle")
				if err != nil {
					return errors.Wrap(err, "InsertBlob")
				}
				pathToBlobIdCache[pathStatus.Path] = id
			}
		}

		TASKLOG.Start("DeleteRedundant")
		err = DeleteRedundant(tx, entry.Commit)
		TASKLOG.Start("idle")
		if err != nil {
			return errors.Wrap(err, "DeleteRedundant")
		}

		tipCommit = entry.Commit
		tipHeight += 1

		TASKLOG.Start("InsertCommit")
		InsertCommit(tx, tipCommit, tipHeight, hops[r])
		TASKLOG.Start("CommitTx")
		err = tx.Commit()
		if err != nil {
			return errors.Wrap(err, "commit transaction")
		}
		TASKLOG.Start("idle")
	}

	return nil
}

func getHops(tx Queryable, commit string) ([]string, error) {
	current := commit
	spine := []string{current}

	for {
		TASKLOG.Start("GetCommit")
		ancestor, _, present, err := GetCommit(tx, current)
		TASKLOG.Start("idle")
		if err != nil {
			return nil, errors.Wrap(err, "GetCommit")
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
	log := exec.Command("git", LogReverseArgs(n, givenCommit)...)
	log.Dir = "/Users/chrismwendt/" + git.repo
	output, err := log.StdoutPipe()
	if err != nil {
		return nil, err
	}

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

	return ParseLogReverse(output)
}

func (git SubprocessGit) RevList(givenCommit string) (commits []string, returnError error) {
	revList := exec.Command("git", RevListArgs(givenCommit)...)
	revList.Dir = "/Users/chrismwendt/" + git.repo
	output, err := revList.StdoutPipe()
	if err != nil {
		return nil, err
	}

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

	return ParseRevList(output)
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

func GetCommit(db Queryable, givenCommit string) (ancestor string, height int, present bool, err error) {
	err = db.QueryRow(`
		SELECT ancestor, height
		FROM rockskip_ancestry
		WHERE id = $1
	`, givenCommit).Scan(&ancestor, &height)
	if err == sql.ErrNoRows {
		return "", 0, false, nil
	} else if err != nil {
		return "", 0, false, fmt.Errorf("GetCommit: %s", err)
	}
	return ancestor, height, true, nil
}

func InsertCommit(db Queryable, commit string, height int, ancestor string) error {
	_, err := db.Exec(`
		INSERT INTO rockskip_ancestry (id, height, ancestor)
		VALUES ($1, $2, $3)
	`, commit, height, ancestor)
	return errors.Wrap(err, "InsertCommit")
}

func GetBlob(db Queryable, hop string, path string) (id int, found bool, err error) {
	err = db.QueryRow(`
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

func UpdateBlobHops(db Queryable, id int, status StatusAD, hop string) error {
	column := statusADToColumn(status)
	// TODO also try `||` instead of `array_append``
	_, err := db.Exec(fmt.Sprintf(`
		UPDATE rockskip_blobs
		SET %s = array_append(%s, $1)
		WHERE id = $2
	`, column, column), hop, id)
	return errors.Wrap(err, "UpdateBlobHops")
}

func InsertBlob(db Queryable, blob Blob) (id int, err error) {
	symbolNames := []string{}
	for _, symbol := range blob.Symbols {
		symbolNames = append(symbolNames, symbol.Name)
	}

	lastInsertId := 0
	err = db.QueryRow(`
		INSERT INTO rockskip_blobs (commit, path, added, deleted, symbol_names, symbol_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, blob.Commit, blob.Path, pg.Array(blob.Added), pg.Array(blob.Deleted), pg.Array(symbolNames), Symbols(blob.Symbols)).Scan(&lastInsertId)
	return lastInsertId, errors.Wrap(err, "InsertBlob")
}

func AppendHop(db Queryable, hops []string, givenStatus StatusAD, newHop string) error {
	column := statusADToColumn(givenStatus)
	_, err := db.Exec(fmt.Sprintf(`
		UPDATE rockskip_blobs
		SET %s = array_append(%s, $1)
		WHERE $2 && %s
	`, column, column, column), newHop, pg.Array(hops))
	return errors.Wrap(err, "AppendHop")
}

func Search(db Queryable, commit string, query *string) ([]Blob, error) {
	var err error

	hops, err := getHops(db, commit)
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows
	if query != nil {
		rows, err = db.Query(`
			SELECT id, commit, path, added, deleted, symbol_data
			FROM rockskip_blobs
			WHERE
				$1 && added
				AND NOT $1 && deleted
				AND $2 && symbol_names
		`, pg.Array(hops), pg.Array([]string{*query}))
	} else {
		rows, err = db.Query(`
			SELECT id, commit, path, added, deleted, symbol_data
			FROM rockskip_blobs
			WHERE
				$1 && added
				AND NOT $1 && deleted
		`, pg.Array(hops))
	}
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
		var symbols Symbols
		err = rows.Scan(&id, &commit, &path, pg.Array(&added), pg.Array(&deleted), &symbols)
		if err != nil {
			return nil, errors.Wrap(err, "Search: Scan")
		}
		blobs = append(blobs, Blob{Commit: commit, Path: path, Added: added, Deleted: deleted, Symbols: symbols})
	}
	return blobs, nil
}

func DeleteRedundant(db Queryable, hop string) error {
	_, err := db.Exec(`
		UPDATE rockskip_blobs
		SET added = array_remove(added, $1), deleted = array_remove(deleted, $1)
		WHERE $2 && added AND $2 && deleted
	`, hop, pg.Array([]string{hop}))
	return errors.Wrap(err, "DeleteRedundant")
}

func PrintInternals(db Queryable) error {
	fmt.Println("Commit ancestry:")
	fmt.Println()

	// print all rows in the rockskip_ancestry table
	rows, err := db.Query(`
		SELECT id, height, ancestor
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

	rows, err = db.Query(`
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

func LogReverseArgs(n int, givenCommit string) []string {
	return []string{
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
	}
}

func ParseLogReverse(stdout io.Reader) (logEntries []LogEntry, returnError error) {
	reader := bufio.NewReader(stdout)

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
					continue
				case 'C':
					// Copied
					return nil, fmt.Errorf("unexpected status 'C' given --no-renames was specified")
				case 'R':
					// Renamed
					return nil, fmt.Errorf("unexpected status 'R' given --no-renames was specified")
				case 'X':
					return nil, fmt.Errorf("unexpected status 'X' indicates a bug in git")
				default:
					fmt.Printf("LogReverse commit %q path %q: unrecognized diff status %q, skipping\n", commit, path, string(statusByte))
					continue
				}

				pathStatuses = append(pathStatuses, PathStatus{Path: string(path), Status: status})
			}

			logEntries = append(logEntries, LogEntry{Commit: commit, PathStatuses: pathStatuses})
		}
	}

	return logEntries, nil
}

func RevListArgs(givenCommit string) []string {
	return []string{"rev-list", "--first-parent", givenCommit}
}

func ParseRevList(stdout io.Reader) (commits []string, returnError error) {
	reader := bufio.NewReader(stdout)

	for {
		commit, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		commit = commit[:len(commit)-1] // Drop the trailing newline
		commits = append(commits, commit)
	}

	return commits, nil
}

type Queryable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
