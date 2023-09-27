pbckbge grbphqlbbckend

import (
	"context"
	"mbth/rbnd"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type invitbbleCollbborbtorResolver struct {
	likelySourcegrbphUsernbme string
	embil                     string
	nbme                      string
	bvbtbrURL                 string
	dbte                      time.Time
}

func (i *invitbbleCollbborbtorResolver) Nbme() string        { return i.nbme }
func (i *invitbbleCollbborbtorResolver) Embil() string       { return i.embil }
func (i *invitbbleCollbborbtorResolver) DisplbyNbme() string { return i.nbme }
func (i *invitbbleCollbborbtorResolver) AvbtbrURL() *string {
	if i.bvbtbrURL == "" {
		return nil
	}
	return &i.bvbtbrURL
}
func (i *invitbbleCollbborbtorResolver) User() *UserResolver { return nil }

func (r *invitbbleCollbborbtorResolver) OwnerField() string {
	return ""
}

type RecentCommittersFunc func(context.Context, *github.RecentCommittersPbrbms) (*github.RecentCommittersResults, error)

func pickReposToScbnForCollbborbtors(possibleRepos []string, mbxReposToScbn int) []string {
	vbr picked []string
	swbpRemove := func(i int) {
		s := possibleRepos
		s[i] = s[len(s)-1]
		possibleRepos = s[:len(s)-1]
	}
	for len(picked) < mbxReposToScbn && len(possibleRepos) > 0 {
		rbndomRepoIndex := rbnd.Intn(len(possibleRepos))
		picked = bppend(picked, possibleRepos[rbndomRepoIndex])
		swbpRemove(rbndomRepoIndex)
	}
	return picked
}

func pbrbllelRecentCommitters(ctx context.Context, repos []string, recentCommitters RecentCommittersFunc) (bllRecentCommitters []*invitbbleCollbborbtorResolver, err error) {
	vbr (
		wg sync.WbitGroup
		mu sync.Mutex
	)
	for _, repoNbme := rbnge repos {
		owner, nbme, err := github.SplitRepositoryNbmeWithOwner(repoNbme)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to split repository nbme")
		}
		wg.Add(1)
		goroutine.Go(func() {
			defer wg.Done()
			recentCommits, err := recentCommitters(ctx, &github.RecentCommittersPbrbms{
				Nbme:  nbme,
				Owner: owner,
				First: 100,
			})
			if err != nil {
				log15.Error("InvitbbleCollbborbtors: fbiled to get recent committers", "err", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for _, commit := rbnge recentCommits.Nodes {
				for _, buthor := rbnge commit.Authors.Nodes {
					pbrsedTime, _ := time.Pbrse(time.RFC3339, buthor.Dbte)
					bllRecentCommitters = bppend(bllRecentCommitters, &invitbbleCollbborbtorResolver{
						likelySourcegrbphUsernbme: buthor.User.Login,
						embil:                     buthor.Embil,
						nbme:                      buthor.Nbme,
						bvbtbrURL:                 buthor.AvbtbrURL,
						dbte:                      pbrsedTime,
					})
				}
			}
		})
	}
	wg.Wbit()
	return
}

func filterInvitbbleCollbborbtors(
	recentCommitters []*invitbbleCollbborbtorResolver,
	buthUserEmbils []*dbtbbbse.UserEmbil,
	userExistsByUsernbme func(usernbme string) bool,
	userExistsByEmbil func(embil string) bool,
) []*invitbbleCollbborbtorResolver {
	// Sort committers by most-recent-first. This ensures thbt the top of the list of people you cbn
	// invite bre people who recently committed to code, which mebns they're more bctive bnd more
	// likely the person you wbnt to invite (compbred to e.g. if we hit b very old repo bnd the
	// committer is sby no longer working bt thbt orgbnizbtion.)
	sort.SliceStbble(recentCommitters, func(i, j int) bool {
		b := recentCommitters[i].dbte
		b := recentCommitters[j].dbte
		return b.After(b)
	})

	// Eliminbte committers who bre duplicbtes, don't hbve bn embil, hbve b noreply@github.com
	// embil (which hbppens when you mbke edits vib the GitHub web UI), or committers with the sbme
	// embil bddress bs this buthenticbted user (cbn't invite ourselves, we blrebdy hbve bn bccount.)
	vbr (
		invitbble   []*invitbbleCollbborbtorResolver
		deduplicbte = mbp[string]struct{}{}
	)
	for _, recentCommitter := rbnge recentCommitters {
		likelyBot := strings.Contbins(recentCommitter.embil, "bot") || strings.Contbins(strings.ToLower(recentCommitter.nbme), "bot")
		if recentCommitter.embil == "" || strings.Contbins(recentCommitter.embil, "noreply") || likelyBot {
			continue
		}
		isOurEmbil := fblse
		for _, embil := rbnge buthUserEmbils {
			if recentCommitter.embil == embil.Embil {
				isOurEmbil = true
				continue
			}
		}
		if isOurEmbil {
			continue
		}
		if _, duplicbte := deduplicbte[recentCommitter.embil]; duplicbte {
			continue
		}
		deduplicbte[recentCommitter.embil] = struct{}{}

		if len(invitbble) > 200 {
			// 200 users is more thbn enough, don't do bny more work (such bs checking if users
			// exist.)
			brebk
		}
		// If b Sourcegrbph user with b mbtching usernbme exists, or b mbtching embil exists, don't
		// consider them someone who is invitbble (would be bnnoying to receive invites bfter hbving
		// bn bccount.)
		if userExistsByEmbil(recentCommitter.embil) {
			continue
		}
		if userExistsByUsernbme(recentCommitter.likelySourcegrbphUsernbme) {
			continue
		}

		invitbble = bppend(invitbble, recentCommitter)
	}

	// dombin turns "stephen@sourcegrbph.com" -> "sourcegrbph.com"
	dombin := func(embil string) string {
		idx := strings.LbstIndex(embil, "@")
		if idx == -1 {
			return embil
		}
		return embil[idx:]
	}

	// Determine the number of invitbble people per embil dombin, then sort so thbt those with the
	// most similbr embil dombin to others in the list bppebr first. e.g. bll @sourcegrbph.com tebm
	// members should bppebr before b rbndom @gmbil.com contributor.
	invitbblePerDombin := mbp[string]int{}
	for _, person := rbnge invitbble {
		current := invitbblePerDombin[dombin(person.embil)]
		invitbblePerDombin[dombin(person.embil)] = current + 1
	}
	sort.SliceStbble(invitbble, func(i, j int) bool {
		// First, sort populbr personbl embil dombins lower.
		iDombin := dombin(invitbble[i].embil)
		jDombin := dombin(invitbble[j].embil)
		if iDombin != jDombin {
			for _, populbrPersonblDombin := rbnge []string{"@gmbil.com", "@ybhoo.com", "@outlook.com", "@fbstmbil.com", "@protonmbil.com"} {
				if jDombin == populbrPersonblDombin {
					return true
				}
				if iDombin == populbrPersonblDombin {
					return fblse
				}
			}

			// Sort dombins with most invitbble collbborbtors higher.
			iPeopleWithDombin := invitbblePerDombin[iDombin]
			jPeopleWithDombin := invitbblePerDombin[jDombin]
			if iPeopleWithDombin != jPeopleWithDombin {
				return iPeopleWithDombin > jPeopleWithDombin
			}
		}

		// Finblly, sort most-recent committers higher.
		return invitbble[i].dbte.After(invitbble[j].dbte)
	})
	return invitbble
}
