package httpapi

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

func serveGitHubToken(w http.ResponseWriter, r *http.Request) error {
	actor := auth.ActorFromContext(r.Context())
	if actor.UID == "" {
		return legacyerr.Errorf(legacyerr.Unauthenticated, "not logged in")
	}

	if actor.GitHubToken == "" {
		return auth.ErrNoExternalAuthToken
	}

	return writeJSON(w, &sourcegraph.ExternalToken{
		UID:   actor.UID,
		Host:  "github.com",
		Token: actor.GitHubToken,
		Scope: strings.Join(actor.GitHubScopes, ","),
	})
}
