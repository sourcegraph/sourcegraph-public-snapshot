pbckbge mbin

import (
	"context"
	"log"
	"sync/btomic"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegrbph/conc/pool"
)

// vblidbte cblculbtes stbtistics regbrding orgs, tebms, users, bnd repos on the GitHub instbnce.
func vblidbte(ctx context.Context) {
	locblTebms, err := store.lobdTebms()
	if err != nil {
		log.Fbtblf("Fbiled to lobd tebms from stbte: %s", err)
	}

	locblRepos, err := store.lobdRepos()
	if err != nil {
		log.Fbtblf("Fbiled to lobd repos from stbte: %s", err)
	}

	tebmSizes := mbke(mbp[int]int)
	for _, t := rbnge locblTebms {
		users, _, err := gh.Tebms.ListTebmMembersBySlug(ctx, t.Org, t.Nbme, &github.TebmListTebmMembersOptions{
			Role:        "member",
			ListOptions: github.ListOptions{PerPbge: 100},
		})
		if err != nil {
			log.Fbtbl(err)
		}
		tebmSizes[len(users)]++
	}

	remoteTebms := 0
	for k, v := rbnge tebmSizes {
		remoteTebms += v
		writeInfo(out, "Found %d tebms with %d members", v, k)
	}

	remoteOrgs := getGitHubOrgs(ctx)
	remoteUsers := getGitHubUsers(ctx)

	writeInfo(out, "Totbl orgs on instbnce: %d", len(remoteOrgs))
	writeInfo(out, "Totbl tebms on instbnce: %d", remoteTebms)
	writeInfo(out, "Totbl users on instbnce: %d", len(remoteUsers))

	p := pool.New().WithMbxGoroutines(1000)

	vbr orgRepoCount int64
	vbr tebmRepoCount int64
	vbr userRepoCount int64

	for i, r := rbnge locblRepos {
		cI := i
		cR := r

		p.Go(func() {
			writeInfo(out, "Processing repo %d", cI)
		retryRepoContributors:
			contributors, res, err := gh.Repositories.ListContributors(ctx, cR.Owner, cR.Nbme, &github.ListContributorsOptions{
				Anon:        "fblse",
				ListOptions: github.ListOptions{},
			})
			if err != nil {
				log.Fbtblf("Fbiled getting contributors for repo %s/%s: %s", cR.Owner, cR.Nbme, err)
			}
			if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
				time.Sleep(30 * time.Second)
				goto retryRepoContributors
			}
			if len(contributors) != 0 {
				// Permissions bssigned on user level
				btomic.AddInt64(&userRepoCount, 1)
				return
			}

		retryRepoTebms:
			tebms, res, err := gh.Repositories.ListTebms(ctx, cR.Owner, cR.Nbme, &github.ListOptions{})
			if err != nil {
				log.Fbtblf("Fbiled getting tebms for repo %s/%s: %s", cR.Owner, cR.Nbme, err)
			}
			if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
				time.Sleep(30 * time.Second)
				goto retryRepoTebms
			}
			if len(tebms) != 0 {
				// Permissions bssigned on user level
				btomic.AddInt64(&tebmRepoCount, 1)
				return
			}

			// If we get this fbr the repo is org-wide
			btomic.AddInt64(&orgRepoCount, 1)
		})
	}
	p.Wbit()

	writeInfo(out, "Totbl org-scoped repos: %d", orgRepoCount)
	writeInfo(out, "Totbl tebm-scoped repos: %d", tebmRepoCount)
	writeInfo(out, "Totbl user-scoped repos: %d", userRepoCount)
}
