package sourcegraph

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kr/pretty"
)

const (
	baseCommit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	headCommit = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestDeltas(t *testing.T) {
	tests := []struct {
		spec          DeltaSpec
		wantRouteVars map[string]string
	}{
		{
			spec: DeltaSpec{
				Base: RepoRevSpec{RepoSpec: RepoSpec{URI: "samerepo"}, Rev: "baserev", CommitID: baseCommit},
				Head: RepoRevSpec{RepoSpec: RepoSpec{URI: "samerepo"}, Rev: "headrev", CommitID: headCommit},
			},
			wantRouteVars: map[string]string{
				"Repo":                 "samerepo",
				"Rev":                  "baserev",
				"CommitID":             baseCommit,
				"DeltaHeadResolvedRev": "headrev===" + headCommit,
			},
		},
	}
	for _, test := range tests {
		vars := test.spec.RouteVars()
		if !reflect.DeepEqual(vars, test.wantRouteVars) {
			t.Errorf("got route vars != want\n\n%s", strings.Join(pretty.Diff(vars, test.wantRouteVars), "\n"))
		}

		spec, err := UnmarshalDeltaSpec(vars)
		if err != nil {
			t.Errorf("UnmarshalDeltaSpec(%+v): %s", vars, err)
			continue
		}
		if !reflect.DeepEqual(spec, test.spec) {
			t.Errorf("got spec != original spec\n\n%s", strings.Join(pretty.Diff(spec, test.spec), "\n"))
		}
	}
}
