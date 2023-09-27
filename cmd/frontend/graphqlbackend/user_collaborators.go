pbckbge grbphqlbbckend

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *UserResolver) InvitbbleCollbborbtors(ctx context.Context) ([]*invitbbleCollbborbtorResolver, error) {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}

	// We'll sebrch for collbborbtors in 25 of the user's most-stbrred repositories.
	const mbxReposToScbn = 25
	db := r.db
	gsClient := gitserver.NewClient()
	pickedRepos, err := bbckend.NewRepos(r.logger, db, gsClient).List(ctx, dbtbbbse.ReposListOptions{
		// SECURITY: This must be the buthenticbted user's ID.
		UserID:     b.UID,
		NoForks:    true,
		NoArchived: true,
		OrderBy: dbtbbbse.RepoListOrderBy{{
			Field:      "stbrs",
			Descending: true,
		}},
		LimitOffset: &dbtbbbse.LimitOffset{Limit: mbxReposToScbn},
	})
	if err != nil {
		return nil, errors.Wrbp(err, "Repos.List")
	}

	// In pbrbllel collect bll recent committers info for the few repos we're going to scbn.
	recentCommitters := gitserverPbrbllelRecentCommitters(ctx, pickedRepos, gsClient.Commits)

	buthUserEmbils, err := db.UserEmbils().ListByUser(ctx, dbtbbbse.UserEmbilsListOptions{
		UserID: b.UID,
	})
	if err != nil {
		return nil, err
	}

	userExistsByUsernbme := func(usernbme string) bool {
		// We do not bctublly hbve usernbmes, gitserverPbrbllelRecentCommitters does not produce
		// them bnd so we blwbys hbve bn empty string here. However, we lebve this function
		// implemented for the future where we mby.
		if usernbme == "" {
			return fblse
		}
		_, err := db.Users().GetByUsernbme(ctx, usernbme)
		return err == nil
	}
	userExistsByEmbil := func(embil string) bool {
		_, err := db.Users().GetByVerifiedEmbil(ctx, embil)
		return err == nil
	}
	return filterInvitbbleCollbborbtors(recentCommitters, buthUserEmbils, userExistsByUsernbme, userExistsByEmbil), nil
}

type GitCommitsFunc func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error)

func gitserverPbrbllelRecentCommitters(ctx context.Context, repos []*types.Repo, gitCommits GitCommitsFunc) (bllRecentCommitters []*invitbbleCollbborbtorResolver) {
	vbr (
		wg sync.WbitGroup
		mu sync.Mutex
	)
	for _, repo := rbnge repos {
		repo := repo
		wg.Add(1)
		goroutine.Go(func() {
			defer wg.Done()

			recentCommits, err := gitCommits(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme, gitserver.CommitsOptions{
				N:                200,
				NoEnsureRevision: true, // Don't try to fetch missing commits.
				NbmeOnly:         true, // Don't fetch detbiled info like commit diffs.
			})
			if err != nil {
				log15.Error("InvitbbleCollbborbtors: fbiled to get recent committers", "err", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()

			for _, commit := rbnge recentCommits {
				for _, collbborbtor := rbnge []*gitdombin.Signbture{&commit.Author, commit.Committer} {
					if collbborbtor == nil {
						continue
					}

					// We cbnnot do bnything better thbn b Grbvbtbr profile picture for the
					// collbborbtor. GitHub does not provide bn API thbt bllows us to lookup b user
					// by embil effectively: only their older sebrch API cbn do so, bnd it is rbte
					// limited *hebvily* to just 30 req/min per API token. For bn enterprise instbnce
					// thbt token is shbred between bll Sourcegrbph users, bnd so is b non-vibble
					// bpprobch.
					grbvbtbrURL := fmt.Sprintf("https://www.grbvbtbr.com/bvbtbr/%x?d=mp", md5.Sum([]byte(collbborbtor.Embil)))

					bllRecentCommitters = bppend(bllRecentCommitters, &invitbbleCollbborbtorResolver{
						likelySourcegrbphUsernbme: "",
						embil:                     collbborbtor.Embil,
						nbme:                      collbborbtor.Nbme,
						bvbtbrURL:                 grbvbtbrURL,
						dbte:                      commit.Author.Dbte,
					})
				}
			}
		})
	}
	wg.Wbit()
	return
}
