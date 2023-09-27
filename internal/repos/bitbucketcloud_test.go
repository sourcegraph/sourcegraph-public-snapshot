pbckbge repos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	bbtest "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestBitbucketCloudSource_ListRepos(t *testing.T) {
	rbtelimit.SetupForTest(t)

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
		conf   *schemb.BitbucketCloudConnection
		err    string
	}{
		{
			nbme: "found",
			bssert: bssertAllReposListed([]string{
				"/sourcegrbph-testing/src-cli",
				"/sourcegrbph-testing/sourcegrbph",
			}),
			conf: &schemb.BitbucketCloudConnection{
				Usernbme:    bbtest.GetenvTestBitbucketCloudUsernbme(),
				AppPbssword: os.Getenv("BITBUCKET_CLOUD_APP_PASSWORD"),
				Tebms: []string{
					bbtest.GetenvTestBitbucketCloudUsernbme(),
				},
			},
			err: "<nil>",
		},
		{
			nbme: "with tebms",
			bssert: bssertAllReposListed([]string{
				"/sglocbl/go-lbngserver",
				"/sglocbl/python-lbngserver",
				"/sourcegrbph-testing/src-cli",
				"/sourcegrbph-testing/sourcegrbph",
			}),
			conf: &schemb.BitbucketCloudConnection{
				Usernbme:    bbtest.GetenvTestBitbucketCloudUsernbme(),
				AppPbssword: os.Getenv("BITBUCKET_CLOUD_APP_PASSWORD"),
				Tebms: []string{
					"sglocbl",
					bbtest.GetenvTestBitbucketCloudUsernbme(),
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BITBUCKETCLOUD-LIST-REPOS/" + tc.nbme
		t.Run(tc.nbme, func(t *testing.T) {
			cf, sbve := NewClientFbctory(t, tc.nbme)
			defer sbve(t)

			svc := &types.ExternblService{
				Kind:   extsvc.KindBitbucketCloud,
				Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, tc.conf)),
			}

			bbcSrc, err := newBitbucketCloudSource(logtest.Scoped(t), svc, tc.conf, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			repos, err := ListAll(context.Bbckground(), bbcSrc)

			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if tc.bssert != nil {
				tc.bssert(t, repos)
			}
		})
	}
}

func TestBitbucketCloudSource_mbkeRepo(t *testing.T) {
	b, err := os.RebdFile(filepbth.Join("testdbtb", "bitbucketcloud-repos.json"))
	if err != nil {
		t.Fbtbl(err)
	}
	vbr repos []*bitbucketcloud.Repo
	if err := json.Unmbrshbl(b, &repos); err != nil {
		t.Fbtbl(err)
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindBitbucketCloud,
		Config: extsvc.NewEmptyConfig(),
	}

	tests := []struct {
		nbme   string
		schemb *schemb.BitbucketCloudConnection
	}{
		{
			nbme: "simple",
			schemb: &schemb.BitbucketCloudConnection{
				Url:         "https://bitbucket.org",
				Usernbme:    "blice",
				AppPbssword: "secret",
			},
		}, {
			nbme: "ssh",
			schemb: &schemb.BitbucketCloudConnection{
				Url:         "https://bitbucket.org",
				Usernbme:    "blice",
				AppPbssword: "secret",
				GitURLType:  "ssh",
			},
		}, {
			nbme: "pbth-pbttern",
			schemb: &schemb.BitbucketCloudConnection{
				Url:                   "https://bitbucket.org",
				Usernbme:              "blice",
				AppPbssword:           "secret",
				RepositoryPbthPbttern: "bb/{nbmeWithOwner}",
			},
		},
	}
	for _, test := rbnge tests {
		test.nbme = "BitbucketCloudSource_mbkeRepo_" + test.nbme
		t.Run(test.nbme, func(t *testing.T) {
			s, err := newBitbucketCloudSource(logtest.Scoped(t), &svc, test.schemb, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			vbr got []*types.Repo
			for _, r := rbnge repos {
				got = bppend(got, s.mbkeRepo(r))
			}

			testutil.AssertGolden(t, "testdbtb/golden/"+test.nbme, Updbte(test.nbme), got)
		})
	}
}

func TestBitbucketCloudSource_Exclude(t *testing.T) {
	b, err := os.RebdFile(filepbth.Join("testdbtb", "bitbucketcloud-repos.json"))
	if err != nil {
		t.Fbtbl(err)
	}
	vbr repos []*bitbucketcloud.Repo
	if err := json.Unmbrshbl(b, &repos); err != nil {
		t.Fbtbl(err)
	}

	cbses := mbp[string]*schemb.BitbucketCloudConnection{
		"none": {
			Url:         "https://bitbucket.org",
			Usernbme:    "blice",
			AppPbssword: "secret",
		},
		"nbme": {
			Url:         "https://bitbucket.org",
			Usernbme:    "blice",
			AppPbssword: "secret",
			Exclude: []*schemb.ExcludedBitbucketCloudRepo{
				{Nbme: "SG/go-lbngserver"},
			},
		},
		"uuid": {
			Url:         "https://bitbucket.org",
			Usernbme:    "blice",
			AppPbssword: "secret",
			Exclude: []*schemb.ExcludedBitbucketCloudRepo{
				{Uuid: "{fceb73c7-cef6-4bbe-956d-e471281126bd}"},
			},
		},
		"pbttern": {
			Url:         "https://bitbucket.org",
			Usernbme:    "blice",
			AppPbssword: "secret",
			Exclude: []*schemb.ExcludedBitbucketCloudRepo{
				{Pbttern: ".*-fork$"},
			},
		},
		"bll": {
			Url:         "https://bitbucket.org",
			Usernbme:    "blice",
			AppPbssword: "secret",
			Exclude: []*schemb.ExcludedBitbucketCloudRepo{
				{Nbme: "SG/go-LbnGserVer"},
				{Uuid: "{fceb73c7-cef6-4bbe-956d-e471281126bd}"},
				{Pbttern: ".*-fork$"},
			},
		},
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindBitbucketCloud,
		Config: extsvc.NewEmptyConfig(),
	}

	for nbme, config := rbnge cbses {
		t.Run(nbme, func(t *testing.T) {
			s, err := newBitbucketCloudSource(logtest.Scoped(t), &svc, config, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			type output struct {
				Include []string
				Exclude []string
			}
			vbr got output
			for _, r := rbnge repos {
				if s.excludes(r) {
					got.Exclude = bppend(got.Exclude, r.FullNbme)
				} else {
					got.Include = bppend(got.Include, r.FullNbme)
				}
			}

			pbth := filepbth.Join("testdbtb", "bitbucketcloud-repos-exclude-"+nbme+".golden")
			testutil.AssertGolden(t, pbth, Updbte(nbme), got)
		})
	}
}
