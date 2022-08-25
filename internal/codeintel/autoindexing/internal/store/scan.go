package store

import (
	"strconv"
	"strings"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// scanIndexes scans a slice of indexes from the return value of `*Store.query`.
var scanIndexes = basestore.NewSliceScanner(scanIndex)

// scanFirstIndex scans a slice of indexes from the return value of `*Store.query` and returns the first.
var scanFirstIndex = basestore.NewFirstScanner(scanIndex)

func scanIndex(s dbutil.Scanner) (index shared.Index, err error) {
	var executionLogs []shared.ExecutionLogEntry
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

	index.ExecutionLogs = append(index.ExecutionLogs, executionLogs...)

	return index, nil
}

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}

// scanFirstIndexConfiguration scans a slice of index configurations from the return value of `*Store.query`
// and returns the first.
var scanFirstIndexConfiguration = basestore.NewFirstScanner(scanIndexConfiguration)

func scanIndexConfiguration(s dbutil.Scanner) (indexConfiguration shared.IndexConfiguration, err error) {
	return indexConfiguration, s.Scan(
		&indexConfiguration.ID,
		&indexConfiguration.RepositoryID,
		&indexConfiguration.Data,
	)
}

var scanIndexesWithCount = basestore.NewSliceWithCountScanner(scanIndexWithCount)

// scanIndexes scans a slice of indexes from the return value of `*Store.query`.
func scanIndexWithCount(s dbutil.Scanner) (index shared.Index, count int, err error) {
	var executionLogs []shared.ExecutionLogEntry

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
		&count,
	); err != nil {
		return index, 0, err
	}

	for _, entry := range executionLogs {
		index.ExecutionLogs = append(index.ExecutionLogs, entry)
	}

	return index, count, nil
}
