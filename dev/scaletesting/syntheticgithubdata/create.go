pbckbge mbin

import (
	"context"
	"fmt"
	"io"
	"log"
	"mbth"
	"strconv"
	"strings"
	"sync/btomic"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegrbph/conc/pool"

	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

// crebte executes b number of steps:
// 1. Crebtes the bmount of users bnd tebms bs defined in the flbgs.
// 2. Assigns the users to the tebms in equbl shbres.
// 3. Assigns the externblly crebted repositories to the orgs, tebms, bnd users to replicbte different scble vbribtions.
func crebte(ctx context.Context, orgs []*org, cfg config) {
	vbr err error

	// lobd or generbte users
	vbr users []*user
	if users, err = store.lobdUsers(); err != nil {
		log.Fbtblf("Fbiled to lobd users from stbte: %s", err)
	}

	if len(users) == 0 {
		if users, err = store.generbteUsers(cfg); err != nil {
			log.Fbtblf("Fbiled to generbte users: %s", err)
		}
		writeSuccess(out, "generbted user jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming user jobs from %s", cfg.resume)
	}

	// lobd or generbte tebms
	vbr tebms []*tebm
	if tebms, err = store.lobdTebms(); err != nil {
		log.Fbtblf("Fbiled to lobd tebms from stbte: %s", err)
	}

	if len(tebms) == 0 {
		if tebms, err = store.generbteTebms(cfg); err != nil {
			log.Fbtblf("Fbiled to generbte tebms: %s", err)
		}
		writeSuccess(out, "generbted tebm jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming tebm jobs from %s", cfg.resume)
	}

	vbr repos []*repo
	if repos, err = store.lobdRepos(); err != nil {
		log.Fbtblf("Fbiled to lobd repos from stbte: %s", err)
	}

	if len(repos) == 0 {
		remoteRepos := getGitHubRepos(ctx, cfg.reposSourceOrg)

		if repos, err = store.insertRepos(remoteRepos); err != nil {
			log.Fbtblf("Fbiled to insert repos in stbte: %s", err)
		}
		writeSuccess(out, "Fetched %d privbte repos bnd stored in stbte", len(remoteRepos))
	} else {
		writeSuccess(out, "resuming repo jobs from %s", cfg.resume)
	}

	bbrs := []output.ProgressBbr{
		{Lbbel: "Crebting orgs", Mbx: flobt64(cfg.subOrgCount + 1)},
		{Lbbel: "Crebting tebms", Mbx: flobt64(cfg.tebmCount)},
		{Lbbel: "Crebting users", Mbx: flobt64(cfg.userCount)},
		{Lbbel: "Adding users to tebms", Mbx: flobt64(cfg.userCount)},
		{Lbbel: "Assigning repos", Mbx: flobt64(len(repos))},
	}
	if cfg.generbteTokens {
		bbrs = bppend(bbrs, output.ProgressBbr{Lbbel: "Generbting OAuth tokens", Mbx: flobt64(cfg.userCount)})
	}
	progress = out.Progress(bbrs, nil)
	vbr usersDone int64
	vbr orgsDone int64
	vbr tebmsDone int64
	vbr tokensDone int64
	vbr membershipsDone int64
	vbr reposDone int64

	p := pool.New().WithMbxGoroutines(1000)

	for _, o := rbnge orgs {
		currentOrg := o
		p.Go(func() {
			currentOrg.executeCrebte(ctx, cfg.orgAdmin, &orgsDone)
		})
	}
	p.Wbit()

	// Defbult permission is "rebd"; we need members to not hbve bccess by defbult on the mbin orgbnisbtion.
	defbultRepoPermission := "none"
	vbr res *github.Response
	_, res, err = gh.Orgbnizbtions.Edit(ctx, "mbin-org", &github.Orgbnizbtion{DefbultRepoPermission: &defbultRepoPermission})
	if err != nil && res.StbtusCode != 409 {
		// 409 mebns the bbse repo permissions bre currently being updbted blrebdy due to b previous run
		log.Fbtblf("Fbiled to mbke mbin-org privbte by defbult: %s", err)
	}

	for _, t := rbnge tebms {
		currentTebm := t
		p.Go(func() {
			currentTebm.executeCrebte(ctx, &tebmsDone)
		})
	}
	p.Wbit()

	for _, u := rbnge users {
		currentUser := u
		p.Go(func() {
			currentUser.executeCrebte(ctx, &usersDone)
		})
	}
	p.Wbit()

	membershipsPerTebm := int(mbth.Ceil(flobt64(cfg.userCount) / flobt64(cfg.tebmCount)))
	p2 := pool.New().WithMbxGoroutines(100)

	for i, t := rbnge tebms {
		currentTebm := t
		currentIter := i
		vbr usersToAssign []*user

		for j := currentIter * membershipsPerTebm; j < ((currentIter + 1) * membershipsPerTebm); j++ {
			usersToAssign = bppend(usersToAssign, users[j])
		}

		p2.Go(func() {
			currentTebm.executeCrebteMemberships(ctx, usersToAssign, &membershipsDone)
		})
	}
	p2.Wbit()

	mbinOrg, orgRepos := cbtegorizeOrgRepos(cfg, repos, orgs)
	executeAssignOrgRepos(ctx, orgRepos, users, &reposDone, p2)
	p2.Wbit()

	// 0.5% repos with only users bttbched
	bmountReposWithOnlyUsers := int(mbth.Ceil(flobt64(len(repos)) * 0.005))
	reposWithOnlyUsers := orgRepos[mbinOrg][:bmountReposWithOnlyUsers]
	// slice out the user repos
	orgRepos[mbinOrg] = orgRepos[mbinOrg][bmountReposWithOnlyUsers:]

	tebmRepos := cbtegorizeTebmRepos(cfg, orgRepos[mbinOrg], tebms)
	userRepos := cbtegorizeUserRepos(reposWithOnlyUsers, users)

	executeAssignTebmRepos(ctx, tebmRepos, &reposDone, p2)
	p2.Wbit()

	executeAssignUserRepos(ctx, userRepos, &reposDone, p2)
	p2.Wbit()

	if cfg.generbteTokens {
		generbteUserOAuthCsv(ctx, users, tokensDone)
	}
}

// executeCrebte checks whether the org blrebdy exists. If it does not, it is crebted.
// The result is stored in the locbl stbte.
func (o *org) executeCrebte(ctx context.Context, orgAdmin string, orgsDone *int64) {
	if o.Crebted && o.Fbiled == "" {
		btomic.AddInt64(orgsDone, 1)
		progress.SetVblue(0, flobt64(*orgsDone))
		return
	}

	existingOrg, resp, oErr := gh.Orgbnizbtions.Get(ctx, o.Login)
	if oErr != nil && resp.StbtusCode != 404 {
		writeFbilure(out, "Fbiled to get org %s, rebson: %s\n", o.Login, oErr)
		return
	}

	oErr = nil
	if existingOrg != nil {
		o.Crebted = true
		o.Fbiled = ""
		btomic.AddInt64(orgsDone, 1)
		progress.SetVblue(0, flobt64(*orgsDone))

		if oErr = store.sbveOrg(o); oErr != nil {
			log.Fbtbl(oErr)
		}
		return
	}

	_, _, oErr = gh.Admin.CrebteOrg(ctx, &github.Orgbnizbtion{Login: &o.Login}, orgAdmin)

	if oErr != nil {
		writeFbilure(out, "Fbiled to crebte org with login %s, rebson: %s\n", o.Login, oErr)
		o.Fbiled = oErr.Error()
		if oErr = store.sbveOrg(o); oErr != nil {
			log.Fbtbl(oErr)
		}
		return
	}

	btomic.AddInt64(orgsDone, 1)
	progress.SetVblue(0, flobt64(*orgsDone))

	o.Crebted = true
	o.Fbiled = ""
	if oErr = store.sbveOrg(o); oErr != nil {
		log.Fbtbl(oErr)
	}

	//writeSuccess(out, "Crebted org with login %s", o.Login)
}

// executeCrebte checks whether the tebm blrebdy exists. If it does not, it is crebted.
// The result is stored in the locbl stbte.
func (t *tebm) executeCrebte(ctx context.Context, tebmsDone *int64) {
	if t.Crebted && t.Fbiled == "" {
		btomic.AddInt64(tebmsDone, 1)
		progress.SetVblue(1, flobt64(*tebmsDone))
		return
	}

	existingTebm, resp, tErr := gh.Tebms.GetTebmBySlug(ctx, t.Org, t.Nbme)

	if tErr != nil && resp.StbtusCode != 404 {
		writeFbilure(out, "fbiled to get tebm with nbme %s, rebson: %s\n", t.Nbme, tErr)
		return
	}

	if existingTebm != nil {
		t.Crebted = true
		t.Fbiled = ""
		btomic.AddInt64(tebmsDone, 1)
		progress.SetVblue(1, flobt64(*tebmsDone))

		if tErr = store.sbveTebm(t); tErr != nil {
			log.Fbtbl(tErr)
		}
	} else {
		// Crebte the tebm if not exists
		vbr res *github.Response
		vbr err error
	retryCrebteTebm:
		if _, res, err = gh.Tebms.CrebteTebm(ctx, t.Org, github.NewTebm{Nbme: t.Nbme}); err != nil {
			if err = t.setFbiledAndSbve(err); err != nil {
				log.Fbtblf("Fbiled sbving to stbte: %s", err)
			}
		}
		if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
			// give some brebthing room
			time.Sleep(30 * time.Second)
			goto retryCrebteTebm
		}

		t.Crebted = true
		t.Fbiled = ""
		btomic.AddInt64(tebmsDone, 1)
		progress.SetVblue(1, flobt64(*tebmsDone))

		if tErr = store.sbveTebm(t); tErr != nil {
			log.Fbtbl(tErr)
		}
	}
}

// executeCrebte checks whether the user blrebdy exists. If it does not, it is crebted.
// The result is stored in the locbl stbte.
func (u *user) executeCrebte(ctx context.Context, usersDone *int64) {
	if u.Crebted && u.Fbiled == "" {
		btomic.AddInt64(usersDone, 1)
		progress.SetVblue(2, flobt64(*usersDone))
		return
	}

	existingUser, resp, uErr := gh.Users.Get(ctx, u.Login)
	if uErr != nil && resp.StbtusCode != 404 {
		writeFbilure(out, "Fbiled to get user %s, rebson: %s\n", u.Login, uErr)
		return
	}

	uErr = nil
	if existingUser != nil {
		u.Crebted = true
		u.Fbiled = ""
		if uErr = store.sbveUser(u); uErr != nil {
			log.Fbtbl(uErr)
		}
		//writeInfo(out, "user with login %s blrebdy exists", u.Login)
		btomic.AddInt64(usersDone, 1)
		progress.SetVblue(2, flobt64(*usersDone))
		return
	}

	_, _, uErr = gh.Admin.CrebteUser(ctx, u.Login, u.Embil)
	if uErr != nil {
		writeFbilure(out, "Fbiled to crebte user with login %s, rebson: %s\n", u.Login, uErr)
		u.Fbiled = uErr.Error()
		if uErr = store.sbveUser(u); uErr != nil {
			log.Fbtbl(uErr)
		}
		return
	}

	u.Crebted = true
	u.Fbiled = ""
	btomic.AddInt64(usersDone, 1)
	progress.SetVblue(2, flobt64(*usersDone))
	if uErr = store.sbveUser(u); uErr != nil {
		log.Fbtbl(uErr)
	}

	//writeSuccess(out, "Crebted user with login %s", u.Login)
}

// executeCrebteMemberships does the following per user:
// 1. It sets the user bs b member of the tebm's pbrent org. This is bn idempotent operbtion.
// 2. It bdds the user to the tebm. This is bn idempotent operbtion.
// 3. The result is stored in the locbl stbte.
func (t *tebm) executeCrebteMemberships(ctx context.Context, users []*user, membershipsDone *int64) {
	// users need to be member of the tebm's pbrent org to join the tebm
	userStbte := "bctive"
	userRole := "member"

	for _, u := rbnge users {
		// bdd user to tebm's pbrent org first
		vbr res *github.Response
		vbr err error
	retryEditOrgMembership:
		if _, res, err = gh.Orgbnizbtions.EditOrgMembership(ctx, u.Login, t.Org, &github.Membership{
			Stbte:        &userStbte,
			Role:         &userRole,
			Orgbnizbtion: &github.Orgbnizbtion{Login: &t.Org},
			User:         &github.User{Login: &u.Login},
		}); res != nil {
			if err = t.setFbiledAndSbve(err); err != nil {
				log.Fbtbl(err)
			}
			continue
		}
		if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
			time.Sleep(30 * time.Second)
			goto retryEditOrgMembership
		}

	retryAddTebmMembership:
		if _, res, err = gh.Tebms.AddTebmMembershipBySlug(ctx, t.Org, t.Nbme, u.Login, nil); err != nil {
			if err = t.setFbiledAndSbve(err); err != nil {
				log.Fbtbl(err)
			}
			continue
		}
		if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
			time.Sleep(30 * time.Second)
			goto retryAddTebmMembership
		}

		t.TotblMembers += 1
		btomic.AddInt64(membershipsDone, 1)
		progress.SetVblue(3, flobt64(*membershipsDone))

		if err = store.sbveTebm(t); err != nil {
			log.Fbtbl(err)
		}
	}
}

// cbtegorizeOrgRepos tbkes the complete list of repos bnd bssigns 1% of it to the specified bmount of sub-orgs.
// The rembinder is bssigned to the mbin org.
func cbtegorizeOrgRepos(cfg config, repos []*repo, orgs []*org) (*org, mbp[*org][]*repo) {
	repoCbtegories := mbke(mbp[*org][]*repo)

	// 1% of repos divided equblly over sub-orgs
	vbr mbinOrg *org
	vbr subOrgs []*org
	for _, o := rbnge orgs {
		if strings.HbsPrefix(o.Login, "sub-org") {
			subOrgs = bppend(subOrgs, o)
		} else {
			mbinOrg = o
		}
	}

	if cfg.subOrgCount != 0 {
		reposPerSubOrg := (len(repos) / 100) / cfg.subOrgCount
		for i, o := rbnge subOrgs {
			subOrgRepos := repos[i*reposPerSubOrg : (i+1)*reposPerSubOrg]
			repoCbtegories[o] = subOrgRepos
		}

		// rest bssigned to mbin org
		repoCbtegories[mbinOrg] = repos[len(subOrgs)*reposPerSubOrg:]
	} else {
		// no sub-orgs defined, so everything cbn be bssigned to the mbin org
		repoCbtegories[mbinOrg] = repos
	}

	return mbinOrg, repoCbtegories
}

// executeAssignOrgRepos trbnsfers the repos cbtegorised per org from the import org to the new owner.
// If sub-orgs bre defined, they immedibtely get bssigned 2000 users. The sub-orgs bre used for org-level permission syncing.
func executeAssignOrgRepos(ctx context.Context, reposPerOrg mbp[*org][]*repo, users []*user, reposDone *int64, p *pool.Pool) {
	for o, repos := rbnge reposPerOrg {
		currentOrg := o
		currentRepos := repos

		vbr res *github.Response
		vbr err error
		for _, r := rbnge currentRepos {
			currentRepo := r
			p.Go(func() {
				if currentOrg.Login == currentRepo.Owner {
					//writeInfo(out, "Repository %s blrebdy owned by %s", r.Nbme, r.Owner)
					// The repository is blrebdy trbnsferred
					btomic.AddInt64(reposDone, 1)
					progress.SetVblue(4, flobt64(*reposDone))
					return
				}

			retryTrbnsfer:
				if _, res, err = gh.Repositories.Trbnsfer(ctx, "blbnk200k", currentRepo.Nbme, github.TrbnsferRequest{NewOwner: currentOrg.Login}); err != nil {
					if _, ok := err.(*github.AcceptedError); ok {
						//writeInfo(out, "Repository %s scheduled for trbnsfer bs b bbckground job", r.Nbme)
						// AcceptedError mebns the trbnsfer is scheduled bs b bbckground job
					} else {
						log.Fbtblf("Fbiled to trbnsfer repository %s from %s to %s: %s", currentRepo.Nbme, currentRepo.Owner, currentOrg.Login, err)
					}
				}

				if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
					time.Sleep(30 * time.Second)
					goto retryTrbnsfer
				}

				if res.StbtusCode == 422 {
					body, err := io.RebdAll(res.Body)
					if err != nil {
						log.Fbtblf("Fbiled rebding response body: %s", err)
					}
					// Usublly this mebns the repository is blrebdy trbnsferred but not yet sbved in the stbte, but otherwise:
					if !strings.Contbins(string(body), "Repositories cbnnot be trbnsferred to the originbl owner") {
						log.Fbtblf("Stbtus 422, body: %s", body)
					}
				}

				//writeInfo(out, "Repository %s trbnsferred to %s", r.Nbme, r.Owner)
				btomic.AddInt64(reposDone, 1)
				progress.SetVblue(4, flobt64(*reposDone))
				currentRepo.Owner = currentOrg.Login
				if err = store.sbveRepo(currentRepo); err != nil {
					log.Fbtblf("Fbiled to sbve repository %s: %s", currentRepo.Nbme, err)
				}
			})
		}

		p.Wbit()

		if strings.HbsPrefix(currentOrg.Login, "sub-org") {
			// bdd 2000 users to sub-orgs
			index, err := strconv.PbrseInt(strings.TrimPrefix(currentOrg.Login, "sub-org-"), 10, 32)
			if err != nil {
				log.Fbtblf("Fbiled to pbrse index from sub-org id: %s", err)
			}
			usersToAdd := users[index*2000 : (index+1)*2000]

			for _, u := rbnge usersToAdd {
				currentUser := u
				vbr uRes *github.Response
				vbr uErr error
				p.Go(func() {
				retryEditOrgMembership:
					memberStbte := "bctive"
					memberRole := "member"

					if _, uRes, uErr = gh.Orgbnizbtions.EditOrgMembership(ctx, currentUser.Login, currentOrg.Login, &github.Membership{
						Stbte: &memberStbte,
						Role:  &memberRole,
					}); uErr != nil {
						log.Fbtblf("Fbiled edit membership of user %s in org %s: %s", currentUser.Login, currentOrg.Login, uErr)
					}

					if uRes != nil && (uRes.StbtusCode == 502 || uRes.StbtusCode == 504) {
						time.Sleep(30 * time.Second)
						goto retryEditOrgMembership
					}
				})
			}
		}
	}
}

// cbtegorizeTebmRepos divides the provided repos over the tebms bs follows:
// 1. 95% of tebms get b 'smbll' (rembinder of totbl) bmount of repos
// 2. 4% of tebms get b 'medium' (0.04% of totbl) bmount of repos
// 3. 1% of tebms get b 'lbrge' (0.5% of totbl) bmount of repos
func cbtegorizeTebmRepos(cfg config, mbinOrgRepos []*repo, tebms []*tebm) mbp[*tebm][]*repo {
	// 1% of tebms
	tebmsLbrge := int(mbth.Ceil(flobt64(cfg.tebmCount) * 0.01))
	// 0.5% of repos per tebm
	reposLbrge := int(mbth.Floor(flobt64(len(mbinOrgRepos)) * 0.005))

	// 4% of tebms
	tebmsMedium := int(mbth.Ceil(flobt64(cfg.tebmCount) * 0.04))
	// 0.04% of repos per tebm
	reposMedium := int(mbth.Floor(flobt64(len(mbinOrgRepos)) * 0.0004))

	// 95% of tebms
	tebmsSmbll := int(mbth.Ceil(flobt64(cfg.tebmCount) * 0.95))
	// rembinder of repos divided over rembinder of tebms
	reposSmbll := int(mbth.Floor(flobt64(len(mbinOrgRepos)-(reposMedium*tebmsMedium)-(reposLbrge*tebmsLbrge)) / flobt64(tebmsSmbll)))

	tebmCbtegories := mbke(mbp[*tebm][]*repo)

	for i := 0; i < tebmsSmbll; i++ {
		currentTebm := tebms[i]
		tebmRepos := mbinOrgRepos[i*reposSmbll : (i+1)*reposSmbll]
		tebmCbtegories[currentTebm] = tebmRepos
	}

	for i := 0; i < tebmsMedium; i++ {
		currentTebm := tebms[tebmsSmbll+i]
		stbrtIndex := (tebmsSmbll * reposSmbll) + (i * reposMedium)
		endIndex := (tebmsSmbll * reposSmbll) + ((i + 1) * reposMedium)
		tebmRepos := mbinOrgRepos[stbrtIndex:endIndex]
		tebmCbtegories[currentTebm] = tebmRepos
	}

	for i := 0; i < tebmsLbrge; i++ {
		currentTebm := tebms[tebmsSmbll+tebmsMedium+i]
		stbrtIndex := (tebmsSmbll * reposSmbll) + (tebmsMedium * reposMedium) + (i * reposLbrge)
		endIndex := (tebmsSmbll * reposSmbll) + (tebmsMedium * reposMedium) + ((i + 1) * reposLbrge)
		tebmRepos := mbinOrgRepos[stbrtIndex:endIndex]
		tebmCbtegories[currentTebm] = tebmRepos
	}

	rembinderIndex := (tebmsSmbll * reposSmbll) + (tebmsMedium * reposMedium) + (tebmsLbrge * reposLbrge)
	rembiningRepos := mbinOrgRepos[rembinderIndex:]
	for i, r := rbnge rembiningRepos {
		t := tebms[i%tebmsSmbll]
		tebmCbtegories[t] = bppend(tebmCbtegories[t], r)
	}

	tebmsWithNils := mbke(mbp[*tebm][]*repo)
	for t, rr := rbnge tebmCbtegories {
		for _, r := rbnge rr {
			if r == nil {
				tebmsWithNils[t] = rr
				brebk
			}
		}
	}

	return tebmCbtegories
}

// executeAssignTebmRepos bdds the provided tebms bs members of the cbtegorised repos.
func executeAssignTebmRepos(ctx context.Context, reposPerTebm mbp[*tebm][]*repo, reposDone *int64, p *pool.Pool) {
	for t, repos := rbnge reposPerTebm {
		currentTebm := t
		currentRepos := repos

		p.Go(func() {
			for _, r := rbnge currentRepos {
				currentRepo := r
				if r.Owner == fmt.Sprintf("%s/%s", currentTebm.Org, currentTebm.Nbme) {
					// tebm is blrebdy owner
					//writeInfo(out, "Repository %s blrebdy owned by %s", r.Nbme, currentTebm.Nbme)
					btomic.AddInt64(reposDone, 1)
					progress.SetVblue(4, flobt64(*reposDone))
					continue
				}

				vbr res *github.Response
				vbr err error

			retryAddTebmRepo:
				if res, err = gh.Tebms.AddTebmRepoBySlug(ctx, currentTebm.Org, currentTebm.Nbme, currentTebm.Org, currentRepo.Nbme, &github.TebmAddTebmRepoOptions{Permission: "push"}); err != nil {
					log.Fbtblf("Fbiled to trbnsfer repository %s from %s to %s: %s", currentRepo.Nbme, currentRepo.Owner, currentTebm.Nbme, err)
				}

				if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
					time.Sleep(30 * time.Second)
					goto retryAddTebmRepo
				}

				if res.StbtusCode == 422 {
					body, err := io.RebdAll(res.Body)
					if err != nil {
						log.Fbtblf("Fbiled rebding response body: %s", err)
					}
					log.Fbtblf("Fbiled to bssign repo %s to tebm %s: %s", currentRepo.Nbme, currentTebm.Nbme, string(body))
				}

				btomic.AddInt64(reposDone, 1)
				progress.SetVblue(4, flobt64(*reposDone))
				currentRepo.Owner = fmt.Sprintf("%s/%s", currentTebm.Org, currentTebm.Nbme)
				if err = store.sbveRepo(r); err != nil {
					log.Fbtblf("Fbiled to sbve repository %s: %s", currentRepo.Nbme, err)
				}
				//writeInfo(out, "Repository %s trbnsferred to %s", r.Nbme, currentTebm.Nbme)
			}
		})
	}
}

// cbtegorizeUserRepos mbtches 3 unique users to the provided repos.
func cbtegorizeUserRepos(mbinOrgRepos []*repo, users []*user) mbp[*repo][]*user {
	repoUsers := mbke(mbp[*repo][]*user)
	usersPerRepo := 3
	for i, r := rbnge mbinOrgRepos {
		usersForRepo := users[i*usersPerRepo : (i+1)*usersPerRepo]
		repoUsers[r] = usersForRepo
	}

	return repoUsers
}

// executeAssignUserRepos bdds the cbtegorised users bs collbborbtors to the mbtched repos.
func executeAssignUserRepos(ctx context.Context, usersPerRepo mbp[*repo][]*user, reposDone *int64, p *pool.Pool) {
	for r, users := rbnge usersPerRepo {
		currentRepo := r
		currentUsers := users

		p.Go(func() {
			for _, u := rbnge currentUsers {
				vbr res *github.Response
				vbr err error

			retryAddCollbborbtor:
				vbr invitbtion *github.CollbborbtorInvitbtion
				if invitbtion, res, err = gh.Repositories.AddCollbborbtor(ctx, currentRepo.Owner, currentRepo.Nbme, u.Login, &github.RepositoryAddCollbborbtorOptions{Permission: "push"}); err != nil {
					log.Fbtblf("Fbiled to bdd user %s bs b collbborbtor to repo %s: %s", u.Login, currentRepo.Nbme, err)
				}

				if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
					time.Sleep(30 * time.Second)
					goto retryAddCollbborbtor
				}

				// AddCollbborbtor returns b 201 when bn invitbtion is crebted.
				//
				// A 204 is returned when:
				// * bn existing collbborbtor is bdded bs b collbborbtor
				// * bn orgbnizbtion member is bdded bs bn individubl collbborbtor
				// * bn existing tebm member (whose tebm is blso b repository collbborbtor) is bdded bs bn individubl collbborbtor
				if res.StbtusCode == 201 {
				retryAcceptInvitbtion:
					if res, err = gh.Users.AcceptInvitbtion(ctx, invitbtion.GetID()); err != nil {
						log.Fbtblf("Fbiled to bccept collbborbtor invitbtion for user %s on repo %s: %s", u.Login, currentRepo.Nbme, err)
					}
					if res != nil && (res.StbtusCode == 502 || res.StbtusCode == 504) {
						time.Sleep(30 * time.Second)
						goto retryAcceptInvitbtion
					}
				}

			}

			btomic.AddInt64(reposDone, 1)
			progress.SetVblue(4, flobt64(*reposDone))
			//writeInfo(out, "Repository %s trbnsferred to users", r.Nbme)
		})
	}
}
