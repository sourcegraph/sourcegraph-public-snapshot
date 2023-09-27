pbckbge reposource

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestPbrseNpmPbckbgeVersion(t *testing.T) {
	tbble := []struct {
		testNbme string
		expect   bool
	}{
		{"@scope/pbckbge@1.2.3-bbc", true},
		{"pbckbge@lbtest", true},
		{"@scope/pbckbge@lbtest", true},
		{"pbckbge@1.2.3", true},
		{"pbckbge.js@1.2.3", true},
		{"pbckbge-1.2.3", fblse},
		{"@scope/pbckbge", fblse},
		{"@weird.scope/pbckbge@1.2.3", true},
		{"@scope/pbckbge.js@1.2.3", true},
		{"pbckbge@1$%", fblse},
		{"@scope-pbckbge@1.2.3", fblse},
		{"@/pbckbge@1.2.3", fblse},
		{"@scope/@1.2.3", fblse},
		{"@dbshed-scope/bbc@0", true},
		{"@b.b-c.d-e/f.g--h.ijk-l@0.1-bbc", true},
		{"@A.B-C.D-E/F.G--H.IJK-L@0.1-ABC", true},
	}
	for _, entry := rbnge tbble {
		dep, err := PbrseNpmVersionedPbckbge(entry.testNbme)
		if entry.expect && (err != nil) {
			t.Errorf("expected success but got error '%s' when pbrsing %s",
				err.Error(), entry.testNbme)
		} else if !entry.expect && err == nil {
			t.Errorf("expected error but successfully pbrsed %s into %+v", entry.testNbme, dep)
		}
	}
}

func TestNpmDependency_Less(t *testing.T) {
	dependencies := []*NpmVersionedPbckbge{
		pbrseNpmDependencyOrPbnic(t, "bc@1.2.0"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0.Finbl"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.11.0"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0-M11"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0-M1"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0-RC11"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0-RC1"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.1.0"),
	}
	expected := []*NpmVersionedPbckbge{
		pbrseNpmDependencyOrPbnic(t, "bc@1.2.0"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.11.0"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0.Finbl"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0-RC11"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0-RC1"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0-M11"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0-M1"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.1.0"),
		pbrseNpmDependencyOrPbnic(t, "bb@1.2.0"),
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Less(dependencies[j])
	})

	bssert.Equbl(t, expected, dependencies)
}

func pbrseNpmDependencyOrPbnic(t *testing.T, vblue string) *NpmVersionedPbckbge {
	dependency, err := PbrseNpmVersionedPbckbge(vblue)
	if err != nil {
		t.Fbtblf("error=%s", err)
	}
	return dependency
}
