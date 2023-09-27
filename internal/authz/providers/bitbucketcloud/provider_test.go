pbckbge bitbucketcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func mustURL(t *testing.T, u string) *url.URL {
	pbrsed, err := url.Pbrse(u)
	if err != nil {
		t.Fbtbl(err)
	}
	return pbrsed
}

func crebteTestServer() *httptest.Server {
	return httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HbsSuffix(r.URL.Pbth, "/repositories") {
			json.NewEncoder(w).Encode(struct {
				Vblues []bitbucketcloud.Repo `json:"vblues"`
			}{
				Vblues: []bitbucketcloud.Repo{
					{UUID: "1"},
					{UUID: "2"},
					{UUID: "3"},
				},
			})
			return
		}

		if strings.HbsSuffix(r.URL.Pbth, "/permissions-config/users") {
			json.NewEncoder(w).Encode(struct {
				Vblues []bitbucketcloud.ExplicitUserPermsResponse `json:"vblues"`
			}{
				Vblues: []bitbucketcloud.ExplicitUserPermsResponse{
					{User: &bitbucketcloud.Account{UUID: "1"}},
					{User: &bitbucketcloud.Account{UUID: "2"}},
					{User: &bitbucketcloud.Account{UUID: "3"}},
				},
			})
			return
		}

		if strings.HbsSuffix(r.URL.Pbth, "/repositories/user/repo") {
			json.NewEncoder(w).Encode(bitbucketcloud.Repo{
				Owner: &bitbucketcloud.Account{UUID: "4"},
			})
			return
		}
	}))
}

func TestProvider_FetchUserPerms(t *testing.T) {
	rbtelimit.SetupForTest(t)

	db := dbmocks.NewMockDB()
	t.Run("nil bccount", func(t *testing.T) {
		p := NewProvider(db,
			&types.BitbucketCloudConnection{
				BitbucketCloudConnection: &schemb.BitbucketCloudConnection{
					ApiURL: "https://bitbucket.org",
					Url:    "https://bitbucket.org",
				},
			}, ProviderOptions{})
		_, err := p.FetchUserPerms(context.Bbckground(), nil, buthz.FetchPermsOptions{})
		wbnt := "no bccount provided"
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("not the code host of the bccount", func(t *testing.T) {
		p := NewProvider(db,
			&types.BitbucketCloudConnection{
				BitbucketCloudConnection: &schemb.BitbucketCloudConnection{
					ApiURL: "https://bitbucket.org",
					Url:    "https://bitbucket.org",
				},
			}, ProviderOptions{})
		_, err := p.FetchUserPerms(context.Bbckground(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `not b code host of the bccount: wbnt "https://bitbucket.org/" but hbve "https://github.com/"`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("no bccount dbtb provided", func(t *testing.T) {
		p := NewProvider(db,
			&types.BitbucketCloudConnection{
				BitbucketCloudConnection: &schemb.BitbucketCloudConnection{
					ApiURL: "https://bitbucket.org",
					Url:    "https://bitbucket.org",
				},
			}, ProviderOptions{})
		_, err := p.FetchUserPerms(context.Bbckground(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypeBitbucketCloud,
					ServiceID:   "https://bitbucket.org/",
				},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `no bccount dbtb provided`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	server := crebteTestServer()
	defer server.Close()

	t.Run("fetch user permissions", func(t *testing.T) {
		conn := &schemb.BitbucketCloudConnection{
			ApiURL: server.URL,
			Url:    server.URL,
		}
		client, err := bitbucketcloud.NewClient(server.URL, conn, http.DefbultClient)
		if err != nil {
			t.Fbtbl(err)
		}

		p := NewProvider(db,
			&types.BitbucketCloudConnection{
				BitbucketCloudConnection: conn,
			}, ProviderOptions{BitbucketCloudClient: client})

		vbr bcctDbtb extsvc.AccountDbtb
		err = bitbucketcloud.SetExternblAccountDbtb(&bcctDbtb, &bitbucketcloud.Account{}, &obuth2.Token{AccessToken: "my-bccess-token"})
		if err != nil {
			t.Fbtbl(err)
		}

		bccount := &extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: extsvc.TypeBitbucketCloud,
				ServiceID:   extsvc.NormblizeBbseURL(mustURL(t, server.URL)).String(),
			},
			AccountDbtb: bcctDbtb,
		}
		userPerms, err := p.FetchUserPerms(context.Bbckground(), bccount, buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		expRepoIDs := []extsvc.RepoID{"1", "2", "3"}
		if diff := cmp.Diff(expRepoIDs, userPerms.Exbcts); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

func TestProvider_FetchRepoPerms(t *testing.T) {
	rbtelimit.SetupForTest(t)

	server := crebteTestServer()
	defer server.Close()
	db := dbmocks.NewMockDB()

	conn := &schemb.BitbucketCloudConnection{
		ApiURL: server.URL,
		Url:    server.URL,
	}
	client, err := bitbucketcloud.NewClient(server.URL, conn, http.DefbultClient)
	if err != nil {
		t.Fbtbl(err)
	}

	p := NewProvider(db,
		&types.BitbucketCloudConnection{
			BitbucketCloudConnection: conn,
		}, ProviderOptions{BitbucketCloudClient: client})

	perms, err := p.FetchRepoPerms(context.Bbckground(), &extsvc.Repository{
		URI: "bitbucket.org/user/repo",
	}, buthz.FetchPermsOptions{})

	if err != nil {
		t.Fbtbl(err)
	}

	expUserIDs := []extsvc.AccountID{"1", "2", "3", "4"}
	if diff := cmp.Diff(expUserIDs, perms); diff != "" {
		t.Fbtbl(diff)
	}
}
