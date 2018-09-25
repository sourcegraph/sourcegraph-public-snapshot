package bg

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// ApplyUserOrgMap enforces auth.userOrgMap, ensuring that users are joined
// to the orgs specified.
func ApplyUserOrgMap(ctx context.Context) {
	// If this exceeds the timeout (e.g., DB lock), there are probably other problems
	// occurring, but it will help if we fail faster and log an error in that case.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for userPattern, orgs := range conf.GetTODO().AuthUserOrgMap {
		if userPattern != "*" {
			log15.Warn("unsupported auth.userOrgMap user pattern (only \"*\" is supported)", "userPattern", userPattern)
			continue
		}
		if err := db.OrgMembers.CreateMembershipInOrgsForAllUsers(ctx, nil, orgs); err != nil {
			log15.Error("error applying auth.userOrgMap (users were not auto-joined to configured orgs)", "userPattern", userPattern, "orgs", orgs, "err", err)
		}
	}
}
