package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanIndexJob(s dbutil.Scanner) (indexJob shared.IndexJob, err error) {
	return indexJob, s.Scan(
		&indexJob.Indexer,
	)
}

var scanIndexJobs = basestore.NewSliceScanner(scanIndexJob)
