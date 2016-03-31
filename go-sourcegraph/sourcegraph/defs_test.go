package sourcegraph

import (
	"reflect"
	"testing"
)

var defRouteVarsTests = []struct {
	defSpec   DefSpec
	routeVars map[string]string
}{
	{
		DefSpec{Repo: "r", CommitID: "", UnitType: "t", Unit: "u", Path: "p"},
		map[string]string{"Repo": "r", "Rev": "", "UnitType": "t", "Unit": "u", "Path": "p"},
	},
	{
		DefSpec{Repo: "r", CommitID: "v", UnitType: "t", Unit: "u", Path: "p"},
		map[string]string{"Repo": "r", "Rev": "@v", "UnitType": "t", "Unit": "u", "Path": "p"},
	},
	{
		DefSpec{Repo: "r", CommitID: "v", UnitType: "t", Unit: "u1/u2/u3", Path: "p1/p2/p3"},
		map[string]string{"Repo": "r", "Rev": "@v", "UnitType": "t", "Unit": "u1/u2/u3", "Path": "p1/p2/p3"},
	},
}

func TestDefRouteVars(t *testing.T) {
	for _, test := range defRouteVarsTests {
		v := test.defSpec.RouteVars()
		if !reflect.DeepEqual(v, test.routeVars) {
			t.Errorf("%v: got %+v, want %+v", test.defSpec, v, test.routeVars)
		}
	}
}

func TestUnmarshalDefSpec(t *testing.T) {
	for _, test := range defRouteVarsTests {
		defSpec, err := UnmarshalDefSpec(test.routeVars)
		if err != nil {
			t.Errorf("%v: UnmarshalDefSpec: %s", test.routeVars, err)
			continue
		}
		if defSpec != test.defSpec {
			t.Errorf("%v: got %+v, want %+v", test.routeVars, defSpec, test.defSpec)
		}
	}
}
