package main

import (
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/streaming"
)

func mockStreamHandler(w http.ResponseWriter, _ *http.Request) {
	writer, _ := streaming.NewWriter(w)
	writer.Event("matches", event)
	writer.Event("done", nil)
}

func testServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	// We need a stable port, because src-cli output contains references to the host.
	// Here we exchange the standard listener with our own.
	l, err := net.Listen("tcp", "127.0.0.1:55128")
	if err != nil {
		t.Fatal(err)
	}
	s := httptest.NewUnstartedServer(handler)
	s.Listener.Close()
	s.Listener = l
	s.Start()
	return s
}

var event = []streaming.EventMatch{
	&streaming.EventFileMatch{
		Type:       streaming.FileMatchType,
		Path:       "path/to/file",
		Repository: "org/repo",
		Branches:   nil,
		Version:    "",
		LineMatches: []streaming.EventLineMatch{
			{
				Line:             "foo bar",
				LineNumber:       4,
				OffsetAndLengths: [][2]int32{{4, 3}},
			},
		},
	},
	&streaming.EventRepoMatch{
		Type:       streaming.RepoMatchType,
		Repository: "sourcegraph/sourcegraph",
		Branches:   []string{},
	},
	&streaming.EventSymbolMatch{
		Type:       streaming.SymbolMatchType,
		Path:       "path/to/file",
		Repository: "org/repo",
		Branches:   []string{},
		Version:    "",
		Symbols: []streaming.Symbol{
			{
				URL:           "github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/graphqlbackend/search_results.go#L1591:26-1591:35",
				Name:          "doResults",
				ContainerName: "",
				Kind:          "FUNCTION",
			},
			{
				URL:           "github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/graphqlbackend/search_results.go#L1591:26-1591:35",
				Name:          "Results",
				ContainerName: "SearchResultsResolver",
				Kind:          "FIELD",
			},
		},
	},
	&streaming.EventCommitMatch{
		Type:    streaming.CommitMatchType,
		Icon:    "",
		Label:   "[sourcegraph/sourcegraph-atom](/github.com/sourcegraph/sourcegraph-atom) › [Stephen Gutekanst](/github.com/sourcegraph/sourcegraph-atom/-/commit/5b098d7fed963d88e23057ed99d73d3c7a33ad89): [all: release v1.0.5](/github.com/sourcegraph/sourcegraph-atom/-/commit/5b098d7fed963d88e23057ed99d73d3c7a33ad89)^",
		URL:     "",
		Detail:  "",
		Content: "```COMMIT_EDITMSG\nfoo bar\n```",
		Ranges: [][3]int32{
			{1, 3, 3},
		},
	},
	&streaming.EventCommitMatch{
		Type:    streaming.CommitMatchType,
		Icon:    "",
		Label:   "[sourcegraph/sourcegraph-atom](/github.com/sourcegraph/sourcegraph-atom) › [Stephen Gutekanst](/github.com/sourcegraph/sourcegraph-atom/-/commit/5b098d7fed963d88e23057ed99d73d3c7a33ad89): [all: release v1.0.5](/github.com/sourcegraph/sourcegraph-atom/-/commit/5b098d7fed963d88e23057ed99d73d3c7a33ad89)^",
		URL:     "",
		Detail:  "",
		Content: "```diff\nsrc/data.ts src/data.ts\n@@ -0,0 +11,4 @@\n+    return of<Data>({\n+        title: 'Acme Corp open-source code search',\n+        summary: 'Instant code search across all Acme Corp open-source code.',\n+        githubOrgs: ['sourcegraph'],\n```",
		Ranges: [][3]int32{
			{4, 44, 6},
		},
	},
}

func TestSearchStream(t *testing.T) {
	s := testServer(t, http.HandlerFunc(mockStreamHandler))
	defer s.Close()

	cfg = &config{
		Endpoint: s.URL,
	}
	defer func() { cfg = nil }()

	cases := []struct {
		name string
		opts streaming.Opts
		want string
	}{
		{
			"Text",
			streaming.Opts{},
			"./testdata/streaming_search_want.txt",
		},
		{
			"JSON",
			streaming.Opts{
				Json: true,
			},
			"./testdata/streaming_search_want.json",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Capture output.
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}

			flagSet := flag.NewFlagSet("test", flag.ExitOnError)
			flags := api.NewFlags(flagSet)
			client := cfg.apiClient(flags, flagSet.Output())
			err = streamSearch("", c.opts, client, w)
			if err != nil {
				t.Fatal(err)
			}
			err = w.Close()
			if err != nil {
				t.Fatal(err)
			}
			got, err := ioutil.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			want, err := ioutil.ReadFile(c.want)
			if err != nil {
				t.Fatal(err)
			}
			if d := cmp.Diff(want, got); d != "" {
				t.Fatalf("(-want +got): %s", d)
			}
		})
	}

}
