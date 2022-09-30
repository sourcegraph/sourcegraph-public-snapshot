package dbstore

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// DependencySyncingJob is a subset of the lsif_dependency_syncing_jobs table and acts as the
// queue and execution record for indexing the dependencies of a particular completed upload.
type DependencySyncingJob struct {
	ID             int        `json:"id"`
	State          string     `json:"state"`
	FailureMessage *string    `json:"failureMessage"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ProcessAfter   *time.Time `json:"processAfter"`
	NumResets      int        `json:"numResets"`
	NumFailures    int        `json:"numFailures"`
	UploadID       int        `json:"uploadId"`
}

func (u DependencySyncingJob) RecordID() int {
	return u.ID
}

func scanDependencySyncingJob(s dbutil.Scanner) (job DependencySyncingJob, err error) {
	return job, s.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.UploadID,
	)
}

// DependencyIndexingJob is a subset of the lsif_dependency_indexing_jobs table and acts as the
// queue and execution record for indexing the dependencies of a particular completed upload.
type DependencyIndexingJob struct {
	ID                  int        `json:"id"`
	State               string     `json:"state"`
	FailureMessage      *string    `json:"failureMessage"`
	StartedAt           *time.Time `json:"startedAt"`
	FinishedAt          *time.Time `json:"finishedAt"`
	ProcessAfter        *time.Time `json:"processAfter"`
	NumResets           int        `json:"numResets"`
	NumFailures         int        `json:"numFailures"`
	UploadID            int        `json:"uploadId"`
	ExternalServiceKind string     `json:"externalServiceKind"`
	ExternalServiceSync time.Time  `json:"externalServiceSync"`
}

func (u DependencyIndexingJob) RecordID() int {
	return u.ID
}

func scanDependencyIndexingJob(s dbutil.Scanner) (job DependencyIndexingJob, err error) {
	return job, s.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.UploadID,
		&job.ExternalServiceKind,
		&job.ExternalServiceSync,
	)
}

func (s *Store) InsertCloneableDependencyRepo(ctx context.Context, dependency precise.Package) (new bool, err error) {
	ctx, _, endObservation := s.operations.insertCloneableDependencyRepo.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Bool("new", new),
			log.Object("dependency", fmt.Sprint(dependency)),
		}})
	}()

	_, new, err = basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(insertCloneableDependencyRepoQuery, dependency.Scheme, dependency.Name, dependency.Version)))
	return
}

const insertCloneableDependencyRepoQuery = `
-- source: internal/codeintel/stores/dbstore/dependency_index.go:InsertCloneableDependencyRepo
INSERT INTO lsif_dependency_repos (scheme, name, version)
VALUES (%s, %s, %s)
ON CONFLICT DO NOTHING
RETURNING 1
`

func (s *Store) InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error) {
	ctx, _, endObservation := s.operations.insertDependencyIndexingJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadId", uploadID),
		log.String("extSvcKind", externalServiceKind),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", id),
		}})
	}()

	id, _, err = basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(insertDependencyIndexingJobQuery, uploadID, externalServiceKind, syncTime)))
	return id, err
}

const insertDependencyIndexingJobQuery = `
-- source: internal/codeintel/stores/dbstore/dependency_index.go:InsertDependencyIndexingJob
INSERT INTO lsif_dependency_indexing_jobs (upload_id, external_service_kind, external_service_sync)
VALUES (%s, %s, %s)
RETURNING id
`
