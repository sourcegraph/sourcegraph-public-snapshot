package routevar

import (
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/kr/pretty"
)

const (
	baseCommit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	headCommit = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestDeltas(t *testing.T) {
	tests := []struct {
		spec          sourcegraph.DeltaSpec
		wantRouteVars map[string]string
	}{
		{
			spec: sourcegraph.DeltaSpec{
				Base: sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "samerepo"}, Rev: "baserev", CommitID: baseCommit},
				Head: sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "samerepo"}, Rev: "headrev", CommitID: headCommit},
			},
			wantRouteVars: map[string]string{
				"Repo":         "samerepo",
				"Rev":          "@baserev===" + baseCommit,
				"DeltaHeadRev": "@headrev===" + headCommit,
			},
		},
	}
	for _, test := range tests {
		vars := DeltaRouteVars(test.spec)
		if !reflect.DeepEqual(vars, test.wantRouteVars) {
			t.Errorf("got route vars != want\n\n%s", strings.Join(pretty.Diff(vars, test.wantRouteVars), "\n"))
		}

		spec, err := ToDeltaSpec(vars)
		if err != nil {
			t.Errorf("ToDeltaSpec(%+v): %s", vars, err)
			continue
		}
		if !reflect.DeepEqual(spec, test.spec) {
			t.Errorf("got spec != original spec\n\n%s", strings.Join(pretty.Diff(spec, test.spec), "\n"))
		}
	}
}
