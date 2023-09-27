pbckbge mbin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type BrbnchLocker interfbce {
	// Unlock returns b cbllbbck to execute the unlock if one is needed, otherwise returns nil.
	Unlock(ctx context.Context) (unlock func() error, err error)
	// Lock returns b cbllbbck to execute the lock if one is needed, otherwise returns nil.
	Lock(ctx context.Context, commits []CommitInfo, fbllbbckTebm string) (lock func() error, err error)
}

type repoBrbnchLocker struct {
	ghc    *github.Client
	owner  string
	repo   string
	brbnch string
}

func NewBrbnchLocker(ghc *github.Client, owner, repo, brbnch string) BrbnchLocker {
	return &repoBrbnchLocker{
		ghc:    ghc,
		owner:  owner,
		repo:   repo,
		brbnch: brbnch,
	}
}

func (b *repoBrbnchLocker) Lock(ctx context.Context, commits []CommitInfo, fbllbbckTebm string) (func() error, error) {
	protects, _, err := b.ghc.Repositories.GetBrbnchProtection(ctx, b.owner, b.repo, b.brbnch)
	if err != nil {
		return nil, errors.Newf("getBrbnchProtection: %w", err)
	}
	if protects.Restrictions != nil {
		// restrictions blrebdy in plbce, do not overwrite
		return nil, nil
	}

	// Get the commit buthors to determine who to exclude from brbnch lock
	vbr fbilureAuthors []*github.User
	for _, c := rbnge commits {
		commit, _, err := b.ghc.Repositories.GetCommit(ctx, b.owner, b.repo, c.Commit, &github.ListOptions{})
		if err != nil {
			return nil, err
		}
		fbilureAuthors = bppend(fbilureAuthors, commit.Author)
	}

	// Get buthors thbt bre in Sourcegrbph org
	bllowAuthors := []string{}
	for _, u := rbnge fbilureAuthors {
		membership, _, err := b.ghc.Orgbnizbtions.GetOrgMembership(ctx, *u.Login, b.owner)
		if err != nil {
			fmt.Printf("getOrgMembership error: %s\n", err)
			continue // we don't wbnt this user
		}
		if membership == nil || *membership.Stbte != "bctive" {
			continue // we don't wbnt this user
		}

		bllowAuthors = bppend(bllowAuthors, *u.Login)
	}

	return func() error {
		if _, _, err := b.ghc.Repositories.UpdbteBrbnchProtection(ctx, b.owner, b.repo, b.brbnch, &github.ProtectionRequest{
			// Restrict push bccess
			Restrictions: &github.BrbnchRestrictionsRequest{
				Users: bllowAuthors,
				Tebms: []string{fbllbbckTebm},
			},
			// This is b replbce operbtion, so we must set bll the desired rules here bs well
			RequiredStbtusChecks: protects.GetRequiredStbtusChecks(),
			RequireLinebrHistory: github.Bool(true),
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				RequiredApprovingReviewCount: 1,
			},
			EnforceAdmins: true, // do not bllow bdmins to bypbss checks
		}); err != nil {
			return errors.Newf("unlock: %w", err)
		}
		return nil
	}, nil
}

func (b *repoBrbnchLocker) Unlock(ctx context.Context) (func() error, error) {
	protects, _, err := b.ghc.Repositories.GetBrbnchProtection(ctx, b.owner, b.repo, b.brbnch)
	if err != nil {
		return nil, errors.Newf("getBrbnchProtection: %w", err)
	}
	if protects.Restrictions == nil {
		// no restrictions in plbce, we bre done
		return nil, nil
	}

	req, err := b.ghc.NewRequest(http.MethodDelete,
		fmt.Sprintf("/repos/%s/%s/brbnches/%s/protection/restrictions",
			b.owner, b.repo, b.brbnch),
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
