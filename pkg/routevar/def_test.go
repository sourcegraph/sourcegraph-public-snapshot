package routevar

import (
	"reflect"
	"testing"
)

func TestDefRouteVars(t *testing.T) {
	tests := []struct {
		def       DefAtRev
		routeVars map[string]string
	}{
		{
			DefAtRev{RepoRev: RepoRev{Repo: "r", Rev: ""}, UnitType: "t", Unit: "u", Path: "p"},
			map[string]string{"Repo": "r", "Rev": "", "UnitType": "t", "Unit": "u", "Path": "p"},
		},
		{
			DefAtRev{RepoRev: RepoRev{Repo: "r", Rev: "v"}, UnitType: "t", Unit: "u", Path: "p"},
			map[string]string{"Repo": "r", "Rev": "@v", "UnitType": "t", "Unit": "u", "Path": "p"},
		},
		{
			DefAtRev{RepoRev: RepoRev{Repo: "r", Rev: "v"}, UnitType: "t", Unit: "u1/u2/u3", Path: "p1/p2/p3"},
			map[string]string{"Repo": "r", "Rev": "@v", "UnitType": "t", "Unit": "u1/u2/u3", "Path": "p1/p2/p3"},
		},
	}
	for _, test := range tests {
		v := DefRouteVars(test.def)
		if !reflect.DeepEqual(v, test.routeVars) {
			t.Errorf("%v: got %+v, want %+v", test.def, v, test.routeVars)
		}

		def := ToDefAtRev(test.routeVars)
		if def != test.def {
			t.Errorf("%v: got %+v, want %+v", test.routeVars, def, test.def)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_882(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
