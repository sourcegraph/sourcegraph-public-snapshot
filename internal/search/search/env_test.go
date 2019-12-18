package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestZoektAddr(t *testing.T) {
	cases := []struct {
		name    string
		environ []string
		want    string
	}{{
		name: "default",
		want: "k8s+rpc://indexed-search:6070",
	}, {
		name:    "old",
		environ: []string{"ZOEKT_HOST=127.0.0.1:3070"},
		want:    "127.0.0.1:3070",
	}, {
		name:    "new",
		environ: []string{"INDEXED_SEARCH_SERVERS=indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070"},
		want:    "indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070",
	}, {
		name: "prefer new",
		environ: []string{
			"ZOEKT_HOST=127.0.0.1:3070",
			"INDEXED_SEARCH_SERVERS=indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070",
		},
		want: "indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070",
	}, {
		name: "unset new",
		environ: []string{
			"ZOEKT_HOST=127.0.0.1:3070",
			"INDEXED_SEARCH_SERVERS=",
		},
		want: "",
	}, {
		name: "unset old",
		environ: []string{
			"ZOEKT_HOST=",
		},
		want: "",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := zoektAddr(tc.environ)
			if got != tc.want {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
