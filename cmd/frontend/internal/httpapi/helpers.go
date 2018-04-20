package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// writeJSON writes a JSON Content-Type header and a JSON-encoded object to the
// http.ResponseWriter.
func writeJSON(w http.ResponseWriter, v interface{}) error {
	// Return "[]" instead of "null" if v is a nil slice.
	if reflect.TypeOf(v).Kind() == reflect.Slice && reflect.ValueOf(v).IsNil() {
		v = []interface{}{}
	}

	// MarshalIndent takes about 30-50% longer, which
	// significantly increases the time it takes to handle and return
	// large HTTP API responses.
	w.Header().Set("content-type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(v)
}

// tryUpdateGitolitePhabricatorMetadata attempts to update Phabricator metadata for a Gitolite-sourced repository, if it
// is appropriate to do so.
func tryUpdateGitolitePhabricatorMetadata(ctx context.Context, gconf schema.GitoliteConnection, repoURI api.RepoURI, repoName string) {
	if gconf.PhabricatorMetadataCommand == "" {
		return
	}
	metadata, err := gitserver.DefaultClient.GetGitolitePhabricatorMetadata(ctx, gconf.Host, repoName)
	if err != nil {
		log15.Warn("Could not fetch valid Phabricator metadata for Gitolite repository", "repo", repoName, "error", err)
		return
	}
	if metadata.Callsign == "" {
		return
	}
	if err := api.InternalClient.PhabricatorRepoCreate(ctx, repoURI, metadata.Callsign, gconf.Host); err != nil {
		log15.Warn("Could not ensure Gitolite Phabricator mapping", "repo", repoName, "error", err)
	}
}
