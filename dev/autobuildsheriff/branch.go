package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
)

type branchLocker interface {
	// Unlock returns a callback to execute the unlock if one is needed, otherwise returns nil.
	Unlock(ctx context.Context) (unlock func() error, err error)
	// Lock returns a callback to execute the lock if one is needed, otherwise returns nil.
	Lock(ctx context.Context, commits []commitInfo, fallbackTeam string) (lock func() error, err error)
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

func (b *repoBranchLocker) Lock(ctx context.Context, commits []commitInfo, fallbackTeam string) (func() error, error) {
	protects, _, err := b.ghc.Repositories.GetBranchProtection(ctx, b.owner, b.repo, b.branch)
	if err != nil {
		return nil, fmt.Errorf("getBranchProtection: %+w", err)
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
	allowPushFromActors := []string{}
	for _, u := range failureAuthors {
		membership, _, err := b.ghc.Organizations.GetOrgMembership(ctx, *u.Login, b.owner)
		if err != nil {
			return nil, fmt.Errorf("getOrgMembership: %+w", err)
		}
		if membership == nil || *membership.State != "active" {
			continue // we don't want this user
		}

		allowPushFromActors = append(allowPushFromActors, *u.NodeID)
	}

	// Get fallback team
	if fallbackTeam != "" {
		team, _, err := b.ghc.Teams.GetTeamBySlug(ctx, b.owner, fallbackTeam)
		if err != nil {
			return nil, fmt.Errorf("getTeam: %+w", err)
		}
		allowPushFromActors = append(allowPushFromActors, *team.NodeID)
	}

	// Update protections with workaround
	var requiredStatusChecks []string
	if protects.GetRequiredStatusChecks() != nil {
		requiredStatusChecks = protects.GetRequiredStatusChecks().Contexts
	}

	return func() error {
		if err := b.setProtectionsWorkaround(ctx, allowPushFromActors, requiredStatusChecks); err != nil {
			return fmt.Errorf("lock: %+w", err)
		}
		return nil
	}, nil
}

// Gnarly workaround until https://github.com/github/feedback/discussions/8692 is resolved
// because the follwoing only works in the GraphQL API and not the REST API.
//
// On the REST API, we can't use the PUT endpoint for just adding restrictions because you
// cannot *enable* restrictions with it, so we have to use .Repositories.UpdateBranchProtection
// instead. This endpoint does not allow you to set:
//
//     RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
//         RequiredApprovingReviewCount: 0, // this fails
//     }
//
// Which we need to get "require pull request BUT do not require review".
func (b *repoBranchLocker) setProtectionsWorkaround(
	ctx context.Context,
	allowActors []string,
	requiredStatusChecks []string,
) error {
	// Get protections GraphQL IDs (not provided by REST endpoint)
	getProtections, err := b.ghc.NewRequest(http.MethodPost, "https://api.github.com/graphql",
		map[string]string{
			"query": fmt.Sprintf(`query {
			repository(owner: "%s", name: "%s") {
			  branchProtectionRules(first: 10) {
				nodes {
				  id
				  pattern
				}
			  }
			}
		  }`, b.owner, b.repo),
		})
	if err != nil {
		return fmt.Errorf("getProtections.Setup: %w", err)
	}
	type ErrorsResp struct {
		Errors []interface{} `json:"errors"`
	}
	type ProtectionsResp struct {
		Data struct {
			Repository struct {
				BranchProtectionRules struct {
					Nodes []struct {
						ID      string `json:"id"`
						Pattern string `json:"pattern"`
					} `json:"nodes"`
				} `json:"branchProtectionRules"`
			} `json:"repository"`
		} `json:"data"`
		ErrorsResp
	}
	var protections *ProtectionsResp
	_, err = b.ghc.Do(ctx, getProtections, &protections)
	if err != nil {
		return fmt.Errorf("getProtections.Do: %w", err)
	}
	if len(protections.Errors) > 0 {
		errs, _ := json.Marshal(protections.Errors)
		return fmt.Errorf("getProtections.Do: %s", errs)
	}

	// Find relevant protection
	var protectionID string
	for _, v := range protections.Data.Repository.BranchProtectionRules.Nodes {
		if v.Pattern == b.branch {
			protectionID = v.ID
			break
		}
	}
	if protectionID == "" {
		return fmt.Errorf("updateProtections.Setup: protection %q not found, got: %+v",
			b.branch, protections)
	}

	// Set up args
	pushActors := "[]"
	if len(allowActors) > 0 {
		data, err := json.Marshal(allowActors)
		if err != nil {
			return fmt.Errorf("updateProtections.Setup.actors: %w", err)
		}
		pushActors = string(data)
	}
	// Not JSON, build manually
	requiredChecks := "["
	for _, check := range requiredStatusChecks {
		requiredChecks += fmt.Sprintf("{context:%q}", check)
	}
	requiredChecks += "]"

	// Update protections
	mutation := fmt.Sprintf(`mutation {
		updateBranchProtectionRule(input: {
		  branchProtectionRuleId: "%s",

		  restrictsPushes: true,
		  pushActorIds: %s,

		  requiresApprovingReviews: true,
		  requiredApprovingReviewCount: 0,

		  requiresLinearHistory: true,
		  requiresStatusChecks: true,
		  requiredStatusChecks: %s,
		  requiresStrictStatusChecks: false
		}) {
		  clientMutationId
		}
	  }`, protectionID, pushActors, requiredChecks)
	fmt.Printf("updating protections:\n%s\n", mutation)
	updateProtections, err := b.ghc.NewRequest(http.MethodPost, "https://api.github.com/graphql",
		map[string]string{
			"query": mutation,
		})
	if err != nil {
		return fmt.Errorf("updateProtections.Setup: %w", err)
	}

	var updateResp *ErrorsResp
	_, err = b.ghc.Do(ctx, updateProtections, &updateResp)
	if err != nil {
		return fmt.Errorf("updateProtections.Do: %w", err)
	}
	if len(updateResp.Errors) > 0 {
		errs, _ := json.Marshal(updateResp.Errors)
		return fmt.Errorf("updateProtections.Do: %s", errs)
	}

	return nil
}

func (b *repoBranchLocker) Unlock(ctx context.Context) (func() error, error) {
	protects, _, err := b.ghc.Repositories.GetBranchProtection(ctx, b.owner, b.repo, b.branch)
	if err != nil {
		return nil, fmt.Errorf("getBranchProtection: %+w", err)
	}
	if protects.Restrictions == nil {
		// no restrictions in place, we are done
		return nil, nil
	}

	req, err := b.ghc.NewRequest(http.MethodDelete,
		fmt.Sprintf("/repos/%s/%s/branches/%s/protection/restrictions",
			b.owner, b.repo, b.branch),
		nil)
	if err != nil {
		return nil, fmt.Errorf("deleteRestrictions: %+w", err)
	}

	return func() error {
		if _, err := b.ghc.Do(ctx, req, nil); err != nil {
			return fmt.Errorf("unlock: %+w", err)
		}
		return nil
	}, nil
}
