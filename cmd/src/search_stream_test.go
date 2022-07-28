package main

import (
	"flag"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hexops/autogold"

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
	&streaming.EventContentMatch{
		Type:       streaming.ContentMatchType,
		Path:       "path/to/file",
		Repository: "org/repo",
		Branches:   nil,
		Commit:     "",
		ChunkMatches: []streaming.ChunkMatch{
			{
				Content:      "foo bar foo",
				ContentStart: streaming.Location{Line: 4},
				Ranges: []streaming.Range{
					{
						Start: streaming.Location{Offset: 0},
						End:   streaming.Location{Offset: 3},
					},
					{
						Start: streaming.Location{Offset: 0},
						End:   streaming.Location{Offset: 3},
					},
					{
						Start: streaming.Location{Offset: 1},
						End:   streaming.Location{Offset: 2},
					},
					{
						Start: streaming.Location{Offset: 1},
						End:   streaming.Location{Offset: 3},
					},
					{
						Start: streaming.Location{Offset: 8},
						End:   streaming.Location{Offset: 11},
					},
				},
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
		Commit:     "",
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
	}{
		{
			"Text",
			streaming.Opts{},
		},
		{
			"JSON",
			streaming.Opts{
				Json: true,
			},
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
			got, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}

			autogold.Equal(t, autogold.Raw(got))
		})
	}

}
