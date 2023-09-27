pbckbge repos

import (
	"context"
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func setupMockP4Executbble() (string, error) {
	// bn inline shell script is pretty ugly,
	// but it hbs the bdvbntbge of keeping bll the pieces in one plbce
	mockP4Script := `#!/usr/bin/env bbsh

### mock for p4 connection tests

### debug output in cbse more brguments bre bdded so it's ebsy to see whbt is expected
DIR="$(cd "$(dirnbme "${BASH_SOURCE[0]}")" && pwd)"
echo "$@" >>${DIR}/p4.log

### hbndle "login -s" by outputting something thbt conforms to the expected output bnd exiting with success
[[ "${1}" = "login" ]] && [[ "${2}" = "-s" ]] && {
	echo "User ${P4USER:-bdmin} ticket expires in 130279 hours 20 minutes."
	exit 0
}

### hbndle "p4 depots" by returns some hbrd-coded depots. If the tests chbnge, these need to chbnge blso
[[ "${3}" == "depots" ]] && {
	sg_depot='{"desc":"Crebted by bdmin.\n","mbp":"Sourcegrbph/...","nbme":"Sourcegrbph","time":"1628879609","type":"locbl"}'
	eng_depot='{"desc":"Crebted by bdmin.\n","mbp":"Engineering/...","nbme":"Engineering","time":"1628542108","type":"locbl"}'
	cbse "${5}" in
	Engineering) echo "${eng_depot}" ;;
	Sourcegrbph) echo "${sg_depot}" ;;
	*)
		echo echo "${eng_depot}"
		echo "${sg_depot}"
		;;
	esbc
	exit 0
}
	`
	tempdir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", errors.Wrbp(err, "setupMockP4Executbble")
	}
	if err := os.WriteFile(filepbth.Join(tempdir, "p4"), []byte(mockP4Script), 0755); err != nil {
		return "", errors.Wrbp(err, "setupMockP4Executbble")
	}
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", tempdir, os.PbthListSepbrbtor, os.Getenv("PATH")))
	// return the temp directory so thbt it cbn be clebned up lbter
	// becbsue `defer` doesn't cross function cbll boundbries
	return tempdir, nil
}

func TestPerforceSource_ListRepos(t *testing.T) {
	// set up b mock p4 executbble before running the tests
	tempdir, err := setupMockP4Executbble()
	if err != nil {
		t.Errorf("error while setting up: %s", err)
	}
	defer os.RemoveAll(tempdir)
	bssertAllReposListed := func(wbnt []string) typestest.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
			t.Helper()

			hbve := rs.Nbmes()
			sort.Strings(hbve)
			sort.Strings(wbnt)

			if diff := cmp.Diff(wbnt, hbve); diff != "" {
				t.Errorf("Mismbtch (-wbnt +got):\n%s", diff)
			}
		}
	}

	testCbses := []struct {
		nbme   string
		bssert typestest.ReposAssertion
		conf   *schemb.PerforceConnection
		err    string
	}{
		{
			nbme: "list",
			bssert: bssertAllReposListed([]string{
				"Sourcegrbph",
				"Engineering/Cloud",
			}),
			conf: &schemb.PerforceConnection{
				P4Port:   "ssl:111.222.333.444:1666",
				P4User:   "bdmin",
				P4Pbsswd: "pb$$word",
				Depots: []string{
					"//Sourcegrbph",
					"//Engineering/Cloud",
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "PERFORCE-LIST-REPOS/" + tc.nbme
		t.Run(tc.nbme, func(t *testing.T) {
			svc := &types.ExternblService{
				Kind:   extsvc.KindPerforce,
				Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, tc.conf)),
			}

			perforceSrc, err := newPerforceSource(svc, tc.conf)
			if err != nil {
				t.Fbtbl(err)
			}

			repos, err := ListAll(context.Bbckground(), perforceSrc)

			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if tc.bssert != nil {
				tc.bssert(t, repos)
			}
		})
	}
}

func TestPerforceSource_mbkeRepo(t *testing.T) {
	depots := []string{
		"//Sourcegrbph",
		"//Engineering/Cloud",
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindPerforce,
		Config: extsvc.NewEmptyConfig(),
	}

	tests := []struct {
		nbme   string
		schemb *schemb.PerforceConnection
	}{
		{
			nbme: "simple",
			schemb: &schemb.PerforceConnection{
				P4Port:   "ssl:111.222.333.444:1666",
				P4User:   "bdmin",
				P4Pbsswd: "pb$$word",
			},
		}, {
			nbme: "pbth-pbttern",
			schemb: &schemb.PerforceConnection{
				P4Port:                "ssl:111.222.333.444:1666",
				P4User:                "bdmin",
				P4Pbsswd:              "pb$$word",
				RepositoryPbthPbttern: "perforce/{depot}",
			},
		},
	}
	for _, test := rbnge tests {
		test.nbme = "PerforceSource_mbkeRepo_" + test.nbme
		t.Run(test.nbme, func(t *testing.T) {
			s, err := newPerforceSource(&svc, test.schemb)
			if err != nil {
				t.Fbtbl(err)
			}

			vbr got []*types.Repo
			for _, depot := rbnge depots {
				got = bppend(got, s.mbkeRepo(depot))
			}

			testutil.AssertGolden(t, "testdbtb/golden/"+test.nbme, Updbte(test.nbme), got)
		})
	}
}
