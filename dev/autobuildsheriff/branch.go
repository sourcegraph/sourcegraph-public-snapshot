package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
)

type branchLocker interface {
	Unlock(ctx context.Context) (modified bool, err error)
	Lock(ctx context.Context, commits []commitInfo, fallbackTeam string) (modified bool, err error)
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

func (b *repoBranchLocker) Lock(ctx context.Context, commits []commitInfo, fallbackTeam string) (bool, error) {
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
	allowPushFromActors := []string{}
	for _, u := range failureAuthors {
		membership, _, err := b.ghc.Organizations.GetOrgMembership(ctx, *u.Login, b.owner)
		if err != nil {
			return false, err
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
			return false, err
		}
		allowPushFromActors = append(allowPushFromActors, *team.NodeID)
	}

	if err := b.setProtections(ctx, allowPushFromActors); err != nil {
		return false, err
	}
	return true, nil
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
func (b *repoBranchLocker) setProtections(ctx context.Context, allowActors []string) error {
	// Query all protection
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
	type protectionsResp struct {
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
	}
	var protections *protectionsResp
	_, err = b.ghc.Do(ctx, getProtections, &protections)
	if err != nil {
		return fmt.Errorf("getProtections.Do: %w", err)
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

	// Update protections
	actors, err := json.Marshal(allowActors)
	if err != nil {
		return fmt.Errorf("updateProtections.Setup: %w", err)
	}
	updateProtections, err := b.ghc.NewRequest(http.MethodPost, "https://api.github.com/graphql",
		map[string]string{
			"query": fmt.Sprintf(`mutation {
			updateBranchProtectionRule(input: {
			  branchProtectionRuleId: "%s",
			  restrictsPushes: true,
			  pushActorIds: %s,
			  requiresApprovingReviews: true,
			  requiredApprovingReviewCount: 0,
			  requiresLinearHistory: false,
			}) {
			  clientMutationId
			}
		  }`, protectionID, actors),
		})
	if err != nil {
		return fmt.Errorf("updateProtections.Setup: %w", err)
	}
	_, err = b.ghc.Do(ctx, updateProtections, nil)
	if err != nil {
		return fmt.Errorf("updateProtections.Do: %w", err)
	}

	return nil
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
