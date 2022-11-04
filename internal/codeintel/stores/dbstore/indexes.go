package dbstore

import (
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// Index is a subset of the lsif_indexes table and stores both processed and unprocessed
// records.
type Index struct {
	ID                 int                            `json:"id"`
	Commit             string                         `json:"commit"`
	QueuedAt           time.Time                      `json:"queuedAt"`
	State              string                         `json:"state"`
	FailureMessage     *string                        `json:"failureMessage"`
	StartedAt          *time.Time                     `json:"startedAt"`
	FinishedAt         *time.Time                     `json:"finishedAt"`
	ProcessAfter       *time.Time                     `json:"processAfter"`
	NumResets          int                            `json:"numResets"`
	NumFailures        int                            `json:"numFailures"`
	RepositoryID       int                            `json:"repositoryId"`
	LocalSteps         []string                       `json:"local_steps"`
	RepositoryName     string                         `json:"repositoryName"`
	DockerSteps        []DockerStep                   `json:"docker_steps"`
	Root               string                         `json:"root"`
	Indexer            string                         `json:"indexer"`
	IndexerArgs        []string                       `json:"indexer_args"` // TODO - convert this to `IndexCommand string`
	Outfile            string                         `json:"outfile"`
	ExecutionLogs      []workerutil.ExecutionLogEntry `json:"execution_logs"`
	Rank               *int                           `json:"placeInQueue"`
	AssociatedUploadID *int                           `json:"associatedUpload"`
}

func (i Index) RecordID() int {
	return i.ID
}

func scanIndex(s dbutil.Scanner) (index Index, err error) {
	var executionLogs []dbworkerstore.ExecutionLogEntry
	if err := s.Scan(
		&index.ID,
		&index.Commit,
		&index.QueuedAt,
		&index.State,
		&index.FailureMessage,
		&index.StartedAt,
		&index.FinishedAt,
		&index.ProcessAfter,
		&index.NumResets,
		&index.NumFailures,
		&index.RepositoryID,
		&index.RepositoryName,
		pq.Array(&index.DockerSteps),
		&index.Root,
		&index.Indexer,
		pq.Array(&index.IndexerArgs),
		&index.Outfile,
		pq.Array(&executionLogs),
		&index.Rank,
		pq.Array(&index.LocalSteps),
		&index.AssociatedUploadID,
	); err != nil {
		return index, err
	}

	for _, entry := range executionLogs {
		index.ExecutionLogs = append(index.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return index, nil
}

const indexAssociatedUploadIDQueryFragment = `
(
	SELECT MAX(id) FROM lsif_uploads WHERE associated_index_id = u.id
) AS associated_upload_id
`

var indexColumnsWithNullRank = []*sqlf.Query{
	sqlf.Sprintf("u.id"),
	sqlf.Sprintf("u.commit"),
	sqlf.Sprintf("u.queued_at"),
	sqlf.Sprintf("u.state"),
	sqlf.Sprintf("u.failure_message"),
	sqlf.Sprintf("u.started_at"),
	sqlf.Sprintf("u.finished_at"),
	sqlf.Sprintf("u.process_after"),
	sqlf.Sprintf("u.num_resets"),
	sqlf.Sprintf("u.num_failures"),
	sqlf.Sprintf("u.repository_id"),
	sqlf.Sprintf(`u.repository_name`),
	sqlf.Sprintf(`u.docker_steps`),
	sqlf.Sprintf(`u.root`),
	sqlf.Sprintf(`u.indexer`),
	sqlf.Sprintf(`u.indexer_args`),
	sqlf.Sprintf(`u.outfile`),
	sqlf.Sprintf(`u.execution_logs`),
	sqlf.Sprintf("NULL"),
	sqlf.Sprintf(`u.local_steps`),
	sqlf.Sprintf(indexAssociatedUploadIDQueryFragment),
}

type IndexesWithRepositoryNamespace struct {
	Root    string
	Indexer string
	Indexes []Index
}
