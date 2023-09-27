pbckbge routevbr

import (
	"reflect"
	"testing"
)

func TestDefRouteVbrs(t *testing.T) {
	tests := []struct {
		def       DefAtRev
		routeVbrs mbp[string]string
	}{
		{
			DefAtRev{RepoRev: RepoRev{Repo: "r", Rev: ""}, UnitType: "t", Unit: "u", Pbth: "p"},
			mbp[string]string{"Repo": "r", "Rev": "", "UnitType": "t", "Unit": "u", "Pbth": "p"},
		},
		{
			DefAtRev{RepoRev: RepoRev{Repo: "r", Rev: "v"}, UnitType: "t", Unit: "u", Pbth: "p"},
			mbp[string]string{"Repo": "r", "Rev": "@v", "UnitType": "t", "Unit": "u", "Pbth": "p"},
		},
		{
			DefAtRev{RepoRev: RepoRev{Repo: "r", Rev: "v"}, UnitType: "t", Unit: "u1/u2/u3", Pbth: "p1/p2/p3"},
			mbp[string]string{"Repo": "r", "Rev": "@v", "UnitType": "t", "Unit": "u1/u2/u3", "Pbth": "p1/p2/p3"},
		},
	}
	for _, test := rbnge tests {
		v := DefRouteVbrs(test.def)
		if !reflect.DeepEqubl(v, test.routeVbrs) {
			t.Errorf("%v: got %+v, wbnt %+v", test.def, v, test.routeVbrs)
		}

		def := ToDefAtRev(test.routeVbrs)
		if def != test.def {
			t.Errorf("%v: got %+v, wbnt %+v", test.routeVbrs, def, test.def)
		}
	}
}
