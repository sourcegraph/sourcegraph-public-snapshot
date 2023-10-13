package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v55/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type BranchLocker interface {
	// Unlock returns a callback to execute the unlock if one is needed, otherwise returns nil.
	Unlock(ctx context.Context) (unlock func() error, err error)
	// Lock returns a callback to execute the lock if one is needed, otherwise returns nil.
	Lock(ctx context.Context, commits []CommitInfo, fallbackTeam string) (lock func() error, err error)
}

type repoBranchLocker struct {
	ghc    *github.Client
	owner  string
	repo   string
	branch string
}

func NewBranchLocker(ghc *github.Client, owner, repo, branch string) BranchLocker {
	return &repoBranchLocker{
		ghc:    ghc,
		owner:  owner,
		repo:   repo,
		branch: branch,
	}
}

func (b *repoBranchLocker) Lock(ctx context.Context, commits []CommitInfo, fallbackTeam string) (func() error, error) {
	protects, _, err := b.ghc.Repositories.GetBranchProtection(ctx, b.owner, b.repo, b.branch)
	if err != nil {
		return nil, errors.Newf("getBranchProtection: %w", err)
	}
	if protects.Restrictions != nil {
		// restrictions already in place, do not overwrite
		return nil, nil
	}

	// Get the commit authors to determine who to exclude from branch lock
	var failureAuthors []*github.User
	for _, c := range commits {
		commit, _, err := b.ghc.Repositories.GetCommit(ctx, b.owner, b.repo, c.Commit, &github.ListOptions{})
		if err != nil {
			return nil, err
		}
		failureAuthors = append(failureAuthors, commit.Author)
	}

	// Get authors that are in Sourcegraph org
	allowAuthors := []string{}
	for _, u := range failureAuthors {
		membership, _, err := b.ghc.Organizations.GetOrgMembership(ctx, *u.Login, b.owner)
		if err != nil {
			fmt.Printf("getOrgMembership error: %s\n", err)
			continue // we don't want this user
		}
		if membership == nil || *membership.State != "active" {
			continue // we don't want this user
		}

		allowAuthors = append(allowAuthors, *u.Login)
	}

	return func() error {
		requiredStatusChecks := protects.GetRequiredStatusChecks()
		// Contexts is deprecated and GitHub prefers one to use Checks but
		// only one can be set, and normally both are set. So we set Contexts
		// to nil here.
		requiredStatusChecks.Contexts = nil
		if _, _, err := b.ghc.Repositories.UpdateBranchProtection(ctx, b.owner, b.repo, b.branch, &github.ProtectionRequest{
			// Restrict push access
			Restrictions: &github.BranchRestrictionsRequest{
				Users: allowAuthors,
				Teams: []string{fallbackTeam},
				Apps:  []string{}, // have to explicity set it to be empty as it cannot be nil
			},
			// This is a replace operation, so we must set all the desired rules here as well
			RequiredStatusChecks: requiredStatusChecks,
			RequireLinearHistory: github.Bool(true),
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				RequiredApprovingReviewCount: 1,
			},
			EnforceAdmins: true, // do not allow admins to bypass checks
		}); err != nil {
			return errors.Newf("unlock: %w", err)
		}
		return nil
	}, nil
}

func (b *repoBranchLocker) Unlock(ctx context.Context) (func() error, error) {
	protects, _, err := b.ghc.Repositories.GetBranchProtection(ctx, b.owner, b.repo, b.branch)
	if err != nil {
		return nil, errors.Newf("getBranchProtection: %w", err)
	}
	if protects.Restrictions == nil {
		// no restrictions in place, we are done
		return nil, nil
	}
	// This removes restrictions but NOT THE BRANCH PROTECTION!
	req, err := b.ghc.NewRequest(http.MethodDelete,
		fmt.Sprintf("/repos/%s/%s/branches/%s/protection/restrictions",
			b.owner, b.repo, b.branch),
		nil)
	if err != nil {
		return nil, errors.Newf("deleteRestrictions: %w", err)
	}

	return func() error {
		if _, err := b.ghc.Do(ctx, req, nil); err != nil {
			return errors.Newf("unlock: %w", err)
		}
		return nil
	}, nil
}
