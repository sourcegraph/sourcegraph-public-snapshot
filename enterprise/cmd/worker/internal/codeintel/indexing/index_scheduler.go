package indexing

import (
	"context"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type IndexScheduler struct {
	dbStore       DBStore
	settingStore  IndexingSettingStore
	repoStore     IndexingRepoStore
	indexEnqueuer IndexEnqueuer
	operations    *schedulerOperations
}

var (
	_ goroutine.Handler      = &IndexScheduler{}
	_ goroutine.ErrorHandler = &IndexScheduler{}
)

func NewIndexScheduler(
	dbStore DBStore,
	settingStore IndexingSettingStore,
	repoStore IndexingRepoStore,
	indexEnqueuer IndexEnqueuer,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	scheduler := &IndexScheduler{
		dbStore:       dbStore,
		settingStore:  settingStore,
		repoStore:     repoStore,
		indexEnqueuer: indexEnqueuer,
		operations:    newOperations(observationContext),
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

	disabledRepoGroups, err := s.getDisabledRepositoryIDMap(ctx)
	if err != nil {
		return err
	}

	repositoryIDSourcers := []func(ctx context.Context) ([]int, error){
		s.getRepositoryIDsWithIndexConfiguration,
		s.getRepositoryIDsFromRepositoryGroups,
		s.getRepositoryIDsByPopularity,
	}

	repositoryIDMap := map[int]struct{}{}
	for _, repositoryIDSourcer := range repositoryIDSourcers {
		repositoryIDs, err := repositoryIDSourcer(ctx)
		if err != nil {
			return err
		}

		for _, repositoryID := range repositoryIDs {
			if _, ok := disabledRepoGroups[repositoryID]; !ok {
				repositoryIDMap[repositoryID] = struct{}{}
			}
		}
	}

	repositoryIDs := make([]int, 0, len(repositoryIDMap))
	for repositoryID := range repositoryIDMap {
		repositoryIDs = append(repositoryIDs, repositoryID)
	}
	sort.Ints(repositoryIDs)

	var queueErr error
	for _, repositoryID := range repositoryIDs {
		if err := s.indexEnqueuer.QueueIndexesForRepository(ctx, repositoryID); err != nil {
			if errors.HasType(err, &gitserver.RevisionNotFoundError{}) {
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

func (s *IndexScheduler) getDisabledRepositoryIDMap(ctx context.Context) (map[int]struct{}, error) {
	disabledRepoGroupsList, err := s.dbStore.GetAutoindexDisabledRepositories(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "DBStore.GetAutoindexDisabledRepositories")
	}

	disabledRepoGroups := make(map[int]struct{}, len(disabledRepoGroupsList))
	for _, v := range disabledRepoGroupsList {
		disabledRepoGroups[v] = struct{}{}
	}

	return disabledRepoGroups, nil
}

func (s *IndexScheduler) getRepositoryIDsWithIndexConfiguration(ctx context.Context) ([]int, error) {
	configuredRepositoryIDs, err := s.dbStore.GetRepositoriesWithIndexConfiguration(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "DBStore.GetRepositoriesWithIndexConfiguration")
	}

	return configuredRepositoryIDs, nil
}

func (s *IndexScheduler) getRepositoryIDsFromRepositoryGroups(ctx context.Context) ([]int, error) {
	// TODO(autoindex): We should create a way to gather _all_ repogroups (including all user repogroups)
	//    https://github.com/sourcegraph/sourcegraph/issues/22130
	settings, err := s.settingStore.GetLastestSchemaSettings(ctx, api.SettingsSubject{})
	if err != nil {
		return nil, errors.Wrap(err, "IndexingSettingStore.GetLastestSchemaSettings")
	}

	// TODO(autoindex): Later we can remove using cncf explicitly and do all of them
	//    https://github.com/sourcegraph/sourcegraph/issues/22130
	groupsByName := searchrepos.ResolveRepoGroupsFromSettings(settings)
	includePatterns, _ := searchrepos.RepoGroupsToIncludePatterns(settings.CodeIntelligenceAutoIndexRepositoryGroups, groupsByName)

	options := database.ReposListOptions{
		IncludePatterns: []string{includePatterns},
		OnlyCloned:      true,
		NoForks:         true,
		NoArchived:      true,
		NoPrivate:       true,
	}

	repositories, err := s.repoStore.ListRepoNames(ctx, options)
	if err != nil {
		return nil, errors.Wrap(err, "IndexingRepoStore.ListRepoNames")
	}

	return extractIDs(repositories), nil
}

func (s *IndexScheduler) getRepositoryIDsByPopularity(ctx context.Context) ([]int, error) {
	settings, err := s.settingStore.GetLastestSchemaSettings(ctx, api.SettingsSubject{})
	if err != nil {
		return nil, errors.Wrap(err, "IndexingSettingStore.GetLastestSchemaSettings")
	}

	if settings.CodeIntelligenceAutoIndexPopularRepoLimit == 0 {
		return nil, nil
	}

	repositories, err := s.repoStore.ListIndexableRepos(ctx, database.ListIndexableReposOptions{
		LimitOffset: &database.LimitOffset{Limit: settings.CodeIntelligenceAutoIndexPopularRepoLimit},
	})
	if err != nil {
		return nil, errors.Wrap(err, "IndexingRepoStore.ListIndexableRepos")
	}

	return extractIDs(repositories), nil
}

func extractIDs(repositories []types.RepoName) []int {
	repositoryIDs := make([]int, 0, len(repositories))
	for _, repoGroupRepository := range repositories {
		repositoryIDs = append(repositoryIDs, int(repoGroupRepository.ID))
	}

	return repositoryIDs
}
