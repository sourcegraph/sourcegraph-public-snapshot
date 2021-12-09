package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v31/github"
)

type branchLocker interface {
	Unlock(ctx context.Context) error
	Lock(ctx context.Context, allowAuthorEmails []string, allowTeams []string) error
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

func (b *repoBranchLocker) Lock(ctx context.Context, allowAuthorEmails []string, allowTeams []string) error {
	users, _, err := b.ghc.Search.Users(ctx, strings.Join(allowAuthorEmails, " OR "), &github.SearchOptions{})
	if err != nil {
		return err
	}

	var failureAuthorsUsers []string
	for _, u := range users.Users {
		// Make sure this user is in the Sourcegraph org
		membership, _, err := b.ghc.Organizations.GetOrgMembership(ctx, *u.Login, "sourcegraph")
		if err != nil {
			return err
		}
		if membership == nil || *membership.State != "active" {
			continue // we don't want this user
		}

		failureAuthorsUsers = append(failureAuthorsUsers, *u.Login)
	}

	restrictions := &github.BranchRestrictionsRequest{
		Users: failureAuthorsUsers,
		Teams: []string{"dev-experience"},
	}
	fmt.Printf("restricting push access to %q to %+v", b.branch, restrictions)
	_, _, err = b.ghc.Repositories.UpdateBranchProtection(ctx, b.owner, b.repo, b.branch, &github.ProtectionRequest{
		Restrictions: restrictions,
	})
	if err != nil {
		return err
	}

	return nil
}

func (b *repoBranchLocker) Unlock(ctx context.Context) error {
	_, _, err := b.ghc.Repositories.UpdateBranchProtection(ctx, b.owner, b.repo, b.branch, &github.ProtectionRequest{
		Restrictions: &github.BranchRestrictionsRequest{
			Users: []string{},
			Teams: []string{},
		},
	})
	return err
}
