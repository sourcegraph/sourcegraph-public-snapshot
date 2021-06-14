package indexing

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

type IndexScheduler struct {
	dbStore                     DBStore
	settingStore                IndexingSettingStore
	repoStore                   IndexingRepoStore
	indexEnqueuer               IndexEnqueuer
	operations                  *operations
	batchSize                   int
	minimumTimeSinceLastEnqueue time.Duration
	minimumSearchCount          int
	minimumSearchRatio          float64
	minimumPreciseCount         int
}

var _ goroutine.Handler = &IndexScheduler{}

func NewIndexScheduler(
	dbStore DBStore,
	settingStore IndexingSettingStore,
	repoStore IndexingRepoStore,
	indexEnqueuer IndexEnqueuer,
	batchSize int,
	minimumTimeSinceLastEnqueue time.Duration,
	minimumSearchCount int,
	minimumSearchRatio float64,
	minimumPreciseCount int,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	scheduler := &IndexScheduler{
		dbStore:                     dbStore,
		settingStore:                settingStore,
		repoStore:                   repoStore,
		indexEnqueuer:               indexEnqueuer,
		batchSize:                   batchSize,
		minimumTimeSinceLastEnqueue: minimumTimeSinceLastEnqueue,
		minimumSearchCount:          minimumSearchCount,
		minimumSearchRatio:          minimumSearchRatio,
		minimumPreciseCount:         minimumPreciseCount,
		operations:                  newOperations(observationContext),
	}

	return goroutine.NewPeriodicGoroutineWithMetrics(
		context.Background(),
		interval,
		scheduler,
		scheduler.operations.HandleIndexScheduler,
	)
}

// For mocking in tests
var indexSchedulerEnabled = conf.CodeIntelAutoIndexingEnabled

func (s *IndexScheduler) Handle(ctx context.Context) error {
	if !indexSchedulerEnabled() {
		return nil
	}

	configuredRepositoryIDs, err := s.dbStore.GetRepositoriesWithIndexConfiguration(ctx)
	if err != nil {
		return errors.Wrap(err, "dbstore.GetRepositoriesWithIndexConfiguration")
	}

	// TODO(autoindex): We should create a way to gather _all_ repogroups (including all user repogroups)
	settings, _ := s.settingStore.GetLastestSchemaSettings(ctx, api.SettingsSubject{})

	// TODO(autoindex): Later we can remove using cncf explicitly and do all of them
	groupsByName := searchrepos.ResolveRepoGroupsFromSettings(settings)
	_, includePatterns := searchrepos.RepoGroupsToIncludePatterns([]string{"cncf"}, groupsByName)

	options := database.ReposListOptions{
		// Good
		UserID:          0,
		IncludePatterns: []string{includePatterns},
		PatternQuery:    nil,
		NoForks:         true,
		NoArchived:      true,
		NoCloned:        true,
		NoPrivate:       true,
		LimitOffset:     &database.LimitOffset{},

		// Not sure
		// Select:          []string{},
		// Query:           "",
		// ExcludePattern:  "",
		// Names:           []string{},
		// URIs:            []string{},
		// IDs:             []api.RepoID{},
		// SearchContextID: 0,
		// ServiceTypes:    []string{},
		// ExternalServiceIDs:          []int64{},
		// ExternalRepos:               []api.ExternalRepoSpec{},
		// ExternalRepoIncludePrefixes: []api.ExternalRepoSpec{},
		// ExternalRepoExcludePrefixes: []api.ExternalRepoSpec{},
	}

	repoGroupRepositoryIDs, err := s.repoStore.ListRepoNames(ctx, options)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(repoGroupRepositoryIDs)
	}

	disabledRepoGroupsList, err := s.dbStore.GetAutoindexDisabledRepositories(ctx)
	disabledRepoGroups := map[int]struct{}{}
	for _, v := range disabledRepoGroupsList {
		disabledRepoGroups[v] = struct{}{}
	}

	var indexableRepositoryIDs []int
	for _, indexableRepository := range repoGroupRepositoryIDs {
		repoID := int(indexableRepository.ID)
		if _, ok := disabledRepoGroups[repoID]; ok {
			continue
		}
		indexableRepositoryIDs = append(indexableRepositoryIDs, repoID)
	}

	var queueErr error
	for _, repositoryID := range deduplicateRepositoryIDs(configuredRepositoryIDs, indexableRepositoryIDs) {
		if err := s.indexEnqueuer.QueueIndexesForRepository(ctx, repositoryID); err != nil {
			if isRepoNotExist(err) {
				continue
			}

			if queueErr != nil {
				queueErr = err
			} else {
				queueErr = multierror.Append(queueErr, err)
			}
		}
	}
	if queueErr != nil {
		return queueErr
	}

	return nil
}

func (s *IndexScheduler) HandleError(err error) {
	log15.Error("Failed to update indexable repositories", "err", err)
}

func deduplicateRepositoryIDs(ids ...[]int) (repositoryIDs []int) {
	repositoryIDMap := map[int]struct{}{}
	for _, s := range ids {
		for _, v := range s {
			repositoryIDMap[v] = struct{}{}
		}
	}

	for repositoryID := range repositoryIDMap {
		repositoryIDs = append(repositoryIDs, repositoryID)
	}

	return repositoryIDs
}

func isRepoNotExist(err error) bool {
	for err != nil {
		if vcs.IsRepoNotExist(err) {
			return true
		}

		err = errors.Unwrap(err)
	}

	return false
}
