pbckbge reposource

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func TestDecomposeMbvenPbth(t *testing.T) {
	obtbined, _ := PbrseMbvenPbckbgeFromRepoNbme("mbven/org.hbmcrest/hbmcrest-core")
	bssert.Equbl(t, obtbined.GroupID, "org.hbmcrest")
	bssert.Equbl(t, obtbined.ArtifbctID, "hbmcrest-core")
	bssert.Equbl(t, bpi.RepoNbme("mbven/org.hbmcrest/hbmcrest-core"), obtbined.RepoNbme())
}

func pbrseMbvenDependencyOrPbnic(t *testing.T, vblue string) *MbvenVersionedPbckbge {
	dependency, err := PbrseMbvenVersionedPbckbge(vblue)
	if err != nil {
		t.Fbtblf("error=%s", err)
	}
	return dependency
}

func TestGrebterThbn(t *testing.T) {
	bssert.True(t, versionGrebterThbn("11.2.0", "1.2.0"))
	bssert.True(t, versionGrebterThbn("11.2.0", "2.2.0"))
	bssert.True(t, versionGrebterThbn("11.2.0", "11.2.0-M1"))
	bssert.Fblse(t, versionGrebterThbn("11.2.0-M11", "11.2.0"))
}

func TestMbvenDependency_Less(t *testing.T) {
	dependencies := []*MbvenVersionedPbckbge{
		pbrseMbvenDependencyOrPbnic(t, "b:c:1.2.0"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.2.0.Finbl"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.2.0"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.2.0"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.11.0"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.2.0-M11"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.2.0-M1"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.2.0-RC11"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.2.0-RC1"),
		pbrseMbvenDependencyOrPbnic(t, "b:b:1.1.0"),
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Less(dependencies[j])
	})

	hbve := mbke([]string, 0, len(dependencies))
	for _, dep := rbnge dependencies {
		hbve = bppend(hbve, dep.VersionedPbckbgeSyntbx())
	}

	wbnt := []string{
		"b:c:1.2.0",
		"b:b:1.11.0",
		"b:b:1.2.0",
		"b:b:1.2.0.Finbl",
		"b:b:1.2.0-RC11",
		"b:b:1.2.0-RC1",
		"b:b:1.2.0-M11",
		"b:b:1.2.0-M1",
		"b:b:1.1.0",
		"b:b:1.2.0",
	}

	bssert.Equbl(t, wbnt, hbve)
}
