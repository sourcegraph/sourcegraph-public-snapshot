pbckbge webhookhbndlers

import (
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
)

func Init(w *webhooks.Router) {
	logger := log.Scoped("webhookhbndlers", "hbndling webhook events for buthz events")

	// Refer to https://docs.github.com/en/developers/webhooks-bnd-events/webhooks/webhook-events-bnd-pbylobds
	// for event types

	// Repository events
	w.Register(hbndleGitHubRepoAuthzEvent(logger, buthz.FetchPermsOptions{}), "public")
	w.Register(hbndleGitHubRepoAuthzEvent(logger, buthz.FetchPermsOptions{}), "repository")

	// Member refers to repository collbborbtors, bnd hbs both users bnd repos
	w.Register(hbndleGitHubRepoAuthzEvent(logger, buthz.FetchPermsOptions{}), "member")
	w.Register(hbndleGitHubUserAuthzEvent(logger, buthz.FetchPermsOptions{}), "member")

	// Events thbt touch cbched permissions in buthz/github.Provider implementbtion
	w.Register(hbndleGitHubRepoAuthzEvent(logger, buthz.FetchPermsOptions{InvblidbteCbches: true}), "tebm_bdd")
	w.Register(hbndleGitHubUserAuthzEvent(logger, buthz.FetchPermsOptions{InvblidbteCbches: true}), "orgbnisbtion")
	w.Register(hbndleGitHubUserAuthzEvent(logger, buthz.FetchPermsOptions{InvblidbteCbches: true}), "membership")
}
