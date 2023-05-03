package database

import (
	"context"
	"encoding/json"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RecentViewSignalStore interface {
	Insert(ctx context.Context, userID int32, repoPathID, count int) error
	InsertPaths(ctx context.Context, userID int32, repoPathIDToCount map[int]int) error
	List(ctx context.Context, opts ListRecentViewSignalOpts) ([]RecentViewSummary, error)
	BuildAggregateFromEvents(ctx context.Context, events []*Event) error
}

type ListRecentViewSignalOpts struct {
	ViewerID int
	RepoID   api.RepoID
	Path     string
}

type RecentViewSummary struct {
	UserID     int32
	FilePathID int
	ViewsCount int
}

func RecentViewSignalStoreWith(other basestore.ShareableStore, logger log.Logger) RecentViewSignalStore {
	return &recentViewSignalStore{Store: basestore.NewWithHandle(other.Handle()), Logger: logger}
}

type recentViewSignalStore struct {
	*basestore.Store
	Logger log.Logger
}

type repoMetadata struct {
	repoID   api.RepoID
	pathToID map[string]int
}

type repoPathAndName struct {
	FilePath string `json:"filePath,omitempty"`
	RepoName string `json:"repoName,omitempty"`
}

// ToID concatenates repo name and path to make a unique ID over a set of repo
// paths (provided that the set of repo names is unique).
func (r repoPathAndName) ToID() string {
	return r.RepoName + r.FilePath
}

const insertRecentViewSignalFmtstr = `
	INSERT INTO own_aggregate_recent_view(viewer_id, viewed_file_path_id, views_count)
	VALUES(%s, %s, %s)
	ON CONFLICT(viewer_id, viewed_file_path_id) DO UPDATE
	SET views_count = EXCLUDED.views_count
`

func (s *recentViewSignalStore) Insert(ctx context.Context, userID int32, repoPathID, count int) error {
	q := sqlf.Sprintf(insertRecentViewSignalFmtstr, userID, repoPathID, count)
	_, err := s.ExecResult(ctx, q)
	return err
}

const bulkInsertRecentViewSignalsFmtstr = `
	INSERT INTO own_aggregate_recent_view(viewer_id, viewed_file_path_id, views_count)
	VALUES %s
	ON CONFLICT(viewer_id, viewed_file_path_id) DO UPDATE
	SET views_count = EXCLUDED.views_count
`

func (s *recentViewSignalStore) InsertPaths(ctx context.Context, userID int32, repoPathIDToCount map[int]int) error {
	pathsNumber := len(repoPathIDToCount)
	if pathsNumber == 0 {
		return nil
	}
	values := make([]*sqlf.Query, 0, pathsNumber)
	for pathID, count := range repoPathIDToCount {
		values = append(values, sqlf.Sprintf("(%s, %s, %s)", userID, pathID, count))
	}
	q := sqlf.Sprintf(bulkInsertRecentViewSignalsFmtstr, sqlf.Join(values, ","))
	_, err := s.ExecResult(ctx, q)
	return err
}

// TODO(sashaostrikov): update query with opts
const listRecentViewSignalsFmtstr = `
	SELECT viewer_id, viewed_file_path_id, views_count
	FROM own_aggregate_recent_view
	ORDER BY id
	LIMIT 1000
`

func (s *recentViewSignalStore) List(ctx context.Context, _ ListRecentViewSignalOpts) ([]RecentViewSummary, error) {
	q := sqlf.Sprintf(listRecentViewSignalsFmtstr)

	// TODO(sashaostrikov): implement paging and use opts
	viewsScanner := basestore.NewSliceScanner(func(scanner dbutil.Scanner) (RecentViewSummary, error) {
		var summary RecentViewSummary
		if err := scanner.Scan(&summary.UserID, &summary.FilePathID, &summary.ViewsCount); err != nil {
			return RecentViewSummary{}, err
		}
		return summary, nil
	})

	return viewsScanner(s.Query(ctx, q))
}

// TODO(sashaostrikov): BuildAggregateFromEvents should be called from worker,
// which queries events like so:
//
//	db := NewDBWith(s.Logger, s)
//	events, err := db.EventLogs().ListEventsByName(ctx, viewBlobEventType, after, limit)
func (s *recentViewSignalStore) BuildAggregateFromEvents(ctx context.Context, events []*Event) error {
	// Map of repo name to repo ID and paths+repoPathIDs of files specified in
	// "ViewBlob" events.
	repoNameToMetadata := make(map[string]repoMetadata)
	// Map of userID specified in a "ViewBlob" event to the map of visited path to
	// count of "ViewBlob"s for this path.
	userToCountByPath := make(map[uint32]map[repoPathAndName]int)
	// Not found repos set, so we don't spam the DB with bad SQL queries more than once.
	notFoundRepos := make(map[string]struct{})

	// For each event we parse its data and fill the maps.
	db := NewDBWith(s.Logger, s)
	for _, event := range events {
		var r repoPathAndName
		err := json.Unmarshal(event.PublicArgument, &r)
		if err != nil {
			s.Logger.Error("unmarshalling repo path and name", log.Error(err))
			continue
		}
		// Incrementing the count for a user and path.
		countByPath, found := userToCountByPath[event.UserID]
		if !found {
			userToCountByPath[event.UserID] = make(map[repoPathAndName]int)
			countByPath = userToCountByPath[event.UserID]
		}
		countByPath[r] = countByPath[r] + 1
		// Finding and updating repo metadata, once per every path rep repo.
		if _, found := repoNameToMetadata[r.RepoName]; !found {
			repo, err := db.Repos().GetByName(ctx, api.RepoName(r.RepoName))
			if err != nil {
				if errcode.IsNotFound(err) {
					notFoundRepos[r.RepoName] = struct{}{}
				} else {
					s.Logger.Error("unmarshalling repo path and name", log.Error(err))
				}
				continue
			}
			paths := make(map[string]int)
			paths[r.FilePath] = 0
			repoNameToMetadata[r.RepoName] = repoMetadata{repoID: repo.ID, pathToID: paths}
		}
		// Filling a repo with paths.
		repoNameToMetadata[r.RepoName].pathToID[r.FilePath] = 0
	}

	// Ensuring paths for every repo.
	for _, repoMetadata := range repoNameToMetadata {
		paths := make([]string, 0, len(repoMetadata.pathToID))
		for path := range repoMetadata.pathToID {
			paths = append(paths, path)
		}
		repoPathIDs, err := ensureRepoPaths(ctx, s.Store, paths, repoMetadata.repoID)
		if err != nil {
			return errors.Wrap(err, "cannot insert repo paths")
		}
		// Populate pathID for every path. `ensureRepoPaths` returns paths in the same
		// order, we can rely on that.
		for idx, path := range paths {
			repoMetadata.pathToID[path] = repoPathIDs[idx]
		}
	}

	for userID, pathAndCount := range userToCountByPath {
		// Make a map of pathID to count.
		repoPathIDToCount := make(map[int]int)
		for rpn, count := range pathAndCount {
			if pathID, found := repoNameToMetadata[rpn.RepoName].pathToID[rpn.FilePath]; found {
				repoPathIDToCount[pathID] = count
			} else {
				return errors.Newf("cannot find id of path %q of repo %q: this is a bug", rpn.FilePath, rpn.RepoName)
			}
		}
		err := s.InsertPaths(ctx, int32(userID), repoPathIDToCount)
		if err != nil {
			return err
		}
	}

	return nil
}
