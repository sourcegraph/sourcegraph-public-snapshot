package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	ossDB "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func enterpriseCreateRepoHook(ctx context.Context, s repos.Store, repo *types.Repo) error {
	// If the repository is public, we don't have to check anything
	if !repo.Private {
		return nil
	}

	if prFeature := (&licensing.FeaturePrivateRepositories{}); licensing.Check(prFeature) == nil {
		if prFeature.Unrestricted {
			return nil
		}

		numPrivateRepos, err := s.RepoStore().Count(ctx, ossDB.ReposListOptions{OnlyPrivate: true})
		if err != nil {
			return err
		}

		if numPrivateRepos >= prFeature.MaxNumPrivateRepos {
			return errors.Newf("maximum number of private repositories (%d) reached", prFeature.MaxNumPrivateRepos)
		}

		return nil
	}

	return licensing.NewFeatureNotActivatedError("The private repositories feature is not activated for this license. Please upgrade your license to use this feature.")
}

func enterpriseUpdateRepoHook(ctx context.Context, s repos.Store, existingRepo *types.Repo, newRepo *types.Repo) error {
	// If the privacy of the repo remains the same, the number of private reops
	// should remain the same, so we don't have to check anything
	if existingRepo.Private == newRepo.Private {
		return nil
	}

	return enterpriseCreateRepoHook(ctx, s, newRepo)
}
