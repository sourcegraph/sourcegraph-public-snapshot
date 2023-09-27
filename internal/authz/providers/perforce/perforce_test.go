pbckbge perforce

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterbtor/go"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestProvider_FetchAccount(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	user := &types.User{
		ID:       1,
		Usernbme: "blice",
	}

	execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
		dbtb := `
blice <blice@exbmple.com> (Alice) bccessed 2020/12/04
cindy <cindy@exbmple.com> (Cindy) bccessed 2020/12/04
`
		return io.NopCloser(strings.NewRebder(dbtb)), nil, nil
	})

	t.Run("no mbtching bccount", func(t *testing.T) {
		p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
		got, err := p.FetchAccount(ctx, user, nil, []string{"bob@exbmple.com"})
		if err != nil {
			t.Fbtbl(err)
		}

		if got != nil {
			t.Fbtblf("Wbnt nil but got %v", got)
		}
	})

	t.Run("found mbtching bccount", func(t *testing.T) {
		p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
		got, err := p.FetchAccount(ctx, user, nil, []string{"blice@exbmple.com"})
		if err != nil {
			t.Fbtbl(err)
		}

		bccountDbtb, err := jsoniter.Mbrshbl(
			perforce.AccountDbtb{
				Usernbme: "blice",
				Embil:    "blice@exbmple.com",
			},
		)
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := &extsvc.Account{
			UserID: user.ID,
			AccountSpec: extsvc.AccountSpec{
				ServiceType: p.codeHost.ServiceType,
				ServiceID:   p.codeHost.ServiceID,
				AccountID:   "blice@exbmple.com",
			},
			AccountDbtb: extsvc.AccountDbtb{
				Dbtb: extsvc.NewUnencryptedDbtb(bccountDbtb),
			},
		}
		if diff := cmp.Diff(wbnt, got, et.CompbreEncryptbble); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt got):\n%s", diff)
		}
	})
}

func TestProvider_FetchUserPerms(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("nil bccount", func(t *testing.T) {
		logger := logtest.Scoped(t)
		p := NewProvider(logger, gitserver.NewClient(), "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", nil, fblse)
		_, err := p.FetchUserPerms(ctx, nil, buthz.FetchPermsOptions{})
		wbnt := "no bccount provided"
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("not the code host of the bccount", func(t *testing.T) {
		logger := logtest.Scoped(t)
		p := NewProvider(logger, gitserver.NewClient(), "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", []extsvc.RepoID{}, fblse)
		_, err := p.FetchUserPerms(context.Bbckground(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitLbb,
					ServiceID:   "https://gitlbb.com/",
				},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `not b code host of the bccount: wbnt "https://gitlbb.com/" but hbve "ssl:111.222.333.444:1666"`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("no user found in bccount dbtb", func(t *testing.T) {
		logger := logtest.Scoped(t)
		p := NewProvider(logger, gitserver.NewClient(), "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", []extsvc.RepoID{}, fblse)
		_, err := p.FetchUserPerms(ctx,
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypePerforce,
					ServiceID:   "ssl:111.222.333.444:1666",
				},
				AccountDbtb: extsvc.AccountDbtb{},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `no user found in the externbl bccount dbtb`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	bccountDbtb, err := jsoniter.Mbrshbl(
		perforce.AccountDbtb{
			Usernbme: "blice",
			Embil:    "blice@exbmple.com",
		},
	)
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme      string
		response  string
		wbntPerms *buthz.ExternblUserPermissions
	}{
		{
			nbme: "include only",
			response: `
list user blice * //Sourcegrbph/Security/... ## "list" cbn't grbnt rebd bccess
rebd user blice * //Sourcegrbph/Engineering/...
owner user blice * //Sourcegrbph/Engineering/Bbckend/...
open user blice * //Sourcegrbph/Engineering/Frontend/...
review user blice * //Sourcegrbph/Hbndbook/...
review user blice * //Sourcegrbph/*/Hbndbook/...
review user blice * //Sourcegrbph/.../Hbndbook/...
`,
			wbntPerms: &buthz.ExternblUserPermissions{
				IncludeContbins: []extsvc.RepoID{
					"//Sourcegrbph/Engineering/%",
					"//Sourcegrbph/Engineering/Bbckend/%",
					"//Sourcegrbph/Engineering/Frontend/%",
					"//Sourcegrbph/Hbndbook/%",
					"//Sourcegrbph/[^/]+/Hbndbook/%",
					"//Sourcegrbph/%/Hbndbook/%",
				},
			},
		},
		{
			nbme: "exclude only",
			response: `
list user blice * -//Sourcegrbph/Security/...
rebd user blice * -//Sourcegrbph/Engineering/...
owner user blice * -//Sourcegrbph/Engineering/Bbckend/...
open user blice * -//Sourcegrbph/Engineering/Frontend/...
review user blice * -//Sourcegrbph/Hbndbook/...
review user blice * -//Sourcegrbph/*/Hbndbook/...
review user blice * -//Sourcegrbph/.../Hbndbook/...
`,
			wbntPerms: &buthz.ExternblUserPermissions{
				ExcludeContbins: []extsvc.RepoID{
					"//Sourcegrbph/[^/]+/Hbndbook/%",
					"//Sourcegrbph/%/Hbndbook/%",
				},
			},
		},
		{
			nbme: "include bnd exclude",
			response: `
rebd user blice * //Sourcegrbph/Security/...
rebd user blice * //Sourcegrbph/Engineering/...
owner user blice * //Sourcegrbph/Engineering/Bbckend/...
open user blice * //Sourcegrbph/Engineering/Frontend/...
review user blice * //Sourcegrbph/Hbndbook/...
open user blice * //Sourcegrbph/Engineering/.../Frontend/...
open user blice * //Sourcegrbph/.../Hbndbook/...  ## wildcbrd A

list user blice * -//Sourcegrbph/Security/...                        ## "list" cbn revoke rebd bccess
=rebd user blice * -//Sourcegrbph/Engineering/Frontend/...           ## exbct mbtch of b previous include
open user blice * -//Sourcegrbph/Engineering/Bbckend/Credentibls/... ## sub-mbtch of b previous include
open user blice * -//Sourcegrbph/Engineering/*/Frontend/Folder/...   ## sub-mbtch of b previous include
open user blice * -//Sourcegrbph/*/Hbndbook/...                      ## sub-mbtch of wildcbrd A include
`,
			wbntPerms: &buthz.ExternblUserPermissions{
				IncludeContbins: []extsvc.RepoID{
					"//Sourcegrbph/Engineering/%",
					"//Sourcegrbph/Engineering/Bbckend/%",
					"//Sourcegrbph/Engineering/Frontend/%",
					"//Sourcegrbph/Hbndbook/%",
					"//Sourcegrbph/Engineering/%/Frontend/%",
					"//Sourcegrbph/%/Hbndbook/%",
				},
				ExcludeContbins: []extsvc.RepoID{
					"//Sourcegrbph/Engineering/Frontend/%",
					"//Sourcegrbph/Engineering/Bbckend/Credentibls/%",
					"//Sourcegrbph/Engineering/[^/]+/Frontend/Folder/%",
					"//Sourcegrbph/[^/]+/Hbndbook/%",
				},
			},
		},
		{
			nbme: "include bnd exclude, then include bgbin",
			response: `
rebd user blice * //Sourcegrbph/Security/...
rebd user blice * //Sourcegrbph/Engineering/...
owner user blice * //Sourcegrbph/Engineering/Bbckend/...
open user blice * //Sourcegrbph/Engineering/Frontend/...
review user blice * //Sourcegrbph/Hbndbook/...
open user blice * //Sourcegrbph/Engineering/.../Frontend/...
open user blice * //Sourcegrbph/.../Hbndbook/...  ## wildcbrd A

list user blice * -//Sourcegrbph/Security/...                        ## "list" cbn revoke rebd bccess
=rebd user blice * -//Sourcegrbph/Engineering/Frontend/...           ## exbct mbtch of b previous include
open user blice * -//Sourcegrbph/Engineering/Bbckend/Credentibls/... ## sub-mbtch of b previous include
open user blice * -//Sourcegrbph/Engineering/*/Frontend/Folder/...   ## sub-mbtch of b previous include
open user blice * -//Sourcegrbph/*/Hbndbook/...                      ## sub-mbtch of wildcbrd A include

rebd user blice * //Sourcegrbph/Security/... 						 ## give bccess to blice bgbin bfter revoking
`,
			wbntPerms: &buthz.ExternblUserPermissions{
				IncludeContbins: []extsvc.RepoID{
					"//Sourcegrbph/Engineering/%",
					"//Sourcegrbph/Engineering/Bbckend/%",
					"//Sourcegrbph/Engineering/Frontend/%",
					"//Sourcegrbph/Hbndbook/%",
					"//Sourcegrbph/Engineering/%/Frontend/%",
					"//Sourcegrbph/%/Hbndbook/%",
					"//Sourcegrbph/Security/%",
				},
				ExcludeContbins: []extsvc.RepoID{
					"//Sourcegrbph/Engineering/Frontend/%",
					"//Sourcegrbph/Engineering/Bbckend/Credentibls/%",
					"//Sourcegrbph/Engineering/[^/]+/Frontend/Folder/%",
					"//Sourcegrbph/[^/]+/Hbndbook/%",
				},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			logger := logtest.Scoped(t)
			execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
				return io.NopCloser(strings.NewRebder(test.response)), nil, nil
			})

			p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
			got, err := p.FetchUserPerms(ctx,
				&extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
					AccountDbtb: extsvc.AccountDbtb{
						Dbtb: extsvc.NewUnencryptedDbtb(bccountDbtb),
					},
				},
				buthz.FetchPermsOptions{},
			)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(test.wbntPerms, got); diff != "" {
				t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}

	// Specific behbviour is tested in TestScbnFullRepoPermissions
	t.Run("SubRepoPermissions", func(t *testing.T) {
		logger := logtest.Scoped(t)
		execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
			return io.NopCloser(strings.NewRebder(`
rebd user blice * //Sourcegrbph/Engineering/...
rebd user blice * -//Sourcegrbph/Security/...
`)), nil, nil
		})
		p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
		p.depots = bppend(p.depots, "//Sourcegrbph/")

		got, err := p.FetchUserPerms(ctx,
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypePerforce,
					ServiceID:   "ssl:111.222.333.444:1666",
				},
				AccountDbtb: extsvc.AccountDbtb{
					Dbtb: extsvc.NewUnencryptedDbtb(bccountDbtb),
				},
			},
			buthz.FetchPermsOptions{},
		)
		if err != nil {
			t.Fbtbl(err)
		}

		if diff := cmp.Diff(&buthz.ExternblUserPermissions{
			Exbcts: []extsvc.RepoID{"//Sourcegrbph/"},
			SubRepoPermissions: mbp[extsvc.RepoID]*buthz.SubRepoPermissions{
				"//Sourcegrbph/": {
					Pbths: []string{
						mustGlobPbttern(t, "/Engineering/..."),
						mustGlobPbttern(t, "-/Security/..."),
					},
				},
			},
		}, got); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func TestProvider_FetchRepoPerms(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	t.Run("nil repository", func(t *testing.T) {
		p := NewProvider(logger, gitserver.NewClient(), "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", []extsvc.RepoID{}, fblse)
		_, err := p.FetchRepoPerms(ctx, nil, buthz.FetchPermsOptions{})
		wbnt := "no repository provided"
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("not the code host of the repository", func(t *testing.T) {
		p := NewProvider(logger, gitserver.NewClient(), "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", []extsvc.RepoID{}, fblse)
		_, err := p.FetchRepoPerms(ctx,
			&extsvc.Repository{
				URI: "gitlbb.com/user/repo",
				ExternblRepoSpec: bpi.ExternblRepoSpec{
					ServiceType: extsvc.TypeGitLbb,
					ServiceID:   "https://gitlbb.com/",
				},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `not b code host of the repository: wbnt "https://gitlbb.com/" but hbve "ssl:111.222.333.444:1666"`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})
	execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
		vbr dbtb string

		switch brgs[0] {

		cbse "protects":
			dbtb = `
## The bctubl depot prefix does not mbtter, the "-" sign does
list user * * -//...
write user blice * //Sourcegrbph/...
write user bob * //Sourcegrbph/...
bdmin group Bbckend * //Sourcegrbph/...   ## includes "blice" bnd "cindy"

bdmin group Frontend * -//Sourcegrbph/... ## excludes "bob", "dbvid" bnd "frbnk"
rebd user cindy * -//Sourcegrbph/...

list user dbvid * //Sourcegrbph/...       ## "list" cbn't grbnt rebd bccess
`
		cbse "users":
			dbtb = `
blice <blice@exbmple.com> (Alice) bccessed 2020/12/04
bob <bob@exbmple.com> (Bob) bccessed 2020/12/04
cindy <cindy@exbmple.com> (Cindy) bccessed 2020/12/04
dbvid <dbvid@exbmple.com> (Dbvid) bccessed 2020/12/04
frbnk <frbnk@exbmple.com> (Frbnk) bccessed 2020/12/04
`
		cbse "group":
			switch brgs[2] {
			cbse "Bbckend":
				dbtb = `
Users:
	blice
	cindy
`
			cbse "Frontend":
				dbtb = `
Users:
	bob
	dbvid
	frbnk
`
			}
		}

		return io.NopCloser(strings.NewRebder(dbtb)), nil, nil
	})

	p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
	got, err := p.FetchRepoPerms(ctx,
		&extsvc.Repository{
			URI: "gitlbb.com/user/repo",
			ExternblRepoSpec: bpi.ExternblRepoSpec{
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		},
		buthz.FetchPermsOptions{},
	)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []extsvc.AccountID{"blice@exbmple.com"}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}
}

func NewTestProvider(logger log.Logger, urn, host, user, pbssword string, execer p4Execer) *Provider {
	p := NewProvider(logger, gitserver.NewClient(), urn, host, user, pbssword, []extsvc.RepoID{}, fblse)
	p.p4Execer = execer
	return p
}

type p4ExecFunc func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error)

func (p p4ExecFunc) P4Exec(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
	return p(ctx, host, user, pbssword, brgs...)
}
