package store

import "time"

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

// ExecutionLogEntry represents a command run by the executor.
type ExecutionLogEntry struct {
	Key        string    `json:"key"`
	Command    []string  `json:"command"`
	StartTime  time.Time `json:"startTime"`
	ExitCode   *int      `json:"exitCode,omitempty"`
	Out        string    `json:"out,omitempty"`
	DurationMs *int      `json:"durationMs,omitempty"`
}

type DockerStep struct {
	Root     string   `json:"root"`
	Image    string   `json:"image"`
	Commands []string `json:"commands"`
}

// Upload is a subset of the lsif_uploads table and stores both processed and unprocessed
// records.
type Upload struct {
	ID                int        `json:"id"`
	Commit            string     `json:"commit"`
	Root              string     `json:"root"`
	VisibleAtTip      bool       `json:"visibleAtTip"`
	UploadedAt        time.Time  `json:"uploadedAt"`
	State             string     `json:"state"`
	FailureMessage    *string    `json:"failureMessage"`
	StartedAt         *time.Time `json:"startedAt"`
	FinishedAt        *time.Time `json:"finishedAt"`
	ProcessAfter      *time.Time `json:"processAfter"`
	NumResets         int        `json:"numResets"`
	NumFailures       int        `json:"numFailures"`
	RepositoryID      int        `json:"repositoryId"`
	RepositoryName    string     `json:"repositoryName"`
	Indexer           string     `json:"indexer"`
	IndexerVersion    string     `json:"indexer_version"`
	NumParts          int        `json:"numParts"`
	UploadedParts     []int      `json:"uploadedParts"`
	UploadSize        *int64     `json:"uploadSize"`
	Rank              *int       `json:"placeInQueue"`
	AssociatedIndexID *int       `json:"associatedIndex"`
}
