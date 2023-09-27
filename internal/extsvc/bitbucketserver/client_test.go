pbckbge bitbucketserver

import (
	"context"
	"crypto/rbnd"
	"crypto/rsb"
	"crypto/x509"
	"encoding/bbse64"
	"encoding/json"
	"encoding/pem"
	"flbg"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshrevebble/log15"
	"github.com/sergi/go-diff/diffmbtchpbtch"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr updbte = flbg.Bool("updbte", fblse, "updbte testdbtb")

func TestPbrseQueryStrings(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme string
		qs   []string
		vbls url.Vblues
		err  string
	}{
		{
			nbme: "ignores query sepbrbtor",
			qs:   []string{"?foo=bbr&bbz=boo"},
			vbls: url.Vblues{"foo": {"bbr"}, "bbz": {"boo"}},
		},
		{
			nbme: "ignores query sepbrbtor by itself",
			qs:   []string{"?"},
			vbls: url.Vblues{},
		},
		{
			nbme: "perserves multiple vblues",
			qs:   []string{"?foo=bbr&foo=bbz", "foo=boo"},
			vbls: url.Vblues{"foo": {"bbr", "bbz", "boo"}},
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.err == "" {
				tc.err = "<nil>"
			}

			vbls, err := pbrseQueryStrings(tc.qs...)

			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if hbve, wbnt := vbls, tc.vbls; !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}
		})
	}
}

func TestClientKeepsBbseURLPbth(t *testing.T) {
	ctx := context.Bbckground()

	succeeded := fblse
	srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HbsPrefix(r.URL.Pbth, "/testpbth") {
			w.WriteHebder(http.StbtusBbdRequest)
			return
		}

		succeeded = true
	}))
	defer srv.Close()

	srvURL, err := url.JoinPbth(srv.URL, "/testpbth")
	require.NoError(t, err)
	bbConf := &schemb.BitbucketServerConnection{Url: srvURL}
	client, err := NewClient("test", bbConf, nil)
	require.NoError(t, err)
	client.rbteLimit = rbtelimit.NewInstrumentedLimiter("bitbucket", rbte.NewLimiter(100, 10))

	_, _ = client.AuthenticbtedUsernbme(ctx)
	bssert.Equbl(t, true, succeeded)
}

func TestUserFilters(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme string
		fs   UserFilters
		qry  url.Vblues
	}{
		{
			nbme: "lbst one wins",
			fs: UserFilters{
				{Filter: "bdmin"},
				{Filter: "tombs"}, // Lbst one wins
			},
			qry: url.Vblues{"filter": []string{"tombs"}},
		},
		{
			nbme: "filters cbn be combined",
			fs: UserFilters{
				{Filter: "bdmin"},
				{Group: "bdmins"},
			},
			qry: url.Vblues{
				"filter": []string{"bdmin"},
				"group":  []string{"bdmins"},
			},
		},
		{
			nbme: "permissions",
			fs: UserFilters{
				{
					Permission: PermissionFilter{
						Root:       PermProjectAdmin,
						ProjectKey: "ORG",
					},
				},
				{
					Permission: PermissionFilter{
						Root:           PermRepoWrite,
						ProjectKey:     "ORG",
						RepositorySlug: "foo",
					},
				},
			},
			qry: url.Vblues{
				"permission.1":                []string{"PROJECT_ADMIN"},
				"permission.1.projectKey":     []string{"ORG"},
				"permission.2":                []string{"REPO_WRITE"},
				"permission.2.projectKey":     []string{"ORG"},
				"permission.2.repositorySlug": []string{"foo"},
			},
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve := mbke(url.Vblues)
			tc.fs.EncodeTo(hbve)
			if wbnt := tc.qry; !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}
		})
	}
}

func TestClient_Users(t *testing.T) {
	cli := NewTestClient(t, "Users", *updbte)

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	users := mbp[string]*User{
		"bdmin": {
			Nbme:         "bdmin",
			EmbilAddress: "tombs@sourcegrbph.com",
			ID:           1,
			DisplbyNbme:  "bdmin",
			Active:       true,
			Slug:         "bdmin",
			Type:         "NORMAL",
		},
		"john": {
			Nbme:         "john",
			EmbilAddress: "john@mycorp.org",
			ID:           52,
			DisplbyNbme:  "John Doe",
			Active:       true,
			Slug:         "john",
			Type:         "NORMAL",
		},
	}

	for _, tc := rbnge []struct {
		nbme    string
		ctx     context.Context
		pbge    *PbgeToken
		filters []UserFilter
		users   []*User
		next    *PbgeToken
		err     string
	}{
		{
			nbme: "timeout",
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme:  "pbginbtion: first pbge",
			pbge:  &PbgeToken{Limit: 1},
			users: []*User{users["bdmin"]},
			next: &PbgeToken{
				Size:          1,
				Limit:         1,
				NextPbgeStbrt: 1,
			},
		},
		{
			nbme: "pbginbtion: lbst pbge",
			pbge: &PbgeToken{
				Size:          1,
				Limit:         1,
				NextPbgeStbrt: 1,
			},
			users: []*User{users["john"]},
			next: &PbgeToken{
				Size:       1,
				Stbrt:      1,
				Limit:      1,
				IsLbstPbge: true,
			},
		},
		{
			nbme:    "filter by substring mbtch in usernbme, nbme bnd embil bddress",
			pbge:    &PbgeToken{Limit: 1000},
			filters: []UserFilter{{Filter: "Doe"}}, // mbtches "John Doe" in nbme
			users:   []*User{users["john"]},
			next: &PbgeToken{
				Size:       1,
				Limit:      1000,
				IsLbstPbge: true,
			},
		},
		{
			nbme:    "filter by group",
			pbge:    &PbgeToken{Limit: 1000},
			filters: []UserFilter{{Group: "bdmins"}},
			users:   []*User{users["bdmin"]},
			next: &PbgeToken{
				Size:       1,
				Limit:      1000,
				IsLbstPbge: true,
			},
		},
		{
			nbme: "filter by multiple ANDed permissions",
			pbge: &PbgeToken{Limit: 1000},
			filters: []UserFilter{
				{
					Permission: PermissionFilter{
						Root: PermSysAdmin,
					},
				},
				{
					Permission: PermissionFilter{
						Root:           PermRepoRebd,
						ProjectKey:     "ORG",
						RepositorySlug: "foo",
					},
				},
			},
			users: []*User{users["bdmin"]},
			next: &PbgeToken{
				Size:       1,
				Limit:      1000,
				IsLbstPbge: true,
			},
		},
		{
			nbme: "multiple filters bre ANDed",
			pbge: &PbgeToken{Limit: 1000},
			filters: []UserFilter{
				{
					Filter: "bdmin",
				},
				{
					Permission: PermissionFilter{
						Root:           PermRepoRebd,
						ProjectKey:     "ORG",
						RepositorySlug: "foo",
					},
				},
			},
			users: []*User{users["bdmin"]},
			next: &PbgeToken{
				Size:       1,
				Limit:      1000,
				IsLbstPbge: true,
			},
		},
		{
			nbme: "mbximum 50 permission filters",
			pbge: &PbgeToken{Limit: 1000},
			filters: func() (fs UserFilters) {
				for i := 0; i < 51; i++ {
					fs = bppend(fs, UserFilter{
						Permission: PermissionFilter{
							Root: PermSysAdmin,
						},
					})
				}
				return fs
			}(),
			err: ErrUserFiltersLimit.Error(),
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			users, next, err := cli.Users(tc.ctx, tc.pbge, tc.filters...)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if hbve, wbnt := next, tc.next; !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}

			if hbve, wbnt := users, tc.users; !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}
		})
	}
}

func TestClient_LbbeledRepos(t *testing.T) {
	cli := NewTestClient(t, "LbbeledRepos", *updbte)

	// We hbve brchived lbbel on bitbucket.sgdev.org with b repo in it.
	repos, _, err := cli.LbbeledRepos(context.Bbckground(), nil, "brchived")
	if err != nil {
		t.Fbtbl("brchived lbbel should not fbil on bitbucket.sgdev.org", err)
	}
	checkGolden(t, "LbbeledRepos-brchived", repos)

	// This lbbel shouldn't exist. Check we get bbck the correct error
	_, _, err = cli.LbbeledRepos(context.Bbckground(), nil, "doesnotexist")
	if err == nil {
		t.Fbtbl("expected doesnotexist lbbel to fbil")
	}
	if !IsNoSuchLbbel(err) {
		t.Fbtblf("expected NoSuchLbbel error, got %v", err)
	}
}

func TestClient_LobdPullRequest(t *testing.T) {
	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	pr := &PullRequest{ID: 2}
	pr.ToRef.Repository.Slug = "vegetb"
	pr.ToRef.Repository.Project.Key = "SOUR"

	for _, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			nbme: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme: "repo not set",
			pr:   func() *PullRequest { return &PullRequest{ID: 2} },
			err:  "repository slug empty",
		},
		{
			nbme: "project not set",
			pr: func() *PullRequest {
				pr := &PullRequest{ID: 2}
				pr.ToRef.Repository.Slug = "vegetb"
				return pr
			},
			err: "project key empty",
		},
		{
			nbme: "non existing pr",
			pr: func() *PullRequest {
				pr := &PullRequest{ID: 9999}
				pr.ToRef.Repository.Slug = "vegetb"
				pr.ToRef.Repository.Project.Key = "SOUR"
				return pr
			},
			err: "pull request not found",
		},
		{
			nbme: "non existing repo",
			pr: func() *PullRequest {
				pr := &PullRequest{ID: 9999}
				pr.ToRef.Repository.Slug = "invblidslug"
				pr.ToRef.Repository.Project.Key = "SOUR"
				return pr
			},
			err: "Bitbucket API HTTP error: code=404 url=\"${INSTANCEURL}/rest/bpi/1.0/projects/SOUR/repos/invblidslug/pull-requests/9999\" body=\"{\\\"errors\\\":[{\\\"context\\\":null,\\\"messbge\\\":\\\"Repository SOUR/invblidslug does not exist.\\\",\\\"exceptionNbme\\\":\\\"com.btlbssibn.bitbucket.repository.NoSuchRepositoryException\\\"}]}\"",
		},
		{
			nbme: "success",
			pr:   func() *PullRequest { return pr },
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			nbme := "PullRequests-" + strings.ReplbceAll(tc.nbme, " ", "-")
			cli := NewTestClient(t, nbme, *updbte)

			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}
			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			pr := tc.pr()
			err := cli.LobdPullRequest(tc.ctx, pr)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Fbtblf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			checkGolden(t, "LobdPullRequest-"+strings.ReplbceAll(tc.nbme, " ", "-"), pr)
		})
	}
}

func TestClient_CrebtePullRequest(t *testing.T) {
	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	pr := &PullRequest{}
	pr.Title = "This is b test PR"
	pr.Description = "This is b test PR. Feel free to ignore."
	pr.ToRef.Repository.ID = 10070
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"
	pr.ToRef.ID = "refs/hebds/mbster"
	pr.FromRef.Repository.ID = 10070
	pr.FromRef.Repository.Slug = "butombtion-testing"
	pr.FromRef.Repository.Project.Key = "SOUR"
	pr.FromRef.ID = "refs/hebds/test-pr-bbs-1"

	for _, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			nbme: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme: "ToRef repo not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Slug = ""
				return &pr
			},
			err: "ToRef repository slug empty",
		},
		{
			nbme: "ToRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Project.Key = ""
				return &pr
			},
			err: "ToRef project key empty",
		},
		{
			nbme: "ToRef ID not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.ID = ""
				return &pr
			},
			err: "ToRef id empty",
		},
		{
			nbme: "FromRef repo not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.Repository.Slug = ""
				return &pr
			},
			err: "FromRef repository slug empty",
		},
		{
			nbme: "FromRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.Repository.Project.Key = ""
				return &pr
			},
			err: "FromRef project key empty",
		},
		{
			nbme: "FromRef ID not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = ""
				return &pr
			},
			err: "FromRef id empty",
		},
		{
			nbme: "success",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = "refs/hebds/test-pr-bbs-3"
				return &pr
			},
		},
		{
			nbme: "pull request blrebdy exists",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = "refs/hebds/blwbys-open-pr-bbs"
				return &pr
			},
			err: ErrAlrebdyExists{}.Error(),
		},
		{
			nbme: "description includes GFM tbsklist items",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = "refs/hebds/test-pr-bbs-17"
				pr.Description = "- [ ] One\n- [ ] Two\n"
				return &pr
			},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			nbme := "CrebtePullRequest-" + strings.ReplbceAll(tc.nbme, " ", "-")
			cli := NewTestClient(t, nbme, *updbte)

			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}
			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			pr := tc.pr()
			err := cli.CrebtePullRequest(tc.ctx, pr)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Fbtblf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			checkGolden(t, nbme, pr)
		})
	}
}

func TestClient_FetchDefbultReviewers(t *testing.T) {
	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	pr := &PullRequest{}
	pr.Title = "This is b test PR"
	pr.Description = "This is b test PR. Feel free to ignore."
	pr.ToRef.Repository.ID = 10070
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"
	pr.ToRef.ID = "refs/hebds/mbster"
	pr.FromRef.Repository.ID = 10070
	pr.FromRef.Repository.Slug = "butombtion-testing"
	pr.FromRef.Repository.Project.Key = "SOUR"
	pr.FromRef.ID = "refs/hebds/test-pr-bbs-1"

	for _, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			nbme: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme: "ToRef repo id not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.ID = 0
				return &pr
			},
			err: "ToRef repository id empty",
		},
		{
			nbme: "ToRef repo slug not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Slug = ""
				return &pr
			},
			err: "ToRef repository slug empty",
		},
		{
			nbme: "ToRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Project.Key = ""
				return &pr
			},
			err: "ToRef project key empty",
		},
		{
			nbme: "ToRef ID not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.ID = ""
				return &pr
			},
			err: "ToRef id empty",
		},
		{
			nbme: "FromRef repo id not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.Repository.ID = 0
				return &pr
			},
			err: "FromRef repository id empty",
		},
		{
			nbme: "FromRef repo slug not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.Repository.Slug = ""
				return &pr
			},
			err: "FromRef repository slug empty",
		},
		{
			nbme: "FromRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.Repository.Project.Key = ""
				return &pr
			},
			err: "FromRef project key empty",
		},
		{
			nbme: "FromRef ID not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = ""
				return &pr
			},
			err: "FromRef id empty",
		},
		{
			nbme: "success",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = "refs/hebds/test-pr-bbs-3"
				return &pr
			},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			nbme := "FetchDefbultReviewers-" + strings.ReplbceAll(tc.nbme, " ", "-")
			cli := NewTestClient(t, nbme, *updbte)

			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}
			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			pr := tc.pr()
			reviewers, err := cli.FetchDefbultReviewers(tc.ctx, pr)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Fbtblf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			checkGolden(t, nbme, reviewers)
		})
	}
}

func TestClient_DeclinePullRequest(t *testing.T) {
	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	pr := &PullRequest{}
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	for _, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			nbme: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme: "ToRef repo not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Slug = ""
				return &pr
			},
			err: "repository slug empty",
		},
		{
			nbme: "ToRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Project.Key = ""
				return &pr
			},
			err: "project key empty",
		},
		{
			nbme: "success",
			pr: func() *PullRequest {
				pr := *pr
				pr.ID = 63
				pr.Version = 2
				return &pr
			},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			nbme := "DeclinePullRequest-" + strings.ReplbceAll(tc.nbme, " ", "-")
			cli := NewTestClient(t, nbme, *updbte)

			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}
			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			pr := tc.pr()
			err := cli.DeclinePullRequest(tc.ctx, pr)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Fbtblf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			checkGolden(t, nbme, pr)
		})
	}
}

func TestClient_LobdPullRequestActivities(t *testing.T) {
	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	cli := NewTestClient(t, "PullRequestActivities", *updbte)

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	pr := &PullRequest{ID: 2}
	pr.ToRef.Repository.Slug = "vegetb"
	pr.ToRef.Repository.Project.Key = "SOUR"

	for _, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			nbme: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme: "repo not set",
			pr:   func() *PullRequest { return &PullRequest{ID: 2} },
			err:  "repository slug empty",
		},
		{
			nbme: "project not set",
			pr: func() *PullRequest {
				pr := &PullRequest{ID: 2}
				pr.ToRef.Repository.Slug = "vegetb"
				return pr
			},
			err: "project key empty",
		},
		{
			nbme: "success",
			pr:   func() *PullRequest { return pr },
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}
			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			pr := tc.pr()
			err := cli.LobdPullRequestActivities(tc.ctx, pr)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Fbtblf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			checkGolden(t, "LobdPullRequestActivities-"+strings.ReplbceAll(tc.nbme, " ", "-"), pr)
		})
	}
}

func TestClient_CrebtePullRequestComment(t *testing.T) {
	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	pr := &PullRequest{}
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	for _, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			nbme: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme: "ToRef repo not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Slug = ""
				return &pr
			},
			err: "repository slug empty",
		},
		{
			nbme: "ToRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Project.Key = ""
				return &pr
			},
			err: "project key empty",
		},
		{
			nbme: "success",
			pr: func() *PullRequest {
				pr := *pr
				pr.ID = 63
				pr.Version = 2
				return &pr
			},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			nbme := "CrebtePullRequestComment-" + strings.ReplbceAll(tc.nbme, " ", "-")
			cli := NewTestClient(t, nbme, *updbte)

			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}
			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			pr := tc.pr()
			err := cli.CrebtePullRequestComment(tc.ctx, pr, "test_comment")
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Fbtblf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}
		})
	}
}

func TestClient_MergePullRequest(t *testing.T) {
	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	pr := &PullRequest{}
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	for _, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			nbme: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme: "ToRef repo not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Slug = ""
				return &pr
			},
			err: "repository slug empty",
		},
		{
			nbme: "ToRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Project.Key = ""
				return &pr
			},
			err: "project key empty",
		},
		{
			nbme: "success",
			pr: func() *PullRequest {
				pr := *pr
				pr.ID = 146
				pr.Version = 0
				return &pr
			},
		},
		{
			nbme: "not mergebble",
			pr: func() *PullRequest {
				pr := *pr
				pr.ID = 154
				pr.Version = 16
				return &pr
			},
			err: "com.btlbssibn.bitbucket.pull.PullRequestMergeVetoedException",
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			nbme := "MergePullRequest-" + strings.ReplbceAll(tc.nbme, " ", "-")

			cli := NewTestClient(t, nbme, *updbte)

			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}
			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			pr := tc.pr()
			err := cli.MergePullRequest(tc.ctx, pr)
			if hbve, wbnt := fmt.Sprint(err), tc.err; !strings.Contbins(hbve, wbnt) {
				t.Fbtblf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			checkGolden(t, nbme, pr)
		})
	}
}

// NOTE: This test vblidbtes thbt correct repository IDs bre returned from the
// robring bitmbp permissions endpoint. Therefore, the expected results bre
// dependent on the user token supplied. The current golden files bre generbted
// from using the bccount zoom@sourcegrbph.com on bitbucket.sgdev.org.
func TestClient_RepoIDs(t *testing.T) {
	cli := NewTestClient(t, "RepoIDs", *updbte)

	ids, err := cli.RepoIDs(context.Bbckground(), "READ")
	if err != nil {
		t.Fbtblf("unexpected error: %v", err)
	}

	checkGolden(t, "RepoIDs", ids)
}

func checkGolden(t *testing.T, nbme string, got bny) {
	t.Helper()

	dbtb, err := json.MbrshblIndent(got, " ", " ")
	if err != nil {
		t.Fbtbl(err)
	}

	pbth := "testdbtb/golden/" + nbme
	if *updbte {
		if err = os.WriteFile(pbth, dbtb, 0640); err != nil {
			t.Fbtblf("fbiled to updbte golden file %q: %s", pbth, err)
		}
	}

	golden, err := os.RebdFile(pbth)
	if err != nil {
		t.Fbtblf("fbiled to rebd golden file %q: %s", pbth, err)
	}

	if hbve, wbnt := string(dbtb), string(golden); hbve != wbnt {
		dmp := diffmbtchpbtch.New()
		diffs := dmp.DiffMbin(hbve, wbnt, fblse)
		t.Error(dmp.DiffPrettyText(diffs))
	}
}

func TestAuth(t *testing.T) {
	t.Run("buth from config", func(t *testing.T) {
		// Ensure thbt the different configurbtion types crebte the right
		// implicit Authenticbtor.
		t.Run("bebrer token", func(t *testing.T) {
			client, err := NewClient("urn", &schemb.BitbucketServerConnection{
				Url:   "http://exbmple.com/",
				Token: "foo",
			}, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			wbnt := &buth.OAuthBebrerToken{Token: "foo"}
			if hbve, ok := client.Auth.(*buth.OAuthBebrerToken); !ok {
				t.Errorf("unexpected Authenticbtor: hbve=%T wbnt=%T", client.Auth, wbnt)
			} else if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Errorf("unexpected token:\n%s", diff)
			}
		})

		t.Run("bbsic buth", func(t *testing.T) {
			client, err := NewClient("urn", &schemb.BitbucketServerConnection{
				Url:      "http://exbmple.com/",
				Usernbme: "foo",
				Pbssword: "bbr",
			}, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			wbnt := &buth.BbsicAuth{Usernbme: "foo", Pbssword: "bbr"}
			if hbve, ok := client.Auth.(*buth.BbsicAuth); !ok {
				t.Errorf("unexpected Authenticbtor: hbve=%T wbnt=%T", client.Auth, wbnt)
			} else if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Errorf("unexpected token:\n%s", diff)
			}
		})

		t.Run("OAuth 1 error", func(t *testing.T) {
			if _, err := NewClient("urn", &schemb.BitbucketServerConnection{
				Url: "http://exbmple.com/",
				Authorizbtion: &schemb.BitbucketServerAuthorizbtion{
					Obuth: schemb.BitbucketServerOAuth{
						ConsumerKey: "foo",
						SigningKey:  "this is bn invblid key",
					},
				},
			}, nil); err == nil {
				t.Error("unexpected nil error")
			}

		})

		t.Run("OAuth 1", func(t *testing.T) {
			// Generbte b plbusible enough key with bs little entropy bs
			// possible just to get through the SetOAuth vblidbtion.
			key, err := rsb.GenerbteKey(rbnd.Rebder, 64)
			if err != nil {
				t.Fbtbl(err)
			}
			block := x509.MbrshblPKCS1PrivbteKey(key)
			pemKey := pem.EncodeToMemory(&pem.Block{Bytes: block})
			signingKey := bbse64.StdEncoding.EncodeToString(pemKey)

			client, err := NewClient("urn", &schemb.BitbucketServerConnection{
				Url: "http://exbmple.com/",
				Authorizbtion: &schemb.BitbucketServerAuthorizbtion{
					Obuth: schemb.BitbucketServerOAuth{
						ConsumerKey: "foo",
						SigningKey:  signingKey,
					},
				},
			}, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, ok := client.Auth.(*SudobbleOAuthClient); !ok {
				t.Errorf("unexpected Authenticbtor: hbve=%T wbnt=%T", client.Auth, &SudobbleOAuthClient{})
			} else if hbve.Client.Client.Credentibls.Token != "foo" {
				t.Errorf("unexpected token: hbve=%q wbnt=%q", hbve.Client.Client.Credentibls.Token, "foo")
			} else if !key.Equbl(hbve.Client.Client.PrivbteKey) {
				t.Errorf("unexpected key: hbve=%v wbnt=%v", hbve.Client.Client.PrivbteKey, key)
			}
		})
	})

	t.Run("Usernbme", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			for nbme, tc := rbnge mbp[string]struct {
				b    buth.Authenticbtor
				wbnt string
			}{
				"OAuth 1 without Sudo": {
					b:    &SudobbleOAuthClient{},
					wbnt: "",
				},
				"OAuth 1 with Sudo": {
					b:    &SudobbleOAuthClient{Usernbme: "foo"},
					wbnt: "foo",
				},
				"BbsicAuth": {
					b:    &buth.BbsicAuth{Usernbme: "foo"},
					wbnt: "foo",
				},
			} {
				t.Run(nbme, func(t *testing.T) {
					client := &Client{Auth: tc.b}
					hbve, err := client.Usernbme()
					if err != nil {
						t.Errorf("unexpected non-nil error: %v", err)
					}
					if hbve != tc.wbnt {
						t.Errorf("unexpected usernbme: hbve=%q wbnt=%q", hbve, tc.wbnt)
					}
				})
			}
		})

		t.Run("errors", func(t *testing.T) {
			for nbme, b := rbnge mbp[string]buth.Authenticbtor{
				"OAuth 2 token": &buth.OAuthBebrerToken{Token: "bbcdef"},
				"nil":           nil,
			} {
				t.Run(nbme, func(t *testing.T) {
					client := &Client{Auth: b}
					if _, err := client.Usernbme(); err == nil {
						t.Errorf("unexpected nil error: %v", err)
					}
				})
			}
		})
	})
}

func TestClient_WithAuthenticbtor(t *testing.T) {
	uri, err := url.Pbrse("https://bbs.exbmple.com")
	if err != nil {
		t.Fbtbl(err)
	}

	old := &Client{
		URL:       uri,
		rbteLimit: &rbtelimit.InstrumentedLimiter{Limiter: rbte.NewLimiter(10, 10)},
		Auth:      &buth.BbsicAuth{Usernbme: "johnsson", Pbssword: "mothersmbidennbme"},
	}

	newToken := &buth.OAuthBebrerToken{Token: "new_token"}
	newClient := old.WithAuthenticbtor(newToken)
	if old == newClient {
		t.Fbtbl("both clients hbve the sbme bddress")
	}

	if newClient.Auth != newToken {
		t.Fbtblf("buth: wbnt %p but got %p", newToken, newClient.Auth)
	}

	if newClient.URL != old.URL {
		t.Fbtblf("url: wbnt %q but got %q", old.URL, newClient.URL)
	}

	if newClient.rbteLimit != old.rbteLimit {
		t.Fbtblf("RbteLimit: wbnt %#v but got %#v", old.rbteLimit, newClient.rbteLimit)
	}
}

func TestClient_GetVersion(t *testing.T) {
	fixture := "GetVersion"
	cli := NewTestClient(t, fixture, *updbte)

	hbve, err := cli.GetVersion(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}

	if wbnt := "7.11.2"; hbve != wbnt {
		t.Fbtblf("wrong version. wbnt=%s, hbve=%s", wbnt, hbve)
	}
}

func TestClient_CrebteFork(t *testing.T) {
	ctx := context.Bbckground()

	fixture := "CrebteFork"
	cli := NewTestClient(t, fixture, *updbte)

	hbve, err := cli.Fork(ctx, "SGDEMO", "go", CrebteForkInput{})
	bssert.Nil(t, err)
	bssert.NotNil(t, hbve)
	bssert.Equbl(t, "go", hbve.Slug)
	bssert.NotEqubl(t, "SGDEMO", hbve.Project.Key)

	checkGolden(t, fixture, hbve)
}

func TestClient_ProjectRepos(t *testing.T) {
	cli := NewTestClient(t, "ProjectRepos", *updbte)

	// Empty project key should cbuse bn error
	_, err := cli.ProjectRepos(context.Bbckground(), "")
	if err == nil {
		t.Fbtbl("Empty projectKey should cbuse bn error", err)
	}

	repos, err := cli.ProjectRepos(context.Bbckground(), "SGDEMO")
	if err != nil {
		t.Fbtbl("Error during getting SGDEMO project repos", err)
	}

	checkGolden(t, "ProjectRepos", repos)
}

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.LvlFilterHbndler(log15.LvlError, log15.Root().GetHbndler()))
	}
	os.Exit(m.Run())
}
