package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Index struct {
	ID                 int                          `json:"id"`
	Commit             string                       `json:"commit"`
	QueuedAt           time.Time                    `json:"queuedAt"`
	State              string                       `json:"state"`
	FailureMessage     *string                      `json:"failureMessage"`
	StartedAt          *time.Time                   `json:"startedAt"`
	FinishedAt         *time.Time                   `json:"finishedAt"`
	ProcessAfter       *time.Time                   `json:"processAfter"`
	NumResets          int                          `json:"numResets"`
	NumFailures        int                          `json:"numFailures"`
	RepositoryID       int                          `json:"repositoryId"`
	LocalSteps         []string                     `json:"local_steps"`
	RepositoryName     string                       `json:"repositoryName"`
	DockerSteps        []DockerStep                 `json:"docker_steps"`
	Root               string                       `json:"root"`
	Indexer            string                       `json:"indexer"`
	IndexerArgs        []string                     `json:"indexer_args"` // TODO - convert this to `IndexCommand string`
	Outfile            string                       `json:"outfile"`
	ExecutionLogs      []executor.ExecutionLogEntry `json:"execution_logs"`
	Rank               *int                         `json:"placeInQueue"`
	AssociatedUploadID *int                         `json:"associatedUpload"`
	ShouldReindex      bool                         `json:"shouldReindex"`
	RequestedEnvVars   []string                     `json:"requestedEnvVars"`
}

func (i Index) RecordID() int {
	return i.ID
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
