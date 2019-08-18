package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func Test_serveReposList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	getRepoURIsViaHTTP := func(t *testing.T) []string {
		t.Helper()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := serveReposList(w, r); err != nil {
				t.Errorf("calling serveReposList: %v", err)
			}
		}))
		defer ts.Close()
		resp, err := http.Post(ts.URL, "application/json; charset=utf8", bytes.NewReader([]byte(`{"Enabled": true, "Index": true}`)))
		if err != nil {
			t.Fatalf("calling http.Get: %v", err)
		}
		// Parse the response as in zoekt-sourcegraph-indexserver/main.go.
		type repo struct {
			URI string
		}
		var repos []repo
		err = json.NewDecoder(resp.Body).Decode(&repos)
		if err != nil {
			t.Fatalf("json decoding response: %v", err)
		}
		resp.Body.Close()
		if err != nil {
			t.Fatalf("closing response body: %v", err)
		}
		var URIs []string
		for _, r := range repos {
			URIs = append(URIs, r.URI)
		}
		return URIs
	}

	t.Run("all repos are returned for non-sourcegraph.com", func(t *testing.T) {
		ctx := dbtesting.TestContext(t)
		qs := []string{
			`INSERT INTO repo(uri, name, created_at, updated_at, description, language) VALUES ('github.com/alice/rabbitmq', 'github.com/alice/rabbitmq', '2015-01-01', '2016-01-01', '', '')`,
			`INSERT INTO repo(uri, name, created_at, updated_at, description, language) VALUES ('github.com/bob/jabberd', 'github.com/bob/jabberd', '2001-01-01', '2019-01-01', '', '')`,
		}
		for _, q := range qs {
			if _, err := dbconn.Global.ExecContext(ctx, q); err != nil {
				t.Fatal(err)
			}
		}
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
			return repos, nil
		}
		defer func() { db.MockAuthzFilter = nil }()
		URIs := getRepoURIsViaHTTP(t)
		wantURIs := []string{"github.com/alice/rabbitmq", "github.com/bob/jabberd"}
		if !reflect.DeepEqual(URIs, wantURIs) {
			t.Errorf("got %v, want %v", URIs, wantURIs)
		}
	})

	t.Run("only default repos are returned for sourcegraph.com", func(t *testing.T) {
		ctx := dbtesting.TestContext(t)
		qs := []string{
			`INSERT INTO repo(id, name) VALUES (1, 'github.com/vim/vim')`,
			`INSERT INTO repo(id, name) VALUES (2, 'github.com/torvalds/linux')`,
			`INSERT INTO default_repos(repo_id) VALUES (2)`,
		}
		for _, q := range qs {
			if _, err := dbconn.Global.ExecContext(ctx, q); err != nil {
				t.Fatal(err)
			}
		}
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(false)
		URIs := getRepoURIsViaHTTP(t)
		wantURIs := []string{"github.com/torvalds/linux"}
		if !reflect.DeepEqual(URIs, wantURIs) {
			t.Errorf("got %v, want %v", URIs, wantURIs)
		}
	})
}
