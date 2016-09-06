package backend_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/testutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/testserver"
)

// TODO: Think if there's a better location for this type of test. It's meant to be run
//       by humans occassionally to look at search quality, and during development. It
//       requires manual providing of GH credentials, and human analysis of the results.

// TestSearchRepos is a test that is not meant to be run automatically as part of CI, or
// when running tests locally via `go test $(go list ./... | grep -v /vendor/)` or equivalent.
//
// Instead, it's meant to be run manually when developing or adjusting the SearchRepos
// implementation. To do so, temporarily comment out the t.Skip line (don't commit that!) and
// run it with `go test -v -run=TestSearchRepos ./services/backend`.
//
// SearchRepos currently uses GitHub search API as secondary source, and these test cases
// expect that it's accessible. You can set a client id and key for increased rate limit
// by setting the usual GITHUB_CLIENT_* environment variables before running the test.
func TestSearchRepos(t *testing.T) {
	t.Skip("SearchRepos is not meant to be run in CI")

	a, ctx := testserver.NewUnstartedServer()
	a.Config.Serve.NoWorker = true
	// TODO: Uncomment after https://github.com/sourcegraph/sourcegraph/pull/987 is merged.
	//for _, v := range os.Environ() {
	//	if strings.HasPrefix(v, "GITHUB_CLIENT_") { // Passthrough GitHub app credentials (if set) for authed API queries.
	//		a.Config.ExtraEnvConfig = append(a.Config.ExtraEnvConfig, v)
	//	}
	//}
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	for _, repoURI := range []string{
		"github.com/sourcegraph/srclib",
		"github.com/sourcegraph/srclib-go",

		"github.com/gorilla/mux",
	} {
		err := testutil.CreateEmptyMirrorRepo(t, ctx, repoURI)
		if err != nil {
			t.Fatal(err)
		}
	}

	cl, _ := sourcegraph.NewClientFromContext(ctx)

	tests := []struct {
		query string
		want  []string
	}{
		{"srclib", []string{"github.com/sourcegraph/srclib"}},
		{"srcli", []string{"github.com/sourcegraph/srclib", "github.com/sourcegraph/srclib-go"}},
		{"source src", []string{"github.com/sourcegraph/srclib", "github.com/sourcegraph/srclib-go"}},
		{"source/src", nil},
		{"sourcegraph/srclib", []string{"github.com/sourcegraph/srclib"}},
		{"sourcegraph/srcli", []string{"github.com/sourcegraph/srclib", "github.com/sourcegraph/srclib-go"}},
		{"github.com/sourcegraph/srclib", []string{"github.com/sourcegraph/srclib", "github.com/sourcegraph/srclib-go"}},
		{"github.com sourcegraph srclib", nil},

		{"mux", []string{"github.com/gorilla/mux"}},
		{"gorilla/mux", []string{"github.com/gorilla/mux"}},
		{"gorilla/", []string{"github.com/gorilla/mux"}},
		{"gorilla mux", []string{"github.com/gorilla/mux"}},
		{"gorilla/mu", []string{"github.com/gorilla/mux"}},
		{"gorilla/m", []string{"github.com/gorilla/mux"}},

		{"gorilla", []string{"github.com/gorilla/mux"}}, // Username only, not guarnateed to find mux repo.

		{"kubernetes", []string{"github.com/kubernetes/kubernetes"}},

		{"sqs/rego", []string{"github.com/sqs/rego"}},
		{"github.com/sqs/rego", []string{"github.com/sqs/rego"}},
	}
	for _, test := range tests {
		results, err := cl.Search.SearchRepos(ctx, &sourcegraph.SearchReposOp{Query: test.query})
		if err != nil {
			t.Fatal(err)
		}

		got := make(map[string]struct{}) // A set of repos we got (for easy querying).
		for _, repo := range results.Repos {
			got[repo.URI] = struct{}{}
		}
		for _, repo := range test.want {
			if _, ok := got[repo]; !ok {
				t.Errorf("%q: got repos %q, didn't get %q", test.query, got, repo)
			}
		}
	}
}
