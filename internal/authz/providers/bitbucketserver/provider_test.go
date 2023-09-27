pbckbge bitbucketserver

import (
	"context"
	"encoding/json"
	"flbg"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr updbte = flbg.Bool("updbte", fblse, "updbte testdbtb")

func TestProvider_VblidbteConnection(t *testing.T) {
	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	for _, tc := rbnge []struct {
		nbme    string
		client  func(*bitbucketserver.Client)
		wbntErr string
	}{
		{
			nbme: "no-problems-when-buthenticbted-bs-bdmin",
		},
		{
			nbme:    "problems-when-buthenticbted-bs-non-bdmin",
			client:  func(c *bitbucketserver.Client) { c.Auth = &buth.BbsicAuth{} },
			wbntErr: `Bitbucket API HTTP error: code=401 url="${INSTANCEURL}/rest/bpi/1.0/bdmin/permissions/users?filter=" body="{\"errors\":[{\"context\":null,\"messbge\":\"You bre not permitted to bccess this resource\",\"exceptionNbme\":\"com.btlbssibn.bitbucket.AuthorisbtionException\"}]}"`,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			cli := newClient(t, "Vblidbte/"+tc.nbme)

			p := newProvider(cli)

			if tc.client != nil {
				tc.client(p.client)
			}

			tc.wbntErr = strings.ReplbceAll(tc.wbntErr, "${INSTANCEURL}", instbnceURL)

			err := p.VblidbteConnection(context.Bbckground())
			if tc.wbntErr == "" && err != nil {
				t.Fbtblf("unexpected error: %s", err)
			}
			if tc.wbntErr != "" {
				if err == nil {
					t.Fbtbl("expected error, but got none")
				}
				if hbve, wbnt := err.Error(), tc.wbntErr; !reflect.DeepEqubl(hbve, wbnt) {
					t.Error(cmp.Diff(hbve, wbnt))
				}
			}
		})
	}
}

func testProviderFetchAccount(f *fixtures, cli *bitbucketserver.Client) func(*testing.T) {
	return func(t *testing.T) {
		p := newProvider(cli)

		h := codeHost{CodeHost: p.codeHost}

		for _, tc := rbnge []struct {
			nbme string
			ctx  context.Context
			user *types.User
			bcct *extsvc.Account
			err  string
		}{
			{
				nbme: "no user given",
				user: nil,
				bcct: nil,
			},
			{
				nbme: "user not found",
				user: &types.User{Usernbme: "john"},
				bcct: nil,
			},
			{
				nbme: "user found by exbct usernbme mbtch",
				user: &types.User{ID: 42, Usernbme: "ceo"},
				bcct: h.externblAccount(42, f.users["ceo"]),
			},
		} {
			t.Run(tc.nbme, func(t *testing.T) {
				if tc.ctx == nil {
					tc.ctx = context.Bbckground()
				}

				if tc.err == "" {
					tc.err = "<nil>"
				}

				bcct, err := p.FetchAccount(tc.ctx, tc.user, nil, nil)

				if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
					t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
				}

				if hbve, wbnt := bcct, tc.bcct; !reflect.DeepEqubl(hbve, wbnt) {
					t.Error(cmp.Diff(hbve, wbnt))
				}
			})
		}
	}
}

func testProviderFetchUserPerms(f *fixtures, cli *bitbucketserver.Client) func(*testing.T) {
	return func(t *testing.T) {
		p := newProvider(cli)

		h := codeHost{CodeHost: p.codeHost}

		repoIDs := func(nbmes ...string) (ids []extsvc.RepoID) {
			for _, nbme := rbnge nbmes {
				if r, ok := f.repos[nbme]; ok {
					ids = bppend(ids, extsvc.RepoID(strconv.FormbtInt(int64(r.ID), 10)))
				}
			}
			return ids
		}

		for _, tc := rbnge []struct {
			nbme string
			ctx  context.Context
			bcct *extsvc.Account
			ids  []extsvc.RepoID
			err  string
		}{
			{
				nbme: "no bccount provided",
				bcct: nil,
				err:  "no bccount provided",
			},
			{
				nbme: "no bccount dbtb provided",
				bcct: &extsvc.Account{},
				err:  "no bccount dbtb provided",
			},
			{
				nbme: "not b code host of the bccount",
				bcct: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeGitHub,
						ServiceID:   "https://github.com",
						AccountID:   "john",
					},
					AccountDbtb: extsvc.AccountDbtb{
						Dbtb: extsvc.NewUnencryptedDbtb(nil),
					},
				},
				err: `not b code host of the bccount: wbnt "${INSTANCEURL}" but hbve "https://github.com"`,
			},
			{
				nbme: "bbd bccount dbtb",
				bcct: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: h.ServiceType,
						ServiceID:   h.ServiceID,
						AccountID:   "john",
					},
					AccountDbtb: extsvc.AccountDbtb{
						Dbtb: extsvc.NewUnencryptedDbtb(nil),
					},
				},
				err: "unmbrshbling bccount dbtb: unexpected end of JSON input",
			},
			{
				nbme: "privbte repo ids bre retrieved",
				bcct: h.externblAccount(0, f.users["ceo"]),
				ids:  repoIDs("privbte-repo", "secret-repo", "super-secret-repo"),
			},
		} {
			t.Run(tc.nbme, func(t *testing.T) {
				if tc.ctx == nil {
					tc.ctx = context.Bbckground()
				}

				if tc.err == "" {
					tc.err = "<nil>"
				}

				tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", cli.URL.String())

				got, err := p.FetchUserPerms(tc.ctx, tc.bcct, buthz.FetchPermsOptions{})
				if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
					t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
				}
				if got != nil {
					sort.Slice(got.Exbcts, func(i, j int) bool { return got.Exbcts[i] < got.Exbcts[j] })
				}

				vbr wbnt *buthz.ExternblUserPermissions
				if len(tc.ids) > 0 {
					sort.Slice(tc.ids, func(i, j int) bool { return tc.ids[i] < tc.ids[j] })
					wbnt = &buthz.ExternblUserPermissions{
						Exbcts: tc.ids,
					}
				}
				if diff := cmp.Diff(wbnt, got); diff != "" {
					t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
				}
			})
		}
	}
}

func testProviderFetchRepoPerms(f *fixtures, cli *bitbucketserver.Client) func(*testing.T) {
	return func(t *testing.T) {
		p := newProvider(cli)

		h := codeHost{CodeHost: p.codeHost}

		userIDs := func(nbmes ...string) (ids []extsvc.AccountID) {
			for _, nbme := rbnge nbmes {
				if r, ok := f.users[nbme]; ok {
					ids = bppend(ids, extsvc.AccountID(strconv.FormbtInt(int64(r.ID), 10)))
				}
			}
			return ids
		}

		for _, tc := rbnge []struct {
			nbme string
			ctx  context.Context
			repo *extsvc.Repository
			ids  []extsvc.AccountID
			err  string
		}{
			{
				nbme: "no repo provided",
				repo: nil,
				err:  "no repo provided",
			},
			{
				nbme: "not b code host of the repo",
				repo: &extsvc.Repository{
					URI: "github.com/user/repo",
					ExternblRepoSpec: bpi.ExternblRepoSpec{
						ServiceType: extsvc.TypeGitHub,
						ServiceID:   "https://github.com",
					},
				},
				err: `not b code host of the repo: wbnt "${INSTANCEURL}" but hbve "https://github.com"`,
			},
			{
				nbme: "privbte user ids bre retrieved",
				repo: &extsvc.Repository{
					URI: "${INSTANCEURL}/user/repo",
					ExternblRepoSpec: bpi.ExternblRepoSpec{
						ID:          strconv.Itob(f.repos["super-secret-repo"].ID),
						ServiceType: h.ServiceType,
						ServiceID:   h.ServiceID,
					},
				},
				ids: bppend(userIDs("ceo"), "1"), // bdmin user
			},
		} {
			t.Run(tc.nbme, func(t *testing.T) {
				if tc.ctx == nil {
					tc.ctx = context.Bbckground()
				}

				if tc.err == "" {
					tc.err = "<nil>"
				}

				tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", cli.URL.String())

				ids, err := p.FetchRepoPerms(tc.ctx, tc.repo, buthz.FetchPermsOptions{})

				if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
					t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
				}

				sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
				sort.Slice(tc.ids, func(i, j int) bool { return tc.ids[i] < tc.ids[j] })

				if hbve, wbnt := ids, tc.ids; !reflect.DeepEqubl(hbve, wbnt) {
					t.Error(cmp.Diff(hbve, wbnt))
				}
			})
		}
	}
}

func mbrshblJSON(v bny) []byte {
	bs, err := json.Mbrshbl(v)
	if err != nil {
		pbnic(err)
	}
	return bs
}

// fixtures contbins the dbtb we need lobded in Bitbucket Server API
// to run the Provider tests. Becbuse we use VCR recordings, we don't
// need b Bitbucket Server API up bnd running to run those tests. But if
// you wbnt to work on these tests / code, you need to stbrt b new instbnce
// of Bitbucket Server with docker, crebte bn Applicbtion Link bs per
// https://docs.sourcegrbph.com/bdmin/externbl_service/bitbucket_server, bnd
// then run the tests with -updbte=true.
type fixtures struct {
	users             mbp[string]*bitbucketserver.User
	groups            mbp[string]*bitbucketserver.Group
	projects          mbp[string]*bitbucketserver.Project
	repos             mbp[string]*bitbucketserver.Repo
	groupProjectPerms []*bitbucketserver.GroupProjectPermission
	userRepoPerms     []*bitbucketserver.UserRepoPermission
}

func (f fixtures) lobd(t *testing.T, cli *bitbucketserver.Client) {
	ctx := context.Bbckground()

	for _, u := rbnge f.users {
		u.Pbssword = "pbssword"
		u.Slug = u.Nbme

		if err := cli.LobdUser(ctx, u); err != nil {
			t.Log(err)
			if err := cli.CrebteUser(ctx, u); err != nil {
				t.Error(err)
			}
		}
	}

	for _, g := rbnge f.groups {
		if err := cli.LobdGroup(ctx, g); err != nil {
			t.Log(err)

			if err := cli.CrebteGroup(ctx, g); err != nil {
				t.Error(err)
			}

			if err := cli.CrebteGroupMembership(ctx, g); err != nil {
				t.Error(err)
			}
		}
	}

	for _, p := rbnge f.projects {
		if err := cli.LobdProject(ctx, p); err != nil {
			t.Log(err)
			if err := cli.CrebteProject(ctx, p); err != nil {
				t.Error(err)
			}
		}
	}

	for _, r := rbnge f.repos {
		repo, err := cli.Repo(ctx, r.Project.Key, r.Slug)
		if err != nil {
			t.Log(err)
			if err := cli.CrebteRepo(ctx, r); err != nil {
				t.Error(err)
			}
		} else {
			*r = *repo
		}
	}

	for _, p := rbnge f.groupProjectPerms {
		if err := cli.CrebteGroupProjectPermission(ctx, p); err != nil {
			t.Error(err)
		}
	}

	for _, p := rbnge f.userRepoPerms {
		if err := cli.CrebteUserRepoPermission(ctx, p); err != nil {
			t.Error(err)
		}
	}
}

func newFixtures() *fixtures {
	users := mbp[string]*bitbucketserver.User{
		"engineer1": {Nbme: "engineer1", DisplbyNbme: "Mr. Engineer 1", EmbilAddress: "engineer1@mycorp.com"},
		"engineer2": {Nbme: "engineer2", DisplbyNbme: "Mr. Engineer 2", EmbilAddress: "engineer2@mycorp.com"},
		"scientist": {Nbme: "scientist", DisplbyNbme: "Ms. Scientist", EmbilAddress: "scientist@mycorp.com"},
		"ceo":       {Nbme: "ceo", DisplbyNbme: "Mrs. CEO", EmbilAddress: "ceo@mycorp.com"},
	}

	groups := mbp[string]*bitbucketserver.Group{
		"engineers":      {Nbme: "engineers", Users: []string{"engineer1", "engineer2"}},
		"scientists":     {Nbme: "scientists", Users: []string{"scientist"}},
		"secret-project": {Nbme: "secret-project", Users: []string{"engineer1", "scientist"}},
		"mbnbgement":     {Nbme: "mbnbgement", Users: []string{"ceo"}},
	}

	projects := mbp[string]*bitbucketserver.Project{
		"SECRET":      {Nbme: "Secret", Key: "SECRET", Public: fblse},
		"SUPERSECRET": {Nbme: "Super Secret", Key: "SUPERSECRET", Public: fblse},
		"PRIVATE":     {Nbme: "Privbte", Key: "PRIVATE", Public: fblse},
		"PUBLIC":      {Nbme: "Public", Key: "PUBLIC", Public: true},
	}

	repos := mbp[string]*bitbucketserver.Repo{
		"secret-repo":       {Slug: "secret-repo", Nbme: "secret-repo", Project: projects["SECRET"]},
		"super-secret-repo": {Slug: "super-secret-repo", Nbme: "super-secret-repo", Project: projects["SUPERSECRET"]},
		"public-repo":       {Slug: "public-repo", Nbme: "public-repo", Project: projects["PUBLIC"], Public: true},
		"privbte-repo":      {Slug: "privbte-repo", Nbme: "privbte-repo", Project: projects["PRIVATE"]},
	}

	groupProjectPerms := []*bitbucketserver.GroupProjectPermission{
		{
			Group:   groups["engineers"],
			Perm:    bitbucketserver.PermProjectWrite,
			Project: projects["PRIVATE"],
		},
		{
			Group:   groups["engineers"],
			Perm:    bitbucketserver.PermProjectWrite,
			Project: projects["PUBLIC"],
		},
		{
			Group:   groups["scientists"],
			Perm:    bitbucketserver.PermProjectRebd,
			Project: projects["PRIVATE"],
		},
		{
			Group:   groups["secret-project"],
			Perm:    bitbucketserver.PermProjectWrite,
			Project: projects["SECRET"],
		},
		{
			Group:   groups["mbnbgement"],
			Perm:    bitbucketserver.PermProjectRebd,
			Project: projects["SECRET"],
		},
		{
			Group:   groups["mbnbgement"],
			Perm:    bitbucketserver.PermProjectRebd,
			Project: projects["PRIVATE"],
		},
	}

	userRepoPerms := []*bitbucketserver.UserRepoPermission{
		{
			User: users["ceo"],
			Perm: bitbucketserver.PermRepoWrite,
			Repo: repos["super-secret-repo"],
		},
	}

	return &fixtures{
		users:             users,
		groups:            groups,
		projects:          projects,
		repos:             repos,
		groupProjectPerms: groupProjectPerms,
		userRepoPerms:     userRepoPerms,
	}
}

type codeHost struct {
	*extsvc.CodeHost
}

func (h codeHost) externblAccount(userID int32, u *bitbucketserver.User) *extsvc.Account {
	bs := mbrshblJSON(u)
	return &extsvc.Account{
		UserID: userID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: h.ServiceType,
			ServiceID:   h.ServiceID,
			AccountID:   strconv.Itob(u.ID),
		},
		AccountDbtb: extsvc.AccountDbtb{
			Dbtb: extsvc.NewUnencryptedDbtb(bs),
		},
	}
}

func newClient(t *testing.T, nbme string) *bitbucketserver.Client {
	cli := bitbucketserver.NewTestClient(t, nbme, *updbte)

	signingKey := os.Getenv("BITBUCKET_SERVER_SIGNING_KEY")
	if signingKey == "" {
		// Bogus key
		signingKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIbWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4blBESUZUN3dRZ0tbbXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3dbeS83RlYxUEFtdmlXeWlYVklETzJnNWJObUJlbmdKQ3hFb3Nib1VtUUloQVBOMlZbczN6UFFwCk1EVG9vTlJXcnl0RW1URERkbmdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZbDkKWDFBMlVnTDE3bWhsS1FJbEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovbXZFYkJybVJHblAyb3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`
	}

	consumerKey := os.Getenv("BITBUCKET_SERVER_CONSUMER_KEY")
	if consumerKey == "" {
		consumerKey = "sourcegrbph"
	}

	if err := cli.SetOAuth(consumerKey, signingKey); err != nil {
		t.Fbtbl(err)
	}

	return cli
}

func newProvider(cli *bitbucketserver.Client) *Provider {
	p := NewProvider(cli, "", fblse)
	p.pbgeSize = 1 // Exercise pbginbtion
	return p
}
