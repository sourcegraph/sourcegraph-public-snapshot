package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var fileMetricsCache = rcache.NewWithTTL("lang:v1:FileMetrics", 300)

type FileMetricsStore interface {
	// CalculateAndStoreFileMetrics will calculate the file metrics and store them in the cache (database table).
	// It returns the metrics that were calculated, and any error that occurs while calculating.
	// If the error is not nil, then the metrics are not complete.
	// Not complete means:
	// - the languages are determined from the file name/extension, not from the content
	// - the byte, line and word counts are probably not accurate
	// The content is gotten from gitserver.
	// Errors are logged internally and returned to the caller.
	CalculateAndStoreFileMetrics(context.Context, types.MinimalRepo, string, api.CommitID) (*fileutil.FileMetrics, error)
	// GetFileMetrics queries the file metrics from the database.
	// Return values in order:
	// - the file metrics
	// - indicator of if the metrics were calculated comletely. If false, the metrics may be inaccurate.
	// returning true if there are metrics stored; false if not.
	GetFileMetrics(context.Context, api.RepoID, string, api.CommitID) *fileutil.FileMetrics
	// PutFileMetrics stores a mapping of file to language, updating an existing record if one exists
	// (based on repo + path + commit)
	PutFileMetrics(context.Context, api.RepoID, string, api.CommitID, *fileutil.FileMetrics, bool) error
}

type fileMetricsStore struct {
	*basestore.Store
	logger log.Logger
}

var _ FileMetricsStore = (*fileMetricsStore)(nil)

func MapFileToLangWith(logger log.Logger, other basestore.ShareableStore) FileMetricsStore {
	return &fileMetricsStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

var newLine = []byte{'\n'}

func (s *fileMetricsStore) GetFileMetrics(ctx context.Context, repoID api.RepoID, path string, commitID api.CommitID) (metrics *fileutil.FileMetrics) {
	cacheKey := fmt.Sprintf("%d:%s:%s", repoID, path, commitID)
	cacheValue, cacheHit := fileMetricsCache.Get(cacheKey)
	if cacheHit {
		metrics = &fileutil.FileMetrics{}
		if err := json.Unmarshal(cacheValue, &metrics); err != nil {
			s.logger.Warn("unmarshal file metrics failed", log.Error(err), log.Int32("repoID", int32(repoID)), log.String("path", path), log.String("commitID", string(commitID)))
			fileMetricsCache.Delete(cacheKey)
		} else {
			return
		}
	}
	if fm, err := s.queryFileMetricsRecord(ctx, repoID, path, commitID); err == nil {
		// cache hit!
		metrics = &fileutil.FileMetrics{}
		metrics.Languages = fm.Languages
		metrics.Bytes = uint64(fm.Bytes)
		metrics.Lines = uint64(fm.Lines)
		metrics.Words = uint64(fm.Words)
		// store in redis so it's accessible faster next time
		s.putInRedis(repoID, path, commitID, metrics)
	} else if !errors.Is(err, sql.ErrNoRows) {
		// Some kind of database error that is not no rows. Log it and continue.
		// Since there's also a redis cache, this should not add too much stress on gitserver.
		s.logger.Warn("select from file_metrics failed", log.Error(err), log.Int32("repoID", int32(repoID)), log.String("path", path), log.String("commitID", string(commitID)))
	}
	return
}

var upsertStoredFileMetrics = `
insert into
file_metrics (
	-- index
	repo_id,
	file_path,
	commit_sha,
	-- data
	size_in_bytes,
	line_count,
	languages,
	complete
)
values (
	-- index
	%s,
	hashtext(%s),
	%s,
	-- data
	%s,
	%s,
	%s,
	%s
)
on conflict (repo_id, file_path, commit_sha)
do update set size_in_bytes = %s, line_count = %s, languages = %s, complete = %s
`

func (s *fileMetricsStore) PutFileMetrics(ctx context.Context, repoID api.RepoID, path string, commitID api.CommitID, metrics *fileutil.FileMetrics, complete bool) error {

	s.putInRedis(repoID, path, commitID, metrics)

	q := sqlf.Sprintf(upsertStoredFileMetrics,
		// index
		repoID,
		path,
		dbutil.CommitBytea(commitID),
		// data for insert
		metrics.Bytes,
		metrics.Lines,
		metrics.Languages,
		complete,
		// data for update
		metrics.Bytes,
		metrics.Lines,
		metrics.Languages,
		complete,
	)
	err := s.Store.Exec(ctx, q)
	if err != nil {
		s.logger.Warn("upsert to file_metrics failed", log.Error(err), log.Int32("repoID", int32(repoID)), log.String("path", path), log.String("commitID", string(commitID)), log.Strings("languages", metrics.Languages))
	}

	// Return the db error, if there was one.
	// Ignore the redis error, as it's just a cache.
	return err
}

func (s *fileMetricsStore) CalculateAndStoreFileMetrics(ctx context.Context, repo types.MinimalRepo, path string, commitID api.CommitID) (metrics *fileutil.FileMetrics, err error) {

	metrics = &fileutil.FileMetrics{}

	// no cache hit (or not complete); calculate the metrics
	err = metrics.CalculateFileMetrics(
		ctx,
		path,
		func(ctx context.Context, path string) (io.ReadCloser, error) {
			return gitserver.NewClient().NewFileReader(ctx, authz.DefaultSubRepoPermsChecker, repo.Name, commitID, path)
		},
	)

	// only warn on errors (instead of fail), because we'll get some metrics anyway
	if err == nil {
		s.logger.Warn("CalculateFileMetrics failed", log.Error(err), log.Int32("repoID", int32(repo.ID)), log.String("path", path), log.String("commitID", string(commitID)))
	}

	// don't make the caller wait for the cache storage
	bgCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		s.PutFileMetrics(bgCtx, repo.ID, path, commitID, metrics, err != nil)
	}()

	// return the metrics and any error that may have occured when calculating
	return metrics, err
}

var getFileMetrics = `
SELECT
	id,
	repo_id,
	file_path,
	size_in_bytes,
	line_count,
	word_count,
	commit_sha,
	languages,
	complete
FROM
	file_metrics
WHERE
	repo_id = %s
	and file_path = %s
	and commit_sha = %s
`

func (s *fileMetricsStore) queryFileMetricsRecord(ctx context.Context, repoID api.RepoID, path string, commitID api.CommitID) (*types.StoredFileMetrics, error) {
	q := sqlf.Sprintf(getFileMetrics, repoID, path, dbutil.CommitBytea(commitID))
	var r types.StoredFileMetrics
	if err := s.QueryRow(ctx, q).Scan(
		&r.ID,
		&r.RepoID,
		&r.Path,
		&r.Bytes,
		&r.Lines,
		&r.Words,
		&r.CommitSHA,
		pq.Array(&r.Languages),
		&r.Complete,
	); err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *fileMetricsStore) putInRedis(repoID api.RepoID, path string, commitID api.CommitID, metrics *fileutil.FileMetrics) error {
	cacheValue, err := json.Marshal(metrics)
	if err != nil {
		s.logger.Warn("cache file metrics in redis failed", log.Error(err), log.Int32("repoID", int32(repoID)), log.String("path", path), log.String("commitID", string(commitID)), log.Strings("languages", metrics.Languages))
		return err
	}
	cacheKey := fmt.Sprintf("%d:%s:%s", repoID, path, commitID)
	fileMetricsCache.Set(cacheKey, cacheValue)
	return nil
}
