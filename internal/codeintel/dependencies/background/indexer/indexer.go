package indexer

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type indexer struct {
	dependenciesSvc *dependencies.Service
	dbStore         DBStore
	policyMatcher   PolicyMatcher
}

var _ goroutine.Handler = &indexer{}
var _ goroutine.ErrorHandler = &indexer{}

// Everything that follows in this file is a skeletal reproduction of existing code in the autoindex
// scheduler. We are piggy-backing on the same structure here until we can consolidate uses of the
// store+policy matcher into a proper policies service.

// For mocking in tests
var lockfileIndexingEnabled = conf.CodeIntelLockfileIndexingEnabled

func (i *indexer) Handle(ctx context.Context) error {
	if !lockfileIndexingEnabled() {
		return nil
	}

	var repositoryMatchLimit *int
	if val := conf.CodeIntelAutoIndexingPolicyRepositoryMatchLimit(); val != -1 {
		repositoryMatchLimit = &val
	}

	repositories, err := i.dbStore.SelectRepositoriesForIndexScan(
		ctx,
		"last_lockfile_scan",
		"last_lockfile_scan_at",
		ConfigInst.RepositoryMinimumCheckInterval,
		conf.CodeIntelAutoIndexingAllowGlobalPolicies(),
		repositoryMatchLimit,
		ConfigInst.RepositoryBatchSize,
	)
	if err != nil {
		return errors.Wrap(err, "dbstore.SelectRepositoriesForIndexScan")
	}
	if len(repositories) == 0 {
		return nil
	}

	now := timeutil.Now()

	for _, repositoryID := range repositories {
		if repositoryErr := i.handleRepository(ctx, repositoryID, now); repositoryErr != nil {
			if err == nil {
				err = repositoryErr
			} else {
				err = errors.Append(err, repositoryErr)
			}
		}
	}

	return err
}

func (i *indexer) handleRepository(
	ctx context.Context,
	repositoryID int,
	now time.Time,
) error {
	repoName, err := i.dbStore.RepoName(ctx, repositoryID)
	if err != nil {
		return err
	}

	offset := 0

	for {
		policies, totalCount, err := i.dbStore.GetConfigurationPolicies(ctx, dbstore.GetConfigurationPoliciesOptions{
			RepositoryID: repositoryID,
			ForIndexing:  true,
			Limit:        ConfigInst.RepositoryBatchSize,
			Offset:       offset,
		})
		if err != nil {
			return errors.Wrap(err, "dbstore.GetConfigurationPolicies")
		}
		offset += len(policies)

		commitMap, err := i.policyMatcher.CommitsDescribedByPolicy(ctx, repositoryID, policies, now)
		if err != nil {
			return errors.Wrap(err, "policies.CommitsDescribedByPolicy")
		}

		revs := types.RevSpecSet{}
		for commit, policyMatches := range commitMap {
			if len(policyMatches) == 0 {
				continue
			}

			revs[api.RevSpec(commit)] = struct{}{}
		}
		repoRevs := map[api.RepoName]types.RevSpecSet{api.RepoName(repoName): revs}

		if err := i.dependenciesSvc.ResolveDependencies(ctx, repoRevs); err != nil {
			return err
		}

		if len(policies) == 0 || offset >= totalCount {
			return nil
		}
	}
}

func (r *indexer) HandleError(err error) {
	// TODO - add additional metrics
	// log.Error("Failed to index lockfiles", "error", err)
}
