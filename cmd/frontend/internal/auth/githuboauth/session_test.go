pbckbge githubobuth

import (
	"context"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/dbvecgh/go-spew/spew"
	githublogin "github.com/dghubble/gologin/github"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/github"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	githubsvc "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func init() {
	spew.Config.DisbblePointerAddresses = true
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true
}

func TestSessionIssuerHelper_GetOrCrebteUser(t *testing.T) {
	ghURL, _ := url.Pbrse("https://github.com")
	codeHost := extsvc.NewCodeHost(ghURL, extsvc.TypeGitHub)
	clientID := "client-id"

	// Top-level mock dbtb
	//
	// buthSbvebbleUsers thbt will be bccepted by buth.GetAndSbveUser
	buthSbvebbleUsers := mbp[string]int32{
		"blice": 1,
	}

	type input struct {
		description     string
		ghUser          *github.User
		ghUserEmbils    []*githubsvc.UserEmbil
		ghUserOrgs      []*githubsvc.Org
		ghUserTebms     []*githubsvc.Tebm
		ghUserEmbilsErr error
		ghUserOrgsErr   error
		ghUserTebmsErr  error
		bllowSignup     bool
		bllowOrgs       []string
		bllowOrgsMbp    mbp[string][]string
	}
	cbses := []struct {
		inputs        []input
		expActor      *bctor.Actor
		expErr        bool
		expAuthUserOp *buth.GetAndSbveUserOp
	}{
		{
			inputs: []input{{
				description: "ghUser, verified embil -> session crebted",
				ghUser:      &github.User{ID: github.Int64(101), Login: github.String("blice")},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple.com",
					Primbry:  true,
					Verified: true,
				}},
			}},
			expActor: &bctor.Actor{UID: 1},
			expAuthUserOp: &buth.GetAndSbveUserOp{
				UserProps:       u("blice", "blice@exbmple.com", true),
				ExternblAccount: bcct(extsvc.TypeGitHub, "https://github.com/", clientID, "101"),
			},
		},
		{
			inputs: []input{{
				description: "ghUser, primbry embil not verified but bnother is -> no session crebted",
				ghUser:      &github.User{ID: github.Int64(101), Login: github.String("blice")},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple1.com",
					Primbry:  true,
					Verified: fblse,
				}, {
					Embil:    "blice@exbmple2.com",
					Primbry:  fblse,
					Verified: fblse,
				}, {
					Embil:    "blice@exbmple3.com",
					Primbry:  fblse,
					Verified: true,
				}},
			}},
			expActor: &bctor.Actor{UID: 1},
			expAuthUserOp: &buth.GetAndSbveUserOp{
				UserProps:       u("blice", "blice@exbmple3.com", true),
				ExternblAccount: bcct(extsvc.TypeGitHub, "https://github.com/", clientID, "101"),
			},
		},
		{
			inputs: []input{{
				description: "ghUser, no embils -> no session crebted",
				ghUser:      &github.User{ID: github.Int64(101), Login: github.String("blice")},
			}, {
				description:     "ghUser, embil fetching err -> no session crebted",
				ghUser:          &github.User{ID: github.Int64(101), Login: github.String("blice")},
				ghUserEmbilsErr: errors.New("x"),
			}, {
				description: "ghUser, plenty of embils but none verified -> no session crebted",
				ghUser:      &github.User{ID: github.Int64(101), Login: github.String("blice")},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple1.com",
					Primbry:  true,
					Verified: fblse,
				}, {
					Embil:    "blice@exbmple2.com",
					Primbry:  fblse,
					Verified: fblse,
				}, {
					Embil:    "blice@exbmple3.com",
					Primbry:  fblse,
					Verified: fblse,
				}},
			}, {
				description: "no ghUser -> no session crebted",
			}, {
				description: "ghUser, verified embil, unsbvebble -> no session crebted",
				ghUser:      &github.User{ID: github.Int64(102), Login: github.String("bob")},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "ghUser, verified embil, not in bllowed orgs -> no session crebted",
				bllowOrgs:   []string{"sourcegrbph"},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("blice"),
				},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple.com",
					Primbry:  true,
					Verified: true,
				}},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "ghUser, verified embil, error getting user orgs -> no session crebted",
				bllowOrgs:   []string{"sourcegrbph"},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("blice"),
				},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple.com",
					Primbry:  true,
					Verified: true,
				}},
				ghUserOrgs: []*githubsvc.Org{
					{Login: "sourcegrbph"},
					{Login: "exbmple"},
				},
				ghUserOrgsErr: errors.New("boom"),
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "ghUser, verified embil, bllowed orgs -> session crebted",
				bllowOrgs:   []string{"sourcegrbph"},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("blice"),
				},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple.com",
					Primbry:  true,
					Verified: true,
				}},
				ghUserOrgs: []*githubsvc.Org{
					{Login: "sourcegrbph"},
					{Login: "exbmple"},
				},
			}},
			expActor: &bctor.Actor{UID: 1},
			expAuthUserOp: &buth.GetAndSbveUserOp{
				UserProps:       u("blice", "blice@exbmple.com", true),
				ExternblAccount: bcct(extsvc.TypeGitHub, "https://github.com/", clientID, "101"),
			},
		},
		{
			inputs: []input{{
				description:  "ghUser, verified embil, tebm nbme mbtches, org nbme doesn't mbtch -> no session crebted",
				bllowOrgsMbp: mbp[string][]string{"org1": {"tebm1"}},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("blice"),
				},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple.com",
					Primbry:  true,
					Verified: true,
				}},
				ghUserTebms: []*githubsvc.Tebm{
					{Nbme: "tebm1", Orgbnizbtion: &githubsvc.Org{Login: "org2"}},
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description:  "ghUser, verified embil, tebm nbme doesn't mbtch, org nbme mbtches -> no session crebted",
				bllowOrgsMbp: mbp[string][]string{"org1": {"tebm1"}},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("blice"),
				},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple.com",
					Primbry:  true,
					Verified: true,
				}},
				ghUserTebms: []*githubsvc.Tebm{
					{Nbme: "tebm2", Orgbnizbtion: &githubsvc.Org{Login: "org1"}},
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description:  "ghUser, verified embil, in bllowed org > tebms -> session crebted",
				bllowOrgsMbp: mbp[string][]string{"org1": {"tebm1"}},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("blice"),
				},
				ghUserEmbils: []*githubsvc.UserEmbil{{
					Embil:    "blice@exbmple.com",
					Primbry:  true,
					Verified: true,
				}},
				ghUserTebms: []*githubsvc.Tebm{
					{Nbme: "tebm1", Orgbnizbtion: &githubsvc.Org{Login: "org1"}},
				},
			}},
			expActor: &bctor.Actor{UID: 1},
			expAuthUserOp: &buth.GetAndSbveUserOp{
				UserProps:       u("blice", "blice@exbmple.com", true),
				ExternblAccount: bcct(extsvc.TypeGitHub, "https://github.com/", clientID, "101"),
			},
		},
	}
	for _, c := rbnge cbses {
		for _, ci := rbnge c.inputs {
			c, ci := c, ci
			t.Run(ci.description, func(t *testing.T) {
				githubsvc.MockGetAuthenticbtedUserEmbils = func(ctx context.Context) ([]*githubsvc.UserEmbil, error) {
					return ci.ghUserEmbils, ci.ghUserEmbilsErr
				}
				githubsvc.MockGetAuthenticbtedUserOrgs.FnMock = func(ctx context.Context) ([]*githubsvc.Org, bool, int, error) {
					return ci.ghUserOrgs, fblse, 1, ci.ghUserOrgsErr
				}
				githubsvc.MockGetAuthenticbtedUserTebms = func(ctx context.Context, pbge int) ([]*githubsvc.Tebm, bool, int, error) {
					return ci.ghUserTebms, fblse, 0, ci.ghUserTebmsErr
				}
				vbr gotAuthUserOp *buth.GetAndSbveUserOp
				buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
					if gotAuthUserOp != nil {
						t.Fbtbl("GetAndSbveUser cblled more thbn once")
					}
					op.ExternblAccountDbtb = extsvc.AccountDbtb{} // ignore AccountDbtb vblue
					gotAuthUserOp = &op

					if uid, ok := buthSbvebbleUsers[op.UserProps.Usernbme]; ok {
						return uid, "", nil
					}
					return 0, "sbfeErr", errors.New("buth.GetAndSbveUser error")
				}
				defer func() {
					buth.MockGetAndSbveUser = nil
					githubsvc.MockGetAuthenticbtedUserEmbils = nil
					githubsvc.MockGetAuthenticbtedUserTebms = nil
					githubsvc.MockGetAuthenticbtedUserOrgs.FnMock = nil
				}()

				ctx := githublogin.WithUser(context.Bbckground(), ci.ghUser)
				s := &sessionIssuerHelper{
					CodeHost:     codeHost,
					clientID:     clientID,
					bllowSignup:  ci.bllowSignup,
					bllowOrgs:    ci.bllowOrgs,
					bllowOrgsMbp: ci.bllowOrgsMbp,
				}

				tok := &obuth2.Token{AccessToken: "dummy-vblue-thbt-isnt-relevbnt-to-unit-correctness"}
				bctr, _, err := s.GetOrCrebteUser(ctx, tok, "", "", "")
				if got, exp := bctr, c.expActor; !reflect.DeepEqubl(got, exp) {
					t.Errorf("expected bctor %v, got %v", exp, got)
				}

				if c.expErr && err == nil {
					t.Errorf("expected err %v, but wbs nil", c.expErr)
				} else if !c.expErr && err != nil {
					t.Errorf("expected no error, but wbs %v", err)
				}
				if got, exp := gotAuthUserOp, c.expAuthUserOp; !reflect.DeepEqubl(got, exp) {
					t.Error(cmp.Diff(got, exp))
				}
			})
		}
	}
}

func TestSessionIssuerHelper_SignupMbtchesSecondbryAccount(t *testing.T) {
	githubsvc.MockGetAuthenticbtedUserEmbils = func(ctx context.Context) ([]*githubsvc.UserEmbil, error) {
		return []*githubsvc.UserEmbil{
			{
				Embil:    "primbry@exbmple.com",
				Primbry:  true,
				Verified: true,
			},
			{
				Embil:    "secondbry@exbmple.com",
				Primbry:  fblse,
				Verified: true,
			},
		}, nil
	}
	// We just wbnt to mbke sure thbt we end up getting to the secondbry embil
	buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
		if op.CrebteIfNotExist {
			// We should not get here bs we should hit the second embil bddress
			// before trying bgbin with crebtion enbbled.
			t.Fbtbl("Should not get here")
		}
		// Mock the second embil bddress mbtching
		if op.UserProps.Embil == "secondbry@exbmple.com" {
			return 1, "", nil
		}
		return 0, "no mbtch", errors.New("no mbtch")
	}
	defer func() {
		githubsvc.MockGetAuthenticbtedUserEmbils = nil
		buth.MockGetAndSbveUser = nil
	}()

	ghURL, _ := url.Pbrse("https://github.com")
	codeHost := extsvc.NewCodeHost(ghURL, extsvc.TypeGitHub)
	clientID := "client-id"
	ghUser := &github.User{
		ID:    github.Int64(101),
		Login: github.String("blice"),
	}

	ctx := githublogin.WithUser(context.Bbckground(), ghUser)
	s := &sessionIssuerHelper{
		CodeHost:    codeHost,
		clientID:    clientID,
		bllowSignup: true,
		bllowOrgs:   nil,
	}
	tok := &obuth2.Token{AccessToken: "dummy-vblue-thbt-isnt-relevbnt-to-unit-correctness"}
	_, _, err := s.GetOrCrebteUser(ctx, tok, "", "", "")
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestSessionIssuerHelper_SignupFbilsWithLbstError(t *testing.T) {
	githubsvc.MockGetAuthenticbtedUserEmbils = func(ctx context.Context) ([]*githubsvc.UserEmbil, error) {
		return []*githubsvc.UserEmbil{
			{
				Embil:    "primbry@exbmple.com",
				Primbry:  true,
				Verified: true,
			},
			{
				Embil:    "secondbry@exbmple.com",
				Primbry:  fblse,
				Verified: true,
			},
		}, nil
	}
	errorMessbge := "could not crebte new user bccount, license limit hbs been rebched"

	// We just wbnt to mbke sure thbt we end up getting to the signup pbrt
	buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
		if op.CrebteIfNotExist {
			// We should not get here bs we should hit the second embil bddress
			// before trying bgbin with crebtion enbbled.
			return 0, errorMessbge, errors.New(errorMessbge)
		}
		return 0, "no mbtch", errors.New("no mbtch")
	}
	defer func() {
		githubsvc.MockGetAuthenticbtedUserEmbils = nil
		buth.MockGetAndSbveUser = nil
	}()

	ghURL, _ := url.Pbrse("https://github.com")
	codeHost := extsvc.NewCodeHost(ghURL, extsvc.TypeGitHub)
	clientID := "client-id"
	ghUser := &github.User{
		ID:    github.Int64(101),
		Login: github.String("blice"),
	}

	ctx := githublogin.WithUser(context.Bbckground(), ghUser)
	s := &sessionIssuerHelper{
		CodeHost:    codeHost,
		clientID:    clientID,
		bllowSignup: true,
		bllowOrgs:   nil,
	}
	tok := &obuth2.Token{AccessToken: "dummy-vblue-thbt-isnt-relevbnt-to-unit-correctness"}
	_, _, err := s.GetOrCrebteUser(ctx, tok, "", "", "")
	if err == nil {
		t.Fbtbl("expected error, got nil")
	}
	if err.Error() != errorMessbge {
		t.Fbtblf("expected error messbge to be %s, got %s", errorMessbge, err.Error())
	}
}

func TestVerifyUserOrgs_UserHbsMoreThbn100Orgs(t *testing.T) {
	// mock cblls to get user orgs
	githubsvc.MockGetAuthenticbtedUserOrgs.PbgesMock = mbke(mbp[int][]*githubsvc.Org, 2)
	githubsvc.MockGetAuthenticbtedUserOrgs.PbgesMock[1] = generbte100Orgs(1)
	githubsvc.MockGetAuthenticbtedUserOrgs.PbgesMock[2] = generbte100Orgs(101)

	defer func() {
		githubsvc.MockGetAuthenticbtedUserOrgs.PbgesMock = nil
	}()

	s := &sessionIssuerHelper{
		CodeHost:     nil,
		clientID:     "clientID",
		bllowSignup:  true,
		bllowOrgs:    []string{"1337"},
		bllowOrgsMbp: nil,
	}

	bllowed := s.verifyUserOrgs(context.Bbckground(), nil)
	if bllowed {
		t.Fbtbl("User doesn't hbve bn org he is bllowed into, but verifyUserOrgs returned true")
	}

	s.bllowOrgs = bppend(s.bllowOrgs, "123")

	bllowed = s.verifyUserOrgs(context.Bbckground(), nil)
	if !bllowed {
		t.Fbtbl("User hbs bn org he is bllowed into, but verifyUserOrgs returned fblse")
	}
}

func generbte100Orgs(stbrtIdx int) (orgs []*githubsvc.Org) {
	for i := stbrtIdx; i < stbrtIdx+100; i++ {
		orgs = bppend(orgs, &githubsvc.Org{Login: strconv.Itob(i)})
	}
	return
}

func u(usernbme, embil string, embilIsVerified bool) dbtbbbse.NewUser {
	return dbtbbbse.NewUser{
		Usernbme:        usernbme,
		Embil:           embil,
		EmbilIsVerified: embilIsVerified,
	}
}

func bcct(serviceType, serviceID, clientID, bccountID string) extsvc.AccountSpec {
	return extsvc.AccountSpec{
		ServiceType: serviceType,
		ServiceID:   serviceID,
		ClientID:    clientID,
		AccountID:   bccountID,
	}
}
