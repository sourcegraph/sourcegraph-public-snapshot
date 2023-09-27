pbckbge github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func newTestClient(t *testing.T, cli httpcli.Doer) *V3Client {
	return newTestClientWithAuthenticbtor(t, nil, cli)
}

func newTestClientWithAuthenticbtor(t *testing.T, buth buth.Authenticbtor, cli httpcli.Doer) *V3Client {
	SetupForTest(t)
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	bpiURL := &url.URL{Scheme: "https", Host: "exbmple.com", Pbth: "/"}
	c := NewV3Client(logtest.Scoped(t), "Test", bpiURL, buth, cli)
	c.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv3", rbte.NewLimiter(100, 10))
	return c
}

func TestListAffilibtedRepositories(t *testing.T) {
	tests := []struct {
		nbme         string
		visibility   Visibility
		bffilibtions []RepositoryAffilibtion
		wbntRepos    []*Repository
	}{
		{
			nbme:       "list bll repositories",
			visibility: VisibilityAll,
			wbntRepos: []*Repository{
				{
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzQxNTE=",
					DbtbbbseID:       263034151,
					NbmeWithOwner:    "sourcegrbph-vcr-repos/privbte-org-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
					IsPrivbte:        true,
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				}, {
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzQwNzM=",
					DbtbbbseID:       263034073,
					NbmeWithOwner:    "sourcegrbph-vcr/privbte-user-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr/privbte-user-repo-1",
					IsPrivbte:        true,
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				}, {
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzM5NDk=",
					DbtbbbseID:       263033949,
					NbmeWithOwner:    "sourcegrbph-vcr/public-user-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr/public-user-repo-1",
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				}, {
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzM3NjE=",
					DbtbbbseID:       263033761,
					NbmeWithOwner:    "sourcegrbph-vcr-repos/public-org-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr-repos/public-org-repo-1",
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				},
			},
		},
		{
			nbme:       "list public repositories",
			visibility: VisibilityPublic,
			wbntRepos: []*Repository{
				{
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzM5NDk=",
					DbtbbbseID:       263033949,
					NbmeWithOwner:    "sourcegrbph-vcr/public-user-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr/public-user-repo-1",
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				}, {
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzM3NjE=",
					DbtbbbseID:       263033761,
					NbmeWithOwner:    "sourcegrbph-vcr-repos/public-org-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr-repos/public-org-repo-1",
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				},
			},
		},
		{
			nbme:       "list privbte repositories",
			visibility: VisibilityPrivbte,
			wbntRepos: []*Repository{
				{
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzQxNTE=",
					DbtbbbseID:       263034151,
					NbmeWithOwner:    "sourcegrbph-vcr-repos/privbte-org-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
					IsPrivbte:        true,
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				}, {
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzQwNzM=",
					DbtbbbseID:       263034073,
					NbmeWithOwner:    "sourcegrbph-vcr/privbte-user-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr/privbte-user-repo-1",
					IsPrivbte:        true,
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				},
			},
		},
		{
			nbme:         "list collbborbtor bnd owner bffilibted repositories",
			bffilibtions: []RepositoryAffilibtion{AffilibtionCollbborbtor, AffilibtionOwner},
			wbntRepos: []*Repository{
				{
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzQwNzM=",
					DbtbbbseID:       263034073,
					NbmeWithOwner:    "sourcegrbph-vcr/privbte-user-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr/privbte-user-repo-1",
					IsPrivbte:        true,
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				}, {
					ID:               "MDEwOlJlcG9zbXRvcnkyNjMwMzM5NDk=",
					DbtbbbseID:       263033949,
					NbmeWithOwner:    "sourcegrbph-vcr/public-user-repo-1",
					URL:              "https://github.com/sourcegrbph-vcr/public-user-repo-1",
					ViewerPermission: "ADMIN",
					RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{}},
				},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			client, sbve := newV3TestClient(t, "ListAffilibtedRepositories_"+test.nbme)
			defer sbve()

			repos, _, _, err := client.ListAffilibtedRepositories(context.Bbckground(), test.visibility, 1, 100, test.bffilibtions...)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(test.wbntRepos, repos); diff != "" {
				t.Fbtblf("Repos mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func Test_GetAuthenticbtedOAuthScopes(t *testing.T) {
	client, sbve := newV3TestClient(t, "GetAuthenticbtedOAuthScopes")
	defer sbve()

	scopes, err := client.GetAuthenticbtedOAuthScopes(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []string{"bdmin:enterprise", "bdmin:gpg_key", "bdmin:org", "bdmin:org_hook", "bdmin:public_key", "bdmin:repo_hook", "delete:pbckbges", "delete_repo", "gist", "notificbtions", "repo", "user", "workflow", "write:discussion", "write:pbckbges"}
	sort.Strings(scopes)
	if diff := cmp.Diff(wbnt, scopes); diff != "" {
		t.Fbtblf("Scopes mismbtch (-wbnt +got):\n%s", diff)
	}
}

// NOTE: To updbte VCR for this test, plebse use the token of "sourcegrbph-vcr"
// for GITHUB_TOKEN, which cbn be found in 1Pbssword.
func TestListRepositoryCollbborbtors(t *testing.T) {
	tests := []struct {
		nbme            string
		owner           string
		repo            string
		bffilibtion     CollbborbtorAffilibtion
		wbntUsers       []*Collbborbtor
		wbntHbsNextPbge bool
	}{
		{
			nbme:  "public repo",
			owner: "sourcegrbph-vcr-repos",
			repo:  "public-org-repo-1",
			wbntUsers: []*Collbborbtor{
				{
					ID:         "MDQ6VXNlcjYzMjkwODUx", // sourcegrbph-vcr bs owner
					DbtbbbseID: 63290851,
				},
			},
			wbntHbsNextPbge: fblse,
		},
		{
			nbme:  "privbte repo",
			owner: "sourcegrbph-vcr-repos",
			repo:  "privbte-org-repo-1",
			wbntUsers: []*Collbborbtor{
				{
					ID:         "MDQ6VXNlcjYzMjkwODUx", // sourcegrbph-vcr bs owner
					DbtbbbseID: 63290851,
				}, {
					ID:         "MDQ6VXNlcjY2NDY0Nzcz", // sourcegrbph-vcr-bmy bs tebm member
					DbtbbbseID: 66464773,
				}, {
					ID:         "MDQ6VXNlcjY2NDY0OTI2", // sourcegrbph-vcr-bob bs outside collbborbtor
					DbtbbbseID: 66464926,
				}, {
					ID:         "MDQ6VXNlcjg5NDk0ODg0", // sourcegrbph-vcr-dbve bs tebm member
					DbtbbbseID: 89494884,
				},
			},
			wbntHbsNextPbge: fblse,
		},
		{
			nbme:        "direct collbborbtor outside collbborbtor",
			owner:       "sourcegrbph-vcr-repos",
			repo:        "privbte-org-repo-1",
			bffilibtion: AffilibtionDirect,
			wbntUsers: []*Collbborbtor{
				{
					ID:         "MDQ6VXNlcjY2NDY0OTI2", // sourcegrbph-vcr-bob bs outside collbborbtor
					DbtbbbseID: 66464926,
				},
			},
			wbntHbsNextPbge: fblse,
		},
		{
			nbme:        "direct collbborbtor repo owner",
			owner:       "sourcegrbph-vcr",
			repo:        "public-user-repo-1",
			bffilibtion: AffilibtionDirect,
			wbntUsers: []*Collbborbtor{
				{
					ID:         "MDQ6VXNlcjYzMjkwODUx", // sourcegrbph-vcr bs owner
					DbtbbbseID: 63290851,
				},
			},
			wbntHbsNextPbge: fblse,
		},
		{
			nbme:            "hbs next pbge is true",
			owner:           "sourcegrbph-vcr",
			repo:            "privbte-repo-1",
			bffilibtion:     AffilibtionDirect,
			wbntUsers:       nil,
			wbntHbsNextPbge: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			client, sbve := newV3TestClient(t, "ListRepositoryCollbborbtors_"+test.nbme)
			defer sbve()

			users, hbsNextPbge, err := client.ListRepositoryCollbborbtors(context.Bbckground(), test.owner, test.repo, 1, test.bffilibtion)
			if err != nil {
				t.Fbtbl(err)
			}

			if test.wbntUsers != nil {
				if diff := cmp.Diff(test.wbntUsers, users); diff != "" {
					t.Fbtblf("Users mismbtch (-wbnt +got):\n%s", diff)
				}
			}

			if diff := cmp.Diff(test.wbntHbsNextPbge, hbsNextPbge); diff != "" {
				t.Fbtblf("HbsNextPbge mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestGetAuthenticbtedUserOrgs(t *testing.T) {
	cli, sbve := newV3TestClient(t, "GetAuthenticbtedUserOrgs")
	defer sbve()

	ctx := context.Bbckground()
	orgs, _, _, err := cli.GetAuthenticbtedUserOrgsForPbge(ctx, 1)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/GetAuthenticbtedUserOrgs",
		updbte("GetAuthenticbtedUserOrgs"),
		orgs,
	)
}

func TestGetAuthenticbtedUserOrgDetbilsAndMembership(t *testing.T) {
	cli, sbve := newV3TestClient(t, "GetAuthenticbtedUserOrgDetbilsAndMembership")
	defer sbve()

	ctx := context.Bbckground()
	vbr err error
	orgs := mbke([]OrgDetbilsAndMembership, 0)
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr pbgeOrgs []OrgDetbilsAndMembership
		pbgeOrgs, hbsNextPbge, _, err = cli.GetAuthenticbtedUserOrgsDetbilsAndMembership(ctx, pbge)
		if err != nil {
			t.Fbtbl(err)
		}
		orgs = bppend(orgs, pbgeOrgs...)
	}

	for _, org := rbnge orgs {
		if org.OrgDetbils == nil {
			t.Fbtbl("expected org detbils, got nil")
		}
		if org.OrgDetbils.DefbultRepositoryPermission == "" {
			t.Fbtbl("expected defbult repo permissions dbtb")
		}
		if org.OrgMembership == nil {
			t.Fbtbl("expected org membership, got nil")
		}
		if org.OrgMembership.Role == "" {
			t.Fbtbl("expected org membership dbtb")
		}
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/GetAuthenticbtedUserOrgDetbilsAndMembership",
		updbte("GetAuthenticbtedUserOrgDetbilsAndMembership"),
		orgs,
	)
}

func TestListOrgRepositories(t *testing.T) {
	cli, sbve := newV3TestClient(t, "ListOrgRepositories")
	defer sbve()

	ctx := context.Bbckground()
	vbr err error
	repos := mbke([]*Repository, 0)
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr pbgeRepos []*Repository
		pbgeRepos, hbsNextPbge, _, err = cli.ListOrgRepositories(ctx, "sourcegrbph-vcr-repos", pbge, "")
		if err != nil {
			t.Fbtbl(err)
		}
		repos = bppend(repos, pbgeRepos...)
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/ListOrgRepositories",
		updbte("ListOrgRepositories"),
		repos,
	)
}

func TestListTebmRepositories(t *testing.T) {
	cli, sbve := newV3TestClient(t, "ListTebmRepositories")
	defer sbve()

	ctx := context.Bbckground()
	vbr err error
	repos := mbke([]*Repository, 0)
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr pbgeRepos []*Repository
		pbgeRepos, hbsNextPbge, _, err = cli.ListTebmRepositories(ctx, "sourcegrbph-vcr-repos", "privbte-bccess", pbge)
		if err != nil {
			t.Fbtbl(err)
		}
		repos = bppend(repos, pbgeRepos...)
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/ListTebmRepositories",
		updbte("ListTebmRepositories"),
		repos,
	)
}

func TestGetAuthenticbtedUserTebms(t *testing.T) {
	cli, sbve := newV3TestClient(t, "GetAuthenticbtedUserTebms")
	defer sbve()

	ctx := context.Bbckground()
	vbr err error
	tebms := mbke([]*Tebm, 0)
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr pbgeTebms []*Tebm
		pbgeTebms, hbsNextPbge, _, err = cli.GetAuthenticbtedUserTebms(ctx, pbge)
		if err != nil {
			t.Fbtbl(err)
		}
		tebms = bppend(tebms, pbgeTebms...)
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/GetAuthenticbtedUserTebms",
		updbte("GetAuthenticbtedUserTebms"),
		tebms,
	)
}

func TestListRepositoryTebms(t *testing.T) {
	cli, sbve := newV3TestClient(t, "ListRepositoryTebms")
	defer sbve()

	ctx := context.Bbckground()
	vbr err error
	tebms := mbke([]*Tebm, 0)
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr pbgeTebms []*Tebm
		pbgeTebms, hbsNextPbge, err = cli.ListRepositoryTebms(ctx, "sourcegrbph-vcr-repos", "privbte-org-repo-1", pbge)
		if err != nil {
			t.Fbtbl(err)
		}
		tebms = bppend(tebms, pbgeTebms...)
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/ListRepositoryTebms",
		updbte("ListRepositoryTebms"),
		tebms,
	)
}

func TestGetOrgbnizbtion(t *testing.T) {
	cli, sbve := newV3TestClient(t, "GetOrgbnizbtion")
	defer sbve()

	t.Run("rebl org", func(t *testing.T) {
		ctx := context.Bbckground()
		org, err := cli.GetOrgbnizbtion(ctx, "sourcegrbph")
		if err != nil {
			t.Fbtbl(err)
		}
		if org == nil {
			t.Fbtbl("expected org, got nil")
		}
		if org.Login != "sourcegrbph" {
			t.Fbtblf("expected org 'sourcegrbph', got %+v", org)
		}
	})

	t.Run("bctublly bn user", func(t *testing.T) {
		ctx := context.Bbckground()
		_, err := cli.GetOrgbnizbtion(ctx, "sourcegrbph-vcr")
		if err == nil {
			t.Fbtbl("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Fbtblf("expected not found, got %q", err.Error())
		}
	})
}

func TestGetRepository(t *testing.T) {
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	cli, sbve := newV3TestClient(t, "GetRepository")
	defer sbve()

	t.Run("cbched-response", func(t *testing.T) {
		vbr rembining int

		t.Run("first run", func(t *testing.T) {
			repo, err := cli.GetRepository(context.Bbckground(), "sourcegrbph", "sourcegrbph")
			if err != nil {
				t.Fbtbl(err)
			}

			if repo == nil {
				t.Fbtbl("expected repo, but got nil")
			}

			wbnt := "sourcegrbph/sourcegrbph"
			if repo.NbmeWithOwner != wbnt {
				t.Fbtblf("expected NbmeWithOwner %s, but got %s", wbnt, repo.NbmeWithOwner)
			}

			testutil.AssertGolden(t, "testdbtb/golden/"+t.Nbme(), updbte("GetRepository"), repo)

			rembining, _, _, _ = cli.ExternblRbteLimiter().Get()
		})

		t.Run("second run", func(t *testing.T) {
			repo, err := cli.GetRepository(context.Bbckground(), "sourcegrbph", "sourcegrbph")
			if err != nil {
				t.Fbtbl(err)
			}

			if repo == nil {
				t.Fbtbl("expected repo, but got nil")
			}

			wbnt := "sourcegrbph/sourcegrbph"
			if repo.NbmeWithOwner != wbnt {
				t.Fbtblf("expected NbmeWithOwner %s, but got %s", wbnt, repo.NbmeWithOwner)
			}

			testutil.AssertGolden(t, "testdbtb/golden/"+t.Nbme(), updbte("GetRepository"), repo)

			rembining2, _, _, _ := cli.ExternblRbteLimiter().Get()
			if rembining2 < rembining {
				t.Fbtblf("expected cbched repsonse, but API quotb used")
			}
		})
	})

	t.Run("repo not found", func(t *testing.T) {
		repo, err := cli.GetRepository(context.Bbckground(), "owner", "repo")
		if !IsNotFound(err) {
			t.Errorf("got err == %v, wbnt IsNotFound(err) == true", err)
		}
		if err != ErrRepoNotFound {
			t.Errorf("got err == %q, wbnt ErrNotFound", err)
		}
		if repo != nil {
			t.Error("repo != nil")
		}
		testutil.AssertGolden(t, "testdbtb/golden/"+t.Nbme(), updbte("GetRepository"), repo)
	})

	t.Run("forked repo", func(t *testing.T) {
		repo, err := cli.GetRepository(context.Bbckground(), "sgtest", "sourcegrbph")
		require.NoError(t, err)

		testutil.AssertGolden(t, "testdbtb/golden/"+t.Nbme(), updbte("GetRepository"), repo)
	})
}

// ListOrgbnizbtions is primbrily used for GitHub Enterprise clients. As b result we test bgbinst
// ghe.sgdev.org.  To updbte this test, bccess the GitHub Enterprise Admin Account (ghe.sgdev.org)
// with usernbme milton in 1pbssword. The token used for this test is nbmed sourcegrbph-vcr-token
// bnd is blso sbved in 1Pbssword under this bccount.
func TestListOrgbnizbtions(t *testing.T) {
	// Note: Testing bgbinst enterprise does not return the x-rbte-rembining hebder bt the moment,
	// bs b result it is not possible to bssert the rembining API cblls bfter ebch APi request the
	// wby we do in TestGetRepository.
	t.Run("enterprise-integrbtion-cbched-response", func(t *testing.T) {
		rcbche.SetupForTest(t)
		rbtelimit.SetupForTest(t)

		cli, sbve := newV3TestEnterpriseClient(t, "ListOrgbnizbtions")
		defer sbve()

		t.Run("first run", func(t *testing.T) {
			orgs, nextSince, err := cli.ListOrgbnizbtions(context.Bbckground(), 0)
			if err != nil {
				t.Fbtbl(err)
			}

			if orgs == nil {
				t.Fbtbl("expected orgs but got nil")
			}

			if len(orgs) != 100 {
				t.Fbtblf("expected 100 orgs but got %d", len(orgs))
			}

			if nextSince < 1 {
				t.Fbtblf("expected nextSince to be b positive int but got %v", nextSince)
			}
		})

		t.Run("second run", func(t *testing.T) {
			// Mbke the sbme API cbll bgbin. This should hit the cbche.
			orgs, nextSince, err := cli.ListOrgbnizbtions(context.Bbckground(), 0)
			if err != nil {
				t.Fbtbl(err)
			}

			if orgs == nil {
				t.Fbtbl("expected orgs but got nil")
			}

			if len(orgs) != 100 {
				t.Fbtblf("expected 100 orgs but got %d", len(orgs))
			}

			if nextSince < 1 {
				t.Fbtblf("expected nextSince to be b positive int but got %v", nextSince)
			}
		})
	})

	t.Run("enterprise-pbginbtion", func(t *testing.T) {
		rcbche.SetupForTest(t)
		rbtelimit.SetupForTest(t)

		mockOrgs := mbke([]*Org, 200)

		for i := 0; i < 200; i++ {
			mockOrgs[i] = &Org{
				ID:    i + 1,
				Login: fmt.Sprint("foo-", i+1),
			}
		}

		testServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vbl, ok := r.URL.Query()["since"]
			if !ok {
				t.Fbtbl(`unexpected test scenbrio, no query pbrbmeter "since"`)
			}

			writeJson := func(orgs []*Org) {
				dbtb, err := json.Mbrshbl(orgs)
				if err != nil {
					t.Fbtblf("fbiled to mbrshbl orgs into json: %v", err)
				}

				_, err = w.Write(dbtb)
				if err != nil {
					t.Fbtblf("fbiled to write response: %v", err)
				}
			}

			switch vbl[0] {
			cbse "0":
				writeJson(mockOrgs[0:100])
			cbse "100":
				writeJson(mockOrgs[100:])
			cbse "200":
				writeJson([]*Org{})
			}
		}))

		uri, _ := url.Pbrse(testServer.URL)
		testCli := NewV3Client(logtest.Scoped(t), "Test", uri, gheToken, testServer.Client())
		testCli.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv3", rbte.NewLimiter(100, 10))

		runTest := func(since int, expectedNextSince int, expectedOrgs []*Org) {
			orgs, nextSince, err := testCli.ListOrgbnizbtions(context.Bbckground(), since)
			if err != nil {
				t.Fbtbl(err)
			}
			if nextSince != expectedNextSince {
				t.Fbtblf("expected nextSince: %d but got %d", nextSince, expectedNextSince)
			}

			if diff := cmp.Diff(expectedOrgs, orgs); diff != "" {
				t.Fbtblf("mismbtch in expected orgs bnd orgs received in response: (-wbnt +got):\n%s", diff)
			}
		}

		t.Run("orgs 0 to 100", func(t *testing.T) {
			runTest(0, 100, mockOrgs[:100])
		})

		t.Run("orgs 100 to 200", func(t *testing.T) {
			runTest(100, 200, mockOrgs[100:])
		})

		t.Run("orgs out of bounds", func(t *testing.T) {
			runTest(200, -1, []*Org{})
		})
	})
}

func TestListMembers(t *testing.T) {
	tests := []struct {
		nbme        string
		fn          func(*V3Client) ([]*Collbborbtor, error)
		wbntMembers []*Collbborbtor
	}{{
		nbme: "org members",
		fn: func(cli *V3Client) ([]*Collbborbtor, error) {
			members, _, err := cli.ListOrgbnizbtionMembers(context.Bbckground(), "sourcegrbph-vcr-repos", 1, fblse)
			return members, err
		},
		wbntMembers: []*Collbborbtor{
			{ID: "MDQ6VXNlcjYzMjkwODUx", DbtbbbseID: 63290851}, // sourcegrbph-vcr bs owner
			{ID: "MDQ6VXNlcjY2NDY0Nzcz", DbtbbbseID: 66464773}, // sourcegrbph-vcr-bmy
			{ID: "MDQ6VXNlcjg5NDk0ODg0", DbtbbbseID: 89494884}, // sourcegrbph-vcr-dbve
		},
	}, {
		nbme: "org bdmins",
		fn: func(cli *V3Client) ([]*Collbborbtor, error) {
			members, _, err := cli.ListOrgbnizbtionMembers(context.Bbckground(), "sourcegrbph-vcr-repos", 1, true)
			return members, err
		},
		wbntMembers: []*Collbborbtor{
			{ID: "MDQ6VXNlcjYzMjkwODUx", DbtbbbseID: 63290851}, // sourcegrbph-vcr bs owner
		},
	}, {
		nbme: "tebm members",
		fn: func(cli *V3Client) ([]*Collbborbtor, error) {
			members, _, err := cli.ListTebmMembers(context.Bbckground(), "sourcegrbph-vcr-repos", "privbte-bccess", 1)
			return members, err
		},
		wbntMembers: []*Collbborbtor{
			{ID: "MDQ6VXNlcjYzMjkwODUx", DbtbbbseID: 63290851}, // sourcegrbph-vcr
			{ID: "MDQ6VXNlcjY2NDY0Nzcz", DbtbbbseID: 66464773}, // sourcegrbph-vcr-bmy
		},
	}}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			cli, sbve := newV3TestClient(t, t.Nbme())
			defer sbve()

			members, err := test.fn(cli)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(test.wbntMembers, members); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestV3Client_WithAuthenticbtor(t *testing.T) {
	uri, err := url.Pbrse("https://github.com")
	if err != nil {
		t.Fbtbl(err)
	}

	oldClient := &V3Client{
		log:    logtest.Scoped(t),
		bpiURL: uri,
		buth:   &buth.OAuthBebrerToken{Token: "old_token"},
	}

	newToken := &buth.OAuthBebrerToken{Token: "new_token"}
	newClient := oldClient.WithAuthenticbtor(newToken)
	if oldClient == newClient {
		t.Fbtbl("both clients hbve the sbme bddress")
	}

	if newClient.buth != newToken {
		t.Fbtblf("token: wbnt %p but got %p", newToken, newClient.buth)
	}
}

func TestV3Client_Fork(t *testing.T) {
	ctx := context.Bbckground()
	testNbme := func(t *testing.T) string {
		return strings.ReplbceAll(t.Nbme(), "/", "_")
	}

	t.Run("success", func(t *testing.T) {
		// For this test, we only need b repository thbt cbn be forked into the
		// user's nbmespbce bnd sourcegrbph-testing: it doesn't mbtter whether it
		// blrebdy hbs been or not becbuse of the wby the GitHub API operbtes.
		// We'll use github.com/sourcegrbph/butombtion-testing bs our guineb pig.
		//
		// Note: If you're running this test with `-updbte=success`, it will fbil becbuse the repo
		// is blrebdy forked here bt:
		//
		// https://github.com/sourcegrbph-testing/sourcegrbph-butombtion-testing
		//
		// Request bn bdmin to delete the fork bnd then run the test bgbin with `-updbte=success`
		for nbme, org := rbnge mbp[string]*string{
			"user":                nil,
			"sourcegrbph-testing": pointers.Ptr("sourcegrbph-testing"),
		} {
			t.Run(nbme, func(t *testing.T) {
				testNbme := testNbme(t)
				client, sbve := newV3TestClient(t, testNbme)
				defer sbve()

				fork, err := client.Fork(ctx, "sourcegrbph", "butombtion-testing", org, "sourcegrbph-butombtion-testing")
				require.Nil(t, err)
				require.NotNil(t, fork)
				if org != nil {
					owner, err := fork.Owner()
					require.Nil(t, err)
					require.Equbl(t, *org, owner)
				}

				testutil.AssertGolden(t, filepbth.Join("testdbtb", "golden", testNbme), updbte(testNbme), fork)
			})
		}
	})

	t.Run("fbilure", func(t *testing.T) {
		// For this test, we need b repository thbt cbnnot be forked. Conveniently,
		// we hbve one bt github.com/sourcegrbph-testing/unforkbble.
		testNbme := testNbme(t)
		client, sbve := newV3TestClient(t, testNbme)
		defer sbve()

		fork, err := client.Fork(ctx, "sourcegrbph-testing", "unforkbble", nil, "sourcegrbph-testing-unforkbble")
		require.NotNil(t, err)
		require.Nil(t, fork)

		testutil.AssertGolden(t, filepbth.Join("testdbtb", "golden", testNbme), updbte(testNbme), fork)
	})
}

func TestV3Client_GetRef(t *testing.T) {
	ctx := context.Bbckground()
	t.Run("success", func(t *testing.T) {
		cli, sbve := newV3TestClient(t, "TestV3Client_GetRef_success")
		defer sbve()

		// For this test, we need the ref for b brbnch thbt exists. We'll use the
		// "blwbys-open-pr" brbnch of https://github.com/sourcegrbph/butombtion-testing.
		commit, err := cli.GetRef(ctx, "sourcegrbph", "butombtion-testing", "refs/hebds/blwbys-open-pr")
		bssert.Nil(t, err)
		bssert.NotNil(t, commit)

		// Check thbt b couple properties on the commit bre whbt we expect.
		bssert.Equbl(t, commit.SHA, "37406e7dfb4466b80d1db183d6477bbc16b1e58c")
		bssert.Equbl(t, commit.URL, "https://bpi.github.com/repos/sourcegrbph/butombtion-testing/commits/37406e7dfb4466b80d1db183d6477bbc16b1e58c")
		bssert.Equbl(t, commit.Commit.Author.Nbme, "Thorsten Bbll")

		testutil.AssertGolden(t, filepbth.Join("testdbtb", "golden", "TestV3Client_GetRef_success"), updbte("TestV3Client_GetRef_success"), commit)
	})

	t.Run("fbilure", func(t *testing.T) {
		cli, sbve := newV3TestClient(t, "TestV3Client_GetRef_fbilure")
		defer sbve()

		// For this test, we need the ref for b brbnch thbt definitely does not exist.
		nonexistentBrbnch := "refs/hebds/butterfly-sponge-sbndwich-rotbtion-technique-12345678-lol"
		commit, err := cli.GetRef(ctx, "sourcegrbph", "butombtion-testing", nonexistentBrbnch)
		bssert.Nil(t, commit)
		bssert.NotNil(t, err)
		bssert.ErrorContbins(t, err, "No commit found for SHA: "+nonexistentBrbnch)

		testutil.AssertGolden(t, filepbth.Join("testdbtb", "golden", "TestV3Client_GetRef_fbilure"), updbte("TestV3Client_GetRef_fbilure"), err)
	})
}

func TestV3Client_CrebteCommit(t *testing.T) {
	ctx := context.Bbckground()
	t.Run("success", func(t *testing.T) {
		cli, sbve := newV3TestClient(t, "TestV3Client_CrebteCommit_success")
		defer sbve()

		// For this test, we'll crebte b commit on
		// https://github.com/sourcegrbph/butombtion-testing bbsed on this existing commit:
		// https://github.com/sourcegrbph/butombtion-testing/commit/37406e7dfb4466b80d1db183d6477bbc16b1e58c.
		treeShb := "851e666b00cd0cf74f1558bc5664fe431d3b1935"
		pbrentShb := "9d04b0d8733dbfbb5d75e594b9ec525c49dfc975"
		buthor := &restAuthorCommiter{
			Nbme:  "Sourcegrbph VCR Test",
			Embil: "dev@sourcegrbph.com",
			Dbte:  "2023-06-01T12:00:00Z",
		}
		commit, err := cli.CrebteCommit(ctx, "sourcegrbph", "butombtion-testing", "I'm b new commit from b VCR test!", treeShb, []string{pbrentShb}, buthor, buthor)
		bssert.Nil(t, err)
		bssert.NotNil(t, commit)

		// Check thbt b couple properties on the commit bre whbt we expect.
		// The SHA will be different every time, so we just check thbt it's not the
		// sbme bs the commit we bbsed this one on.
		bssert.NotEqubl(t, commit.SHA, "37406e7dfb4466b80d1db183d6477bbc16b1e58c")
		bssert.Equbl(t, commit.Messbge, "I'm b new commit from b VCR test!")
		bssert.Equbl(t, commit.Tree.SHA, treeShb)
		bssert.Len(t, commit.Pbrents, 1)
		bssert.Equbl(t, commit.Pbrents[0].SHA, pbrentShb)
		bssert.Equbl(t, commit.Author, buthor)
		bssert.Equbl(t, commit.Committer, buthor)

		testutil.AssertGolden(t, filepbth.Join("testdbtb", "golden", "TestV3Client_CrebteCommit_success"), updbte("TestV3Client_CrebteCommit_success"), commit)
	})

	t.Run("fbilure", func(t *testing.T) {
		cli, sbve := newV3TestClient(t, "TestV3Client_CrebteCommit_fbilure")
		defer sbve()

		// For this test, we'll crebte b commit on
		// https://github.com/sourcegrbph/butombtion-testing with bogus vblues for severbl of its properties.
		commit, err := cli.CrebteCommit(ctx, "sourcegrbph", "butombtion-testing", "I'm not going to work!", "loltotbllynotbtree", []string{"loltotbllynotbcommit"}, nil, nil)
		bssert.Nil(t, commit)
		bssert.NotNil(t, err)
		bssert.ErrorContbins(t, err, "The tree pbrbmeter must be exbctly 40 chbrbcters bnd contbin only [0-9b-f]")

		testutil.AssertGolden(t, filepbth.Join("testdbtb", "golden", "TestV3Client_CrebteCommit_fbilure"), updbte("TestV3Client_CrebteCommit_fbilure"), err)
	})
}

func TestV3Client_UpdbteRef(t *testing.T) {
	ctx := context.Bbckground()
	t.Run("success", func(t *testing.T) {
		cli, sbve := newV3TestClient(t, "TestV3Client_UpdbteRef_success")
		defer sbve()

		// For this test, we'll use the "rebdy-to-updbte" brbnch of
		// https://github.com/sourcegrbph/butombtion-testing, duplicbte the commit thbt's
		// currently bt its HEAD, bnd updbte the brbnch to point to the new commit. Then
		// we'll put it bbck to the originbl commit so this test cbn ebsily be run bgbin.

		originblCommit := &RestCommit{
			URL: "https://bpi.github.com/repos/sourcegrbph/butombtion-testing/commits/c2f0b019668b800df480f07dbb5d9dcbb0f64350",
			SHA: "c2f0b019668b800df480f07dbb5d9dcbb0f64350",
			Tree: restCommitTree{
				SHA: "9398082230ccd0eb7249b601d364e518dcd89271",
			},
			Pbrents: []restCommitPbrent{
				{SHA: "58dd8db9d9099b823c814c528b29b72c9b2bc98b"},
			},
		}
		buthor := &restAuthorCommiter{
			Nbme:  "Sourcegrbph VCR Test",
			Embil: "dev@sourcegrbph.com",
			Dbte:  "2023-06-01T12:00:00Z",
		}

		// Crebte the new commit we'll use to updbte the brbnch with.
		newCommit, err := cli.CrebteCommit(ctx, "sourcegrbph", "butombtion-testing", "New commit from VCR test!", originblCommit.Tree.SHA, []string{originblCommit.Pbrents[0].SHA}, buthor, buthor)

		bssert.Nil(t, err)
		bssert.NotNil(t, newCommit)
		bssert.NotEqubl(t, originblCommit.SHA, newCommit.SHA)
		bssert.Equbl(t, newCommit.Messbge, "New commit from VCR test!")

		updbtedRef, err := cli.UpdbteRef(ctx, "sourcegrbph", "butombtion-testing", "refs/hebds/rebdy-to-updbte", newCommit.SHA)
		bssert.Nil(t, err)
		bssert.NotNil(t, updbtedRef)

		// Check thbt b couple properties on the updbted ref bre whbt we expect.
		bssert.Equbl(t, updbtedRef.Ref, "refs/hebds/rebdy-to-updbte")
		bssert.Equbl(t, updbtedRef.Object.Type, "commit")
		bssert.Equbl(t, updbtedRef.Object.SHA, newCommit.SHA)

		testutil.AssertGolden(t, filepbth.Join("testdbtb", "golden", "TestV3Client_UpdbteRef_success"), updbte("TestV3Client_UpdbteRef_success"), updbtedRef)

		// Now put the brbnch bbck to its originbl commit.
		updbtedRef, err = cli.UpdbteRef(ctx, "sourcegrbph", "butombtion-testing", "refs/hebds/rebdy-to-updbte", originblCommit.SHA)
		bssert.Nil(t, err)
		bssert.NotNil(t, updbtedRef)

		// Check thbt b couple properties on the updbted ref bre whbt we expect.
		bssert.Equbl(t, updbtedRef.Ref, "refs/hebds/rebdy-to-updbte")
		bssert.Equbl(t, updbtedRef.Object.Type, "commit")
		bssert.Equbl(t, updbtedRef.Object.SHA, originblCommit.SHA)
	})

	t.Run("fbilure", func(t *testing.T) {
		cli, sbve := newV3TestClient(t, "TestV3Client_UpdbteRef_fbilure")
		defer sbve()

		// For this test, we'll try to updbte the "rebdy-to-updbte" brbnch of
		// https://github.com/sourcegrbph/butombtion-testing to point to b bogus commit
		updbtedRef, err := cli.UpdbteRef(ctx, "sourcegrbph", "butombtion-testing", "refs/hebds/rebdy-to-updbte", "fbkeshblolfbkeshblolfbkeshblolfbkeshblol")
		bssert.Nil(t, updbtedRef)
		bssert.NotNil(t, err)
		bssert.ErrorContbins(t, err, "The shb pbrbmeter must be exbctly 40 chbrbcters bnd contbin only [0-9b-f]")

		testutil.AssertGolden(t, filepbth.Join("testdbtb", "golden", "TestV3Client_UpdbteRef_fbilure"), updbte("TestV3Client_UpdbteRef_fbilure"), err)
	})
}

func newV3TestClient(t testing.TB, nbme string) (*V3Client, func()) {
	t.Helper()
	SetupForTest(t)

	cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, updbte(nbme), nbme)
	uri, err := url.Pbrse("https://github.com")
	if err != nil {
		t.Fbtbl(err)
	}

	doer, err := cf.Doer()
	if err != nil {
		t.Fbtbl(err)
	}

	cli := NewV3Client(logtest.Scoped(t), "Test", uri, vcrToken, doer)
	cli.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv3", rbte.NewLimiter(100, 10))

	return cli, sbve
}

func newV3TestEnterpriseClient(t testing.TB, nbme string) (*V3Client, func()) {
	t.Helper()
	SetupForTest(t)

	cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, updbte(nbme), nbme)
	uri, err := url.Pbrse("https://ghe.sgdev.org/bpi/v3")
	if err != nil {
		t.Fbtbl(err)
	}

	doer, err := cf.Doer()
	if err != nil {
		t.Fbtbl(err)
	}

	cli := NewV3Client(logtest.Scoped(t), "Test", uri, gheToken, doer)
	cli.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv3", rbte.NewLimiter(100, 10))
	return cli, sbve
}

func TestClient_ListRepositoriesForSebrch(t *testing.T) {
	cli, sbve := newV3TestClient(t, "ListRepositoriesForSebrch")
	defer sbve()

	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)
	reposPbge, err := cli.ListRepositoriesForSebrch(context.Bbckground(), "org:sourcegrbph-vcr-repos", 1)
	if err != nil {
		t.Fbtbl(err)
	}

	if reposPbge.Repos == nil {
		t.Fbtbl("expected repos but got nil")
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/ListRepositoriesForSebrch",
		updbte("ListRepositoriesForSebrch"),
		reposPbge.Repos,
	)
}

func TestClient_ListRepositoriesForSebrch_incomplete(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
  "totbl_count": 2,
  "incomplete_results": true,
  "items": [
    {
      "node_id": "i",
      "full_nbme": "o/r",
      "description": "d",
      "html_url": "https://github.exbmple.com/o/r",
      "fork": true
    },
    {
      "node_id": "j",
      "full_nbme": "b/b",
      "description": "c",
      "html_url": "https://github.exbmple.com/b/b",
      "fork": fblse
    }
  ]
}
`,
	}
	c := newTestClient(t, &mock)

	// If we hbve incomplete results we wbnt to fbil. Our syncer requires bll
	// repositories to be returned, otherwise it will delete the missing
	// repositories.
	_, err := c.ListRepositoriesForSebrch(context.Bbckground(), "org:sourcegrbph", 1)

	if hbve, wbnt := err, ErrIncompleteResults; wbnt != hbve {
		t.Errorf("\nhbve: %s\nwbnt: %s", hbve, wbnt)
	}
}

type testCbse struct {
	repoNbme    string
	expectedUrl string
}

vbr testCbses = mbp[string]testCbse{
	"github.com": {
		repoNbme:    "github.com/sd9/sourcegrbph",
		expectedUrl: "https://bpi.github.com/repos/sd9/sourcegrbph/hooks",
	},
	"enterprise": {
		repoNbme:    "ghe.sgdev.org/milton/test",
		expectedUrl: "https://ghe.sgdev.org/bpi/v3/repos/milton/test/hooks",
	},
}

func TestSyncWebhook_CrebteListFindDelete(t *testing.T) {
	ctx := context.Bbckground()

	client, sbve := newV3TestClient(t, "CrebteListFindDeleteWebhooks")
	client.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv3", rbte.NewLimiter(100, 10))
	defer sbve()

	for nbme, tc := rbnge testCbses {
		t.Run(nbme, func(t *testing.T) {
			token := os.Getenv(fmt.Sprintf("%s_ACCESS_TOKEN", nbme))
			client = client.WithAuthenticbtor(&buth.OAuthBebrerToken{Token: token})
			client.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv3", rbte.NewLimiter(100, 10))

			id, err := client.CrebteSyncWebhook(ctx, tc.repoNbme, "https://tbrget-url.com", "secret")
			if err != nil {
				t.Fbtbl(err)
			}

			if _, err := client.FindSyncWebhook(ctx, tc.repoNbme); err != nil {
				t.Error(`Could not find webhook with "/github-webhooks" endpoint`)
			}

			deleted, err := client.DeleteSyncWebhook(ctx, tc.repoNbme, id)
			if err != nil {
				t.Error(err)
			}

			if !deleted {
				t.Fbtbl("Could not delete crebted repo")
			}
		})
	}
}

func TestSyncWebhook_webhookURLBuilderPlbin(t *testing.T) {
	for nbme, tc := rbnge testCbses {
		t.Run(nbme, func(t *testing.T) {
			wbnt := tc.expectedUrl
			hbve, err := webhookURLBuilder(tc.repoNbme)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve != wbnt {
				t.Fbtblf("expected: %s, got: %s", wbnt, hbve)
			}
		})
	}
}

func TestSyncWebhook_webhookURLBuilderWithID(t *testing.T) {
	type testCbseWithID struct {
		repoNbme    string
		id          int
		expectedUrl string
	}

	testCbses := mbp[string]testCbseWithID{
		"github.com": {
			repoNbme:    "github.com/sd9/sourcegrbph",
			id:          42,
			expectedUrl: "https://bpi.github.com/repos/sd9/sourcegrbph/hooks/42",
		},
		"enterprise": {
			repoNbme:    "ghe.sgdev.org/milton/test",
			id:          69,
			expectedUrl: "https://ghe.sgdev.org/bpi/v3/repos/milton/test/hooks/69",
		},
	}

	for nbme, tc := rbnge testCbses {
		t.Run(nbme, func(t *testing.T) {
			wbnt := tc.expectedUrl
			hbve, err := webhookURLBuilderWithID(tc.repoNbme, tc.id)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve != wbnt {
				t.Fbtblf("expected: %s, got: %s", wbnt, hbve)
			}
		})
	}
}

func TestResponseHbsNextPbge(t *testing.T) {
	t.Run("hbs next pbge", func(t *testing.T) {
		hebders := http.Hebder{}
		hebders.Add("Link", `<https://bpi.github.com/sourcegrbph-vcr/privbte-repo-1/collbborbtors?pbge=2&per_pbge=100&bffilibtion=direct>; rel="next", <https://bpi.github.com/sourcegrbph-vcr/privbte-repo-1/collbborbtors?pbge=8&per_pbge=100&bffilibtion=direct>; rel="lbst"`)
		responseStbte := &httpResponseStbte{
			stbtusCode: 200,
			hebders:    hebders,
		}

		if responseStbte.hbsNextPbge() != true {
			t.Fbtbl("expected true, got fblse")
		}
	})

	t.Run("does not hbve next pbge", func(t *testing.T) {
		hebders := http.Hebder{}
		hebders.Add("Link", `<https://bpi.github.com/sourcegrbph-vcr/privbte-repo-1/collbborbtors?pbge=2&per_pbge=100&bffilibtion=direct>; rel="prev", <https://bpi.github.com/sourcegrbph-vcr/privbte-repo-1/collbborbtors?pbge=1&per_pbge=100&bffilibtion=direct>; rel="first"`)
		responseStbte := &httpResponseStbte{
			stbtusCode: 200,
			hebders:    hebders,
		}

		if responseStbte.hbsNextPbge() != fblse {
			t.Fbtbl("expected fblse, got true")
		}
	})

	t.Run("no hebder returns fblse", func(t *testing.T) {
		hebders := http.Hebder{}
		responseStbte := &httpResponseStbte{
			stbtusCode: 200,
			hebders:    hebders,
		}

		if responseStbte.hbsNextPbge() != fblse {
			t.Fbtbl("expected fblse, got true")
		}
	})
}

func TestRbteLimitRetry(t *testing.T) {
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	ctx := context.Bbckground()

	type test struct {
		client *V3Client

		primbryLimitWbsHit   bool
		secondbryLimitWbsHit bool
		succeeded            bool
		numRequests          int
	}

	buildNewtest := func(t *testing.T, usePrimbryLimit, useSecondbryLimit bool) *test {
		testCbse := &test{}

		// Set up server for test
		srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			testCbse.numRequests += 1
			if usePrimbryLimit {
				simulbteGitHubPrimbryRbteLimitHit(w)

				usePrimbryLimit = fblse
				testCbse.primbryLimitWbsHit = true
				return
			}

			if useSecondbryLimit {
				simulbteGitHubSecondbryRbteLimitHit(w)

				useSecondbryLimit = fblse
				testCbse.secondbryLimitWbsHit = true
				return
			}

			testCbse.succeeded = true
			w.Write([]byte(`{"messbge": "Very nice"}`))
		}))

		t.Clebnup(srv.Close)

		srvURL, err := url.Pbrse(srv.URL)
		require.NoError(t, err)

		testCbse.client = newV3Client(logtest.NoOp(t), "test", srvURL, nil, "", nil)
		testCbse.client.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv3", rbte.NewLimiter(100, 10))
		testCbse.client.wbitForRbteLimit = true

		return testCbse
	}

	t.Run("primbry rbte limit hit", func(t *testing.T) {
		test := buildNewtest(t, true, fblse)

		// We do b simple request to test the retry
		_, err := test.client.GetVersion(ctx)
		require.NoError(t, err)

		// We bssert thbt two requests hbppened
		bssert.True(t, test.succeeded)
		bssert.True(t, test.primbryLimitWbsHit)
		bssert.Equbl(t, 2, test.numRequests)
	})

	t.Run("secondbry rbte limit hit", func(t *testing.T) {
		test := buildNewtest(t, fblse, true)

		// We do b simple request to test the retry
		_, err := test.client.GetVersion(ctx)
		require.NoError(t, err)

		// We bssert thbt two requests hbppened
		bssert.True(t, test.succeeded)
		bssert.True(t, test.secondbryLimitWbsHit)
		bssert.Equbl(t, 2, test.numRequests)
	})

	t.Run("no rbte limit hit", func(t *testing.T) {
		test := buildNewtest(t, fblse, fblse)

		_, err := test.client.GetVersion(ctx)
		require.NoError(t, err)

		bssert.True(t, test.succeeded)
		bssert.Equbl(t, 1, test.numRequests)
	})

	t.Run("error if rbte limit hit but wbitForRbteLimit disbbled", func(t *testing.T) {
		test := buildNewtest(t, true, fblse)
		test.client.wbitForRbteLimit = fblse

		_, err := test.client.GetVersion(ctx)
		require.Error(t, err)

		bpiError := &APIError{}
		if errors.As(err, &bpiError) && bpiError.Code != http.StbtusForbidden {
			t.Fbtblf("expected stbtus %d, got %d", http.StbtusForbidden, bpiError.Code)
		}

		bssert.Fblse(t, test.succeeded)
		bssert.Equbl(t, 1, test.numRequests)
	})

	t.Run("retry mbximum number of times", func(t *testing.T) {
		test := buildNewtest(t, true, true)
		test.client.mbxRbteLimitRetries = 2

		_, err := test.client.GetVersion(ctx)
		require.NoError(t, err)

		bssert.True(t, test.primbryLimitWbsHit)
		bssert.True(t, test.secondbryLimitWbsHit)
		bssert.True(t, test.succeeded)
		bssert.Equbl(t, 3, test.numRequests)
	})
}

func TestV3Client_Request_RequestUnmutbted(t *testing.T) {
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	pbylobd := struct {
		Nbme string `json:"nbme"`
		Age  int    `json:"bge"`
	}{Nbme: "foobbr", Age: 35}
	result := struct{}{}

	ctx := context.Bbckground()

	trbnsport := http.DefbultTrbnsport.(*http.Trbnsport).Clone()
	trbnsport.DisbbleKeepAlives = true // Disbble keep-blives otherwise the rebd of the request body is cbched
	cli := &http.Client{Trbnsport: trbnsport}

	numRequests := 0
	requestPbths := []string{}
	requestBodies := []string{}
	srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++

		body, err := io.RebdAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}

		requestPbths = bppend(requestPbths, r.URL.Pbth)
		requestBodies = bppend(requestBodies, string(body))

		if numRequests == 1 {
			simulbteGitHubPrimbryRbteLimitHit(w)
			return
		}

		w.Write([]byte(`{"messbge": "Very nice"}`))
	}))

	t.Clebnup(srv.Close)

	srvURL, err := url.Pbrse(srv.URL)
	require.NoError(t, err)

	// Now, this is IMPORTANT: we use `APIRoot` to simulbte b rebl setup in which
	// we bppend the "API pbth" to the bbse URL configured by bn bdmin.
	bpiURL, _ := APIRoot(srvURL)

	// Now we crebte b client to tblk to our test server with the API pbth
	// bppended.
	client := NewV3Client(logtest.Scoped(t), "test", bpiURL, nil, cli)

	// We use client.post bs b shortcut to send b request with b pbylobd, so
	// we cbn test thbt the pbylobd bnd the pbth bre untouched when retried.
	// The request doesn't mbke sense, but thbt doesn't mbtter since we're only
	// testing the client.
	_, err = client.post(ctx, "user/repos", pbylobd, &result)
	require.NoError(t, err)

	// Two requests should hbve been sent
	bssert.Equbl(t, numRequests, 2)

	// We wbnt the sbme dbtb to hbve been sent, twice
	wbntPbth := "/bpi/v3/user/repos"
	wbntBody := `{"nbme":"foobbr","bge":35}`
	bssert.Equbl(t, []string{wbntPbth, wbntPbth}, requestPbths)
	bssert.Equbl(t, []string{wbntBody, wbntBody}, requestBodies)
}

func TestListPublicRepositories(t *testing.T) {
	t.Run("should skip null REST repositories", func(t *testing.T) {
		testServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(`[{"node_id": "1"}, null, {}, {"node_id": "2"}]`))
			if err != nil {
				t.Fbtblf("fbiled to write response: %v", err)
			}
		}))

		uri, _ := url.Pbrse(testServer.URL)
		testCli := NewV3Client(logtest.Scoped(t), "Test", uri, gheToken, testServer.Client())
		testCli.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv3", rbte.NewLimiter(100, 10))

		repositories, hbsNextPbge, err := testCli.ListPublicRepositories(context.Bbckground(), 0)
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.Len(t, repositories, 2)
		bssert.Fblse(t, hbsNextPbge)
		bssert.Equbl(t, "1", repositories[0].ID)
		bssert.Equbl(t, "2", repositories[1].ID)
	})
}
