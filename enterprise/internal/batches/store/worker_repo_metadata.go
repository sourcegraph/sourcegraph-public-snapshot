package store

import (
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

const (
	repoMetadataMaxNumRetries = 10
	repoMetadataMaxNumResets  = 3
	repoMetadataMaxStalledAge = 60 * time.Second
	repoMetadataRetryAfter    = 60 * time.Second
)

var repoMetadataWorkerStoreOpts = dbworkerstore.Options[*types.RepoMetadataWithName]{
	Name:                 "batches_repo_metadata_worker_store",
	TableName:            "batch_changes_repo_metadata",
	ViewName:             "batch_changes_repo_metadata_with_repo_name",
	AlternateColumnNames: map[string]string{"id": "repo_id"},
	ColumnExpressions:    append(repoMetadataColumns, sqlf.Sprintf("name")),
	Scan:                 dbworkerstore.BuildWorkerScan(scanRepoMetadataWithName),
	OrderByExpression:    sqlf.Sprintf("queued_at"),
	StalledMaxAge:        repoMetadataMaxStalledAge,
	MaxNumResets:         repoMetadataMaxNumResets,
	RetryAfter:           repoMetadataRetryAfter,
	MaxNumRetries:        repoMetadataMaxNumRetries,
}

func NewRepoMetadataWorkerStore(handle basestore.TransactableHandle, observationContext *observation.Context) dbworkerstore.Store[*types.RepoMetadataWithName] {
	return dbworkerstore.NewWithMetrics(handle, repoMetadataWorkerStoreOpts, observationContext)
}

func scanRepoMetadataWithName(sc dbutil.Scanner) (*types.RepoMetadataWithName, error) {
	var meta types.RepoMetadataWithName

	err := sc.Scan(
		&meta.RepoID,
		&meta.CreatedAt,
		&meta.UpdatedAt,
		&meta.Ignored,
		&meta.RepoName,
	)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}
