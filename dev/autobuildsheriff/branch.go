package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
)

type branchLocker interface {
	Unlock(ctx context.Context) (modified bool, err error)
	Lock(ctx context.Context, commits []commitInfo, allowTeams []string) (modified bool, err error)
}

type repoBranchLocker struct {
	ghc    *github.Client
	owner  string
	repo   string
	branch string
}

func newBranchLocker(ghc *github.Client, owner, repo, branch string) branchLocker {
	return &repoBranchLocker{
		ghc:    ghc,
		owner:  owner,
		repo:   repo,
		branch: branch,
	}
}

func (b *repoBranchLocker) Lock(ctx context.Context, commits []commitInfo, allowTeams []string) (bool, error) {
	protects, _, err := b.ghc.Repositories.GetBranchProtection(ctx, b.owner, b.repo, b.branch)
	if err != nil {
		return false, err
	}
	if protects.Restrictions != nil {
		// restrictions already in place, do not overwrite
		return false, nil
	}

	// Get the commit authors to determine who to exclude from branch lock
	var failureAuthors []*github.User
	for _, c := range commits {
		commit, _, err := b.ghc.Repositories.GetCommit(ctx, b.owner, b.repo, c.Commit, &github.ListOptions{})
		if err != nil {
			return false, err
		}
		failureAuthors = append(failureAuthors, commit.Author)
	}

	// Get authors that are in Sourcegraph org
	failureAuthorsLogins := []string{}
	for _, u := range failureAuthors {
		membership, _, err := b.ghc.Organizations.GetOrgMembership(ctx, *u.Login, "sourcegraph")
		if err != nil {
			return false, err
		}
		if membership == nil || *membership.State != "active" {
			continue // we don't want this user
		}

		failureAuthorsLogins = append(failureAuthorsLogins, *u.Login)
	}

	// We can't use the PUT endpoint for just adding restrictions because you cannot
	// enable restrictions with it, so we use this endpoint to update all protections
	// instead.
	if _, _, err := b.ghc.Repositories.UpdateBranchProtection(ctx, b.owner, b.repo, b.branch, &github.ProtectionRequest{
		Restrictions: &github.BranchRestrictionsRequest{
			Users: failureAuthorsLogins,
			Teams: allowTeams,
		},
		// This is a replace operation, so we must set all the desired rules here as well
		RequireLinearHistory: github.Bool(true),
		// Internally GitHub represents "require PR" as:
		//
		//     has_required_reviews: on
		//     required_approving_review_count: 0
		//
		// Enabling "require PR" returns the following RequiredPullRequestReviews:
		//
		//     &{..., RequiredApprovingReviewCount:0}
		//
		// This is impossible to set via the API, however, since it failes with an error
		// saying that RequiredApprovingReviewCount must be > 0.
		//
		// RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
		// 	RequiredApprovingReviewCount: 0, // this fails
		// },
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (b *repoBranchLocker) Unlock(ctx context.Context) (bool, error) {
	protects, _, err := b.ghc.Repositories.GetBranchProtection(ctx, b.owner, b.repo, b.branch)
	if err != nil {
		return false, err
	}
	if protects.Restrictions == nil {
		// no restrictions in place, we are done
		return false, nil
	}

	req, err := b.ghc.NewRequest(http.MethodDelete,
		fmt.Sprintf("/repos/%s/%s/branches/%s/protection/restrictions",
			b.owner, b.repo, b.branch),
		nil)
	if err != nil {
		return false, err
	}
	_, err = b.ghc.Do(ctx, req, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}
