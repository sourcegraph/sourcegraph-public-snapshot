package shared

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IndexJob struct {
	Indexer string
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

// Index is a subset of the lsif_indexes table and stores both processed and unprocessed
// records.
type Index struct {
	ID                 int                 `json:"id"`
	Commit             string              `json:"commit"`
	QueuedAt           time.Time           `json:"queuedAt"`
	State              string              `json:"state"`
	FailureMessage     *string             `json:"failureMessage"`
	StartedAt          *time.Time          `json:"startedAt"`
	FinishedAt         *time.Time          `json:"finishedAt"`
	ProcessAfter       *time.Time          `json:"processAfter"`
	NumResets          int                 `json:"numResets"`
	NumFailures        int                 `json:"numFailures"`
	RepositoryID       int                 `json:"repositoryId"`
	LocalSteps         []string            `json:"local_steps"`
	RepositoryName     string              `json:"repositoryName"`
	DockerSteps        []DockerStep        `json:"docker_steps"`
	Root               string              `json:"root"`
	Indexer            string              `json:"indexer"`
	IndexerArgs        []string            `json:"indexer_args"` // TODO - convert this to `IndexCommand string`
	Outfile            string              `json:"outfile"`
	ExecutionLogs      []ExecutionLogEntry `json:"execution_logs"`
	Rank               *int                `json:"placeInQueue"`
	AssociatedUploadID *int                `json:"associatedUpload"`
}

type DockerStep struct {
	Root     string   `json:"root"`
	Image    string   `json:"image"`
	Commands []string `json:"commands"`
}

func (s *DockerStep) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("value is not []byte: %T", value)
	}

	return json.Unmarshal(b, &s)
}

func (s DockerStep) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// ExecutionLogEntry represents a command run by the executor.
type ExecutionLogEntry struct {
	Key        string    `json:"key"`
	Command    []string  `json:"command"`
	StartTime  time.Time `json:"startTime"`
	ExitCode   *int      `json:"exitCode,omitempty"`
	Out        string    `json:"out,omitempty"`
	DurationMs *int      `json:"durationMs,omitempty"`
}

func (e *ExecutionLogEntry) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("value is not []byte: %T", value)
	}

	return json.Unmarshal(b, &e)
}

func (e ExecutionLogEntry) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func ExecutionLogEntries(raw []workerutil.ExecutionLogEntry) (entries []ExecutionLogEntry) {
	for _, entry := range raw {
		entries = append(entries, ExecutionLogEntry(entry))
	}

	return entries
}

// IndexConfiguration stores the index configuration for a repository.
type IndexConfiguration struct {
	ID           int    `json:"id"`
	RepositoryID int    `json:"repository_id"`
	Data         []byte `json:"data"`
}

type GetIndexesOptions struct {
	RepositoryID int
	State        string
	Term         string
	Limit        int
	Offset       int
}

type IndexesWithRepositoryNamespace struct {
	Root    string
	Indexer string
	Indexes []Index
}
