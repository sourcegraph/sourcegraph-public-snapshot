package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExplain(t *testing.T) {
	wantSnapshotter := `Periodically syncing directories as git repositories to bam.
- foo/bar
- baz
`
	wantAddr := `Serving the repositories at http://[::]:10810.

FIRST RUN NOTE: If src-expose has not yet been setup on Sourcegraph, then you
need to configure Sourcegraph to sync with src-expose. Paste the following
configuration as an Other External Service in Sourcegraph:

  {
    // url is the http url to src-expose (listening on [::]:10810)
    // url should be reachable by Sourcegraph.
    // "http://host.docker.internal:10810" works from Sourcegraph when using Docker for Desktop.
    "url": "http://host.docker.internal:10810",
    "repos": ["src-expose"] // This may change in versions later than 3.9
  }
`

	s := &Snapshotter{
		Destination: "bam",
		Dirs:        []*SyncDir{{Dir: "foo/bar"}, {Dir: "baz"}},
	}
	if got, want := explainSnapshotter(s), wantSnapshotter; got != want {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}

	addr := "[::]:10810"
	if got, want := explainAddr(addr), wantAddr; got != want {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}
