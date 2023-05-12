package database

import (
	"context"
	"encoding/json"

	"github.com/keegancsmith/sqlf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var eventUnmarshalErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "own_event_logs_processing_errors_total",
	Help:      "Number of errors during event logs processing for Sourcegraph Own",
})

type RecentViewSignalStore interface {
	Insert(ctx context.Context, userID int32, repoPathID, count int) error
	InsertPaths(ctx context.Context, userID int32, repoPathIDToCount map[int]int) error
	List(ctx context.Context, opts ListRecentViewSignalOpts) ([]RecentViewSummary, error)
	BuildAggregateFromEvents(ctx context.Context, events []*Event) error
}

type ListRecentViewSignalOpts struct {
	ViewerUserID int
	RepoID       api.RepoID
	Path         string
	LimitOffset  *LimitOffset
}

type RecentViewSummary struct {
	UserID     int32
	FilePathID int
	ViewsCount int
}

func RecentViewSignalStoreWith(other basestore.ShareableStore, logger log.Logger) RecentViewSignalStore {
	lgr := logger.Scoped("RecentViewSignalStore", "Store for a table containing a number of views of a single file by a given viewer")
	return &recentViewSignalStore{Store: basestore.NewWithHandle(other.Handle()), Logger: lgr}
}

type recentViewSignalStore struct {
	*basestore.Store
	Logger log.Logger
}

// repoMetadata is a struct with all necessary data related to repo which is
// needed for signal creation.
type repoMetadata struct {
	// repoID is an ID of the repo in the DB.
	repoID api.RepoID
	// pathToID is a map of actual absolute file path to its ID in `repo_paths`
	// table. This map is written twice in `BuildAggregateFromEvents` because pathID
	// is calculated after all the paths (i.e. keys of this map) are gathered and put
	// into this map.
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
	SET views_count = EXCLUDED.views_count + own_aggregate_recent_view.views_count
`

func (s *recentViewSignalStore) Insert(ctx context.Context, userID int32, repoPathID, count int) error {
	q := sqlf.Sprintf(insertRecentViewSignalFmtstr, userID, repoPathID, count)
	return s.Exec(ctx, q)
}

const bulkInsertRecentViewSignalsFmtstr = `
	INSERT INTO own_aggregate_recent_view(viewer_id, viewed_file_path_id, views_count)
	VALUES %s
	ON CONFLICT(viewer_id, viewed_file_path_id) DO UPDATE
	SET views_count = EXCLUDED.views_count + own_aggregate_recent_view.views_count
`

// InsertPaths inserts paths and view counts for a given `userID`. This function
// has a hard limit of 5000 entries to be inserted. If more than 5000 paths are
// provided, this function will only add the first 5000 values read from the map
// (i.e. 5000 random paths) without any errors.
func (s *recentViewSignalStore) InsertPaths(ctx context.Context, userID int32, repoPathIDToCount map[int]int) error {
	pathsNumber := len(repoPathIDToCount)
	if pathsNumber == 0 {
		return nil
	}
	if pathsNumber > 5000 {
		pathsNumber = 5000
	}
	values := make([]*sqlf.Query, 0, pathsNumber)
	for pathID, count := range repoPathIDToCount {
		if pathsNumber == 0 {
			break
		}
		values = append(values, sqlf.Sprintf("(%s, %s, %s)", userID, pathID, count))
		pathsNumber--
	}
	q := sqlf.Sprintf(bulkInsertRecentViewSignalsFmtstr, sqlf.Join(values, ","))
	return s.Exec(ctx, q)
}

const listRecentViewSignalsFmtstr = `
	SELECT o.viewer_id, o.viewed_file_path_id, o.views_count
	FROM own_aggregate_recent_view AS o
	-- Optional join with repo_paths table
	%s
	-- Optional WHERE clauses
	WHERE %s
	-- Order, limit
	ORDER BY o.id
	%s
`

func (s *recentViewSignalStore) List(ctx context.Context, opts ListRecentViewSignalOpts) ([]RecentViewSummary, error) {
	viewsScanner := basestore.NewSliceScanner(func(scanner dbutil.Scanner) (RecentViewSummary, error) {
		var summary RecentViewSummary
		if err := scanner.Scan(&summary.UserID, &summary.FilePathID, &summary.ViewsCount); err != nil {
			return RecentViewSummary{}, err
		}
		return summary, nil
	})
	return viewsScanner(s.Query(ctx, createListQuery(opts)))
}

func createListQuery(opts ListRecentViewSignalOpts) *sqlf.Query {
	joinClause := &sqlf.Query{}
	if opts.RepoID != 0 || opts.Path != "" {
		joinClause = sqlf.Sprintf("INNER JOIN repo_paths AS p ON p.id = o.viewed_file_path_id")
	}
	whereClause := sqlf.Sprintf("TRUE")
	wherePredicates := make([]*sqlf.Query, 0)
	if opts.RepoID != 0 {
		wherePredicates = append(wherePredicates, sqlf.Sprintf("p.repo_id = %s", opts.RepoID))
	}
	if opts.Path != "" {
		wherePredicates = append(wherePredicates, sqlf.Sprintf("p.absolute_path = %s", opts.Path))
	}
	if opts.ViewerUserID != 0 {
		wherePredicates = append(wherePredicates, sqlf.Sprintf("o.viewer_id = %s", opts.ViewerUserID))
	}
	if len(wherePredicates) > 0 {
		whereClause = sqlf.Sprintf("%s", sqlf.Join(wherePredicates, "AND"))
	}
	return sqlf.Sprintf(listRecentViewSignalsFmtstr, joinClause, whereClause, opts.LimitOffset.SQL())
}

// BuildAggregateFromEvents builds recent view signals from provided "ViewBlob"
// events. One signal has a userID, repoPathID and a count. This data is derived
// from the event, please refer to inline comments for more implementation
// details.
func (s *recentViewSignalStore) BuildAggregateFromEvents(ctx context.Context, events []*Event) error {
	// Map of repo name to repo ID and paths+repoPathIDs of files specified in
	// "ViewBlob" events. Used to aggregate all the paths for a single repo to then
	// call `ensureRepoPaths` and receive all path IDs necessary to store the
	// signals.
	repoNameToMetadata := make(map[string]repoMetadata)
	// Map of userID specified in a "ViewBlob" event to the map of visited path to
	// count of "ViewBlob"s for this path. Used to aggregate counts of path visits
	// for specific users and then insert this structured data into
	// `own_aggregate_recent_view` table.
	userToCountByPath := make(map[uint32]map[repoPathAndName]int)
	// Not found repos set, so we don't spam the DB with bad SQL queries more than once.
	notFoundRepos := make(map[string]struct{})

	// Iterating over each event only once and gathering data for both
	// `repoNameToMetadata` and `userToCountByPath` at the same time.
	db := NewDBWith(s.Logger, s)
	for _, event := range events {
		// Checking if the event has a repo name and a path. If it is not the case, we
		// cannot proceed with given event and skip it.
		var r repoPathAndName
		err := json.Unmarshal(event.PublicArgument, &r)
		if err != nil {
			eventUnmarshalErrorCounter.Inc()
			continue
		}
		// Incrementing the count for a user and path in a "compute if absent" way.
		countByPath, found := userToCountByPath[event.UserID]
		if !found {
			userToCountByPath[event.UserID] = make(map[repoPathAndName]int)
			countByPath = userToCountByPath[event.UserID]
		}
		countByPath[r] = countByPath[r] + 1
		// Finding and updating repo metadata, once per every path rep repo.
		if _, found := repoNameToMetadata[r.RepoName]; !found {
			// If the repo is not present in `repoNameToMetadata`, we need to query it from
			// the DB.
			if _, notFound := notFoundRepos[r.RepoName]; notFound {
				// If we already know that the repo cannot be found in the DB, we don't need to
				// make an extra unsuccessful query.
				continue
			}
			repo, err := db.Repos().GetByName(ctx, api.RepoName(r.RepoName))
			if err != nil {
				if errcode.IsNotFound(err) {
					notFoundRepos[r.RepoName] = struct{}{}
				} else {
					return errors.Wrap(err, "error during fetching the repository")
				}
				continue
			}
			// For each repo we need to initialize a map of path to pathID. PathID is
			// initially set to 0, because we will know the real ID only after
			// `ensureRepoPaths` call.
			paths := make(map[string]int)
			paths[r.FilePath] = 0
			repoNameToMetadata[r.RepoName] = repoMetadata{repoID: repo.ID, pathToID: paths}
		}
		// At this point repoMetadata is initialized, and we only need to add current
		// file path to it.
		repoNameToMetadata[r.RepoName].pathToID[r.FilePath] = 0
	}

	// Ensuring paths for every repo.
	for _, repoMetadata := range repoNameToMetadata {
		// `ensureRepoPaths` accepts a repoID (we have it) and a slice of paths we want
		// to ensure. For the sake of constant-time path lookups we have a map of paths,
		// that's why we need to convert it to slice here in order to pass to
		// `ensureRepoPaths`.
		paths := make([]string, 0, len(repoMetadata.pathToID))
		for path := range repoMetadata.pathToID {
			paths = append(paths, path)
		}
		repoPathIDs, err := ensureRepoPaths(ctx, s.Store, paths, repoMetadata.repoID)
		if err != nil {
			return errors.Wrap(err, "cannot insert repo paths")
		}
		// Populate pathID for every path. `ensureRepoPaths` returns paths in the same
		// order as we passed them as an input, we can rely on that.
		for idx, path := range paths {
			repoMetadata.pathToID[path] = repoPathIDs[idx]
		}
	}

	// Now that we have all the necessary data, we go on and create signals.
	for userID, pathAndCount := range userToCountByPath {
		// Make a map of pathID->count from 2 maps that we have: path->count and
		// path->pathID.
		repoPathIDToCount := make(map[int]int)
		for rpn, count := range pathAndCount {
			if pathID, found := repoNameToMetadata[rpn.RepoName].pathToID[rpn.FilePath]; found {
				repoPathIDToCount[pathID] = count
			} else if _, notFound := notFoundRepos[rpn.RepoName]; notFound {
				// repo was not found in the database, that's fine.
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
