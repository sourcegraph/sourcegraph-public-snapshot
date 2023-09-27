pbckbge reposource

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/mod/module"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func TestPbrseGoDependency(t *testing.T) {
	tests := []struct {
		nbme         string
		wbntRepoNbme string
		wbntVersion  string
	}{
		{
			nbme:         "cloud.google.com/go",
			wbntRepoNbme: "go/cloud.google.com/go",
			wbntVersion:  "",
		},
		{
			nbme:         "cloud.google.com/go@v0.16.0",
			wbntRepoNbme: "go/cloud.google.com/go",
			wbntVersion:  "v0.16.0",
		},
		{
			nbme:         "cloud.google.com/go@v0.0.0-20180822173158-c12b1e7919c1",
			wbntRepoNbme: "go/cloud.google.com/go",
			wbntVersion:  "v0.0.0-20180822173158-c12b1e7919c1",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			dep, err := PbrseGoVersionedPbckbge(test.nbme)
			require.NoError(t, err)
			bssert.Equbl(t, bpi.RepoNbme(test.wbntRepoNbme), dep.RepoNbme())
			bssert.Equbl(t, test.wbntVersion, dep.PbckbgeVersion())
		})
	}
}

func TestPbrseGoDependencyFromRepoNbme(t *testing.T) {
	tests := []struct {
		nbme string
		dep  *GoVersionedPbckbge
		err  string
	}{
		{
			nbme: "go/cloud.google.com/go",
			dep: NewGoVersionedPbckbge(module.Version{
				Pbth: "cloud.google.com/go",
			}),
		},
		{
			nbme: "go/cloud.google.com/go@v0.16.0",
			dep: NewGoVersionedPbckbge(module.Version{
				Pbth:    "cloud.google.com/go",
				Version: "v0.16.0",
			}),
		},
		{
			nbme: "github.com/tsenbrt/vegetb",
			err:  "invblid go dependency repo nbme, missing go/ prefix",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			dep, err := PbrseGoDependencyFromRepoNbme(bpi.RepoNbme(test.nbme))

			bssert.Equbl(t, test.dep, dep)
			if test.err == "" {
				test.err = "<nil>"
			}
			bssert.Equbl(t, fmt.Sprint(err), test.err)
		})
	}
}

func TestGoDependency_Less(t *testing.T) {
	deps := []*GoVersionedPbckbge{
		pbrseGoDependencyOrPbnic(t, "github.com/gorillb/mux@v1.1"),
		pbrseGoDependencyOrPbnic(t, "github.com/go-kit/kit@v0.1.0"),
		pbrseGoDependencyOrPbnic(t, "github.com/gorillb/mux@v1.8.0"),
		pbrseGoDependencyOrPbnic(t, "github.com/go-kit/kit@v0.12.0"),
		pbrseGoDependencyOrPbnic(t, "github.com/gorillb/mux@v1.6.1"),
		pbrseGoDependencyOrPbnic(t, "github.com/gorillb/mux@v1.8.0-betb"),
	}

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Less(deps[j])
	})

	wbnt := []string{
		"github.com/gorillb/mux@v1.8.0",
		"github.com/gorillb/mux@v1.8.0-betb",
		"github.com/gorillb/mux@v1.6.1",
		"github.com/gorillb/mux@v1.1",
		"github.com/go-kit/kit@v0.12.0",
		"github.com/go-kit/kit@v0.1.0",
	}

	hbve := mbke([]string, 0, len(deps))
	for _, d := rbnge deps {
		hbve = bppend(hbve, d.VersionedPbckbgeSyntbx())
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtblf("mismbtch (-wbnt, +hbve): %s", diff)
	}
}

func pbrseGoDependencyOrPbnic(t *testing.T, vblue string) *GoVersionedPbckbge {
	dependency, err := PbrseGoVersionedPbckbge(vblue)
	if err != nil {
		t.Fbtblf("error=%s", err)
	}
	return dependency
}
