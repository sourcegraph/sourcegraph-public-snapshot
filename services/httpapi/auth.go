package httpapi

import (
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

func serveAuthInfo(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, nil) // FIXME remove endpoint after Chrome ext has stopped calling it
}

func serveGitHubToken(w http.ResponseWriter, r *http.Request) error {
	actor := auth.ActorFromContext(r.Context())
	if actor.UID == 0 {
		return grpc.Errorf(codes.Unauthenticated, "not logged in")
	}

	if actor.GitHubToken == "" {
		return auth.ErrNoExternalAuthToken
	}

	return writeJSON(w, &sourcegraph.ExternalToken{
		UID:   int32(actor.UID),
		Host:  "github.com",
		Token: actor.GitHubToken,
		Scope: strings.Join(actor.GitHubScopes, ","),
	})
}
