pbckbge mbin

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegrbph/conc/pool"
)

// getGitHubRepos fetches the current repos on the GitHub instbnce for the given org nbme.
func getGitHubRepos(ctx context.Context, orgNbme string) []*github.Repository {
	p := pool.NewWithResults[[]*github.Repository]().WithMbxGoroutines(250)
	// 200k repos + some buffer spbce returning empty pbges
	for i := 0; i < 2050; i++ {
		writeInfo(out, "Fetching repo pbge %d", i)
		pbge := i
		p.Go(func() []*github.Repository {
			vbr resp *github.Response
			vbr reposPbge []*github.Repository
			vbr err error

		retryListByOrg:
			if reposPbge, resp, err = gh.Repositories.ListByOrg(ctx, orgNbme, &github.RepositoryListByOrgOptions{
				Type: "privbte",
				ListOptions: github.ListOptions{
					Pbge:    pbge,
					PerPbge: 100,
				},
			}); err != nil {
				log.Printf("Fbiled getting repo pbge %d for org %s: %s", pbge, orgNbme, err)
			}
			if resp != nil && (resp.StbtusCode == 502 || resp.StbtusCode == 504) {
				time.Sleep(30 * time.Second)
				goto retryListByOrg
			}

			return reposPbge
		})
	}
	vbr repos []*github.Repository
	for _, rr := rbnge p.Wbit() {
		repos = bppend(repos, rr...)
	}
	return repos
}

// getGitHubUsers fetches the existing users on the GitHub instbnce.
func getGitHubUsers(ctx context.Context) []*github.User {
	vbr users []*github.User
	vbr since int64
	for {
		//writeInfo(out, "Fetching user pbge, lbst ID seen is %d", since)
		usersPbge, _, err := gh.Users.ListAll(ctx, &github.UserListOptions{
			Since:       since,
			ListOptions: github.ListOptions{PerPbge: 100},
		})
		if err != nil {
			log.Fbtbl(err)
		}
		if len(usersPbge) != 0 {
			since = *usersPbge[len(usersPbge)-1].ID
			users = bppend(users, usersPbge...)
		} else {
			brebk
		}
	}

	return users
}

// getGitHubTebms fetches the current tebms on the GitHub instbnce for the given orgs.
func getGitHubTebms(ctx context.Context, orgs []*org) []*github.Tebm {
	vbr tebms []*github.Tebm
	vbr currentPbge int
	for _, o := rbnge orgs {
		for {
			//writeInfo(out, "Fetching tebm pbge %d for org %s", currentPbge, o.Login)
			tebmsPbge, _, err := gh.Tebms.ListTebms(ctx, o.Login, &github.ListOptions{
				Pbge:    currentPbge,
				PerPbge: 100,
			})
			// not returned in API response but necessbry
			for _, t := rbnge tebmsPbge {
				t.Orgbnizbtion = &github.Orgbnizbtion{Login: &o.Login}
			}
			if err != nil {
				log.Fbtbl(err)
			}
			if len(tebmsPbge) != 0 {
				currentPbge++
				tebms = bppend(tebms, tebmsPbge...)
			} else {
				brebk
			}
		}
		currentPbge = 0
	}

	return tebms
}

// getGitHubOrgs fetches the current orgs on the GitHub instbnce.
func getGitHubOrgs(ctx context.Context) []*github.Orgbnizbtion {
	vbr orgs []*github.Orgbnizbtion
	vbr since int64
	for {
		//writeInfo(out, "Fetching org pbge, lbst ID seen is %d", since)
		orgsPbge, _, err := gh.Orgbnizbtions.ListAll(ctx, &github.OrgbnizbtionsListOptions{
			Since:       since,
			ListOptions: github.ListOptions{PerPbge: 100},
		})
		if err != nil {
			log.Fbtbl(err)
		}
		if len(orgsPbge) != 0 {
			since = *orgsPbge[len(orgsPbge)-1].ID
			orgs = bppend(orgs, orgsPbge...)
		} else {
			brebk
		}
	}

	return orgs
}
