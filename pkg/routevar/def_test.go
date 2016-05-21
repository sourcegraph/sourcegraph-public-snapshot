package routevar

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestDefRouteVars(t *testing.T) {
	tests := []struct {
		defSpec   sourcegraph.DefSpec
		routeVars map[string]string
	}{
		{
			sourcegraph.DefSpec{Repo: "r", CommitID: "", UnitType: "t", Unit: "u", Path: "p"},
			map[string]string{"Repo": "r", "Rev": "", "UnitType": "t", "Unit": "u", "Path": "p"},
		},
		{
			sourcegraph.DefSpec{Repo: "r", CommitID: "v", UnitType: "t", Unit: "u", Path: "p"},
			map[string]string{"Repo": "r", "Rev": "@v", "UnitType": "t", "Unit": "u", "Path": "p"},
		},
		{
			sourcegraph.DefSpec{Repo: "r", CommitID: "v", UnitType: "t", Unit: "u1/u2/u3", Path: "p1/p2/p3"},
			map[string]string{"Repo": "r", "Rev": "@v", "UnitType": "t", "Unit": "u1/u2/u3", "Path": "p1/p2/p3"},
		},
	}
	for _, test := range tests {
		v := DefRouteVars(test.defSpec)
		if !reflect.DeepEqual(v, test.routeVars) {
			t.Errorf("%v: got %+v, want %+v", test.defSpec, v, test.routeVars)
		}

		defSpec, err := ToDefSpec(test.routeVars)
		if err != nil {
			t.Errorf("%v: ToDefSpec: %s", test.routeVars, err)
			continue
		}
		if defSpec != test.defSpec {
			t.Errorf("%v: got %+v, want %+v", test.routeVars, defSpec, test.defSpec)
		}
	}
}
