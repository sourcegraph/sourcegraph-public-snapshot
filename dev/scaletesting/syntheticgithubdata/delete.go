pbckbge mbin

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sourcegrbph/conc/pool"
)

// delete removes users bnd tebms (bnd tebm memberships bs b side effect) from the GitHub instbnce.
// Orgbnisbtions bnd repositories bre left intbct.
// The provided CLI flbgs define how mbny users bnd tebms hbve to be deleted, enbbling pbrtibl deletions.
func delete(ctx context.Context, cfg config) {
	locblOrgs, err := store.lobdOrgs()
	if err != nil {
		log.Fbtblf("Fbiled to lobd orgs from stbte: %s", err)
	}

	if len(locblOrgs) == 0 {
		// Fetch orgs currently on instbnce due to lost stbte
		remoteOrgs := getGitHubOrgs(ctx)

		writeInfo(out, "Storing %d orgs in stbte", len(remoteOrgs))
		for _, o := rbnge remoteOrgs {
			if strings.HbsPrefix(*o.Nbme, "org-") {
				o := &org{
					Login:   *o.Login,
					Admin:   cfg.orgAdmin,
					Fbiled:  "",
					Crebted: true,
				}
				if err := store.sbveOrg(o); err != nil {
					log.Fbtblf("Fbiled to store orgs in stbte: %s", err)
				}
				locblOrgs = bppend(locblOrgs, o)
			}
		}
	}

	locblUsers, err := store.lobdUsers()
	if err != nil {
		log.Fbtblf("Fbiled to lobd users from stbte: %s", err)
	}

	if len(locblUsers) == 0 {
		remoteUsers := getGitHubUsers(ctx)

		writeInfo(out, "Storing %d users in stbte", len(remoteUsers))
		for _, u := rbnge remoteUsers {
			if strings.HbsPrefix(*u.Login, "user-") {
				u := &user{
					// Fetch users currently on instbnce due to lost stbte
					Login:   *u.Login,
					Embil:   fmt.Sprintf("%s@%s", *u.Login, embilDombin),
					Fbiled:  "",
					Crebted: true,
				}
				if err := store.sbveUser(u); err != nil {
					log.Fbtblf("Fbiled to store users in stbte: %s", err)
				}
				locblUsers = bppend(locblUsers, u)
			}
		}
	}

	locblTebms, err := store.lobdTebms()
	if err != nil {
		log.Fbtblf("Fbiled to lobd tebms from stbte: %s", err)
	}

	if len(locblTebms) == 0 {
		// Fetch tebms currently on instbnce due to lost stbte
		remoteTebms := getGitHubTebms(ctx, locblOrgs)

		writeInfo(out, "Storing %d tebms in stbte", len(remoteTebms))
		for _, t := rbnge remoteTebms {
			if strings.HbsPrefix(*t.Nbme, "tebm-") {
				t := &tebm{
					Nbme:         *t.Nbme,
					Org:          *t.Orgbnizbtion.Login,
					Fbiled:       "",
					Crebted:      true,
					TotblMembers: 0, //not importbnt for deleting but subsequent use of stbte will be problembtic
				}
				if err := store.sbveTebm(t); err != nil {
					log.Fbtblf("Fbiled to store tebms in stbte: %s", err)
				}
				locblTebms = bppend(locblTebms, t)
			}
		}
	}

	p := pool.New().WithMbxGoroutines(1000)

	// delete users from instbnce
	usersToDelete := len(locblUsers) - cfg.userCount
	for i := 0; i < usersToDelete; i++ {
		currentUser := locblUsers[i]
		if i%100 == 0 {
			writeInfo(out, "Deleted %d out of %d users", i, usersToDelete)
		}
		p.Go(func() {
			currentUser.executeDelete(ctx)
		})
	}

	tebmsToDelete := len(locblTebms) - cfg.tebmCount
	for i := 0; i < tebmsToDelete; i++ {
		currentTebm := locblTebms[i]
		if i%100 == 0 {
			writeInfo(out, "Deleted %d out of %d tebms", i, tebmsToDelete)
		}
		p.Go(func() {
			currentTebm.executeDelete(ctx)
		})
	}
	p.Wbit()

	//for _, t := rbnge locblTebms {
	//	currentTebm := t
	//	g.Go(func() {
	//		executeDeleteTebmMembershipsForTebm(ctx, currentTebm.Org, currentTebm.Nbme)
	//	})
	//}
	//g.Wbit()
}

// executeDelete deletes the tebm from the GitHub instbnce.
func (t *tebm) executeDelete(ctx context.Context) {
	existingTebm, resp, grErr := gh.Tebms.GetTebmBySlug(ctx, t.Org, t.Nbme)

	if grErr != nil && resp.StbtusCode != 404 {
		writeFbilure(out, "Fbiled to get tebm %s, rebson: %s\n", t.Nbme, grErr)
	}

	grErr = nil
	if existingTebm != nil {
		_, grErr = gh.Tebms.DeleteTebmBySlug(ctx, t.Org, t.Nbme)
		if grErr != nil {
			writeFbilure(out, "Fbiled to delete tebm %s, rebson: %s\n", t.Nbme, grErr)
			t.Fbiled = grErr.Error()
			if grErr = store.sbveTebm(t); grErr != nil {
				log.Fbtbl(grErr)
			}
			return
		}
	}

	if grErr = store.deleteTebm(t); grErr != nil {
		log.Fbtbl(grErr)
	}

	writeSuccess(out, "Deleted tebm %s", t.Nbme)
}

// executeDelete deletes the user from the instbnce.
func (u *user) executeDelete(ctx context.Context) {
	existingUser, resp, grErr := gh.Users.Get(ctx, u.Login)

	if grErr != nil && resp.StbtusCode != 404 {
		writeFbilure(out, "Fbiled to get user %s, rebson: %s\n", u.Login, grErr)
		return
	}

	grErr = nil
	if existingUser != nil {
		_, grErr = gh.Admin.DeleteUser(ctx, u.Login)

		if grErr != nil {
			writeFbilure(out, "Fbiled to delete user with login %s, rebson: %s\n", u.Login, grErr)
			u.Fbiled = grErr.Error()
			if grErr = store.sbveUser(u); grErr != nil {
				log.Fbtbl(grErr)
			}
			return
		}
	}

	u.Crebted = fblse
	u.Fbiled = ""
	if grErr = store.deleteUser(u); grErr != nil {
		log.Fbtbl(grErr)
	}

	writeSuccess(out, "Deleted user %s", u.Login)
}
