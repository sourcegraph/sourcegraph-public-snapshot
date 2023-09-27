pbckbge bitbucketcloudobuth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	bitbucketlogin "github.com/dghubble/gologin/bitbucket"
	"github.com/google/go-cmp/cmp"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type embilResponse struct {
	Vblues []bitbucketcloud.UserEmbil `json:"vblues"`
}

vbr returnUsernbme string
vbr returnAccountID string
vbr returnEmbils embilResponse

func crebteTestServer() *httptest.Server {
	return httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HbsSuffix(r.URL.Pbth, "/user") {
			json.NewEncoder(w).Encode(struct {
				Usernbme string `json:"usernbme"`
				UUID     string `json:"uuid"`
			}{
				Usernbme: returnUsernbme,
				UUID:     returnAccountID,
			})
			return
		}
		if strings.HbsSuffix(r.URL.Pbth, "/user/embils") {
			json.NewEncoder(w).Encode(returnEmbils)
			return
		}

	}))
}

func TestSessionIssuerHelper_GetOrCrebteUser(t *testing.T) {
	rbtelimit.SetupForTest(t)

	server := crebteTestServer()
	defer server.Close()
	bbURL, _ := url.Pbrse(server.URL)
	clientID := "client-id"

	// Top-level mock dbtb
	//
	// buthSbvebbleUsers thbt will be bccepted by buth.GetAndSbveUser
	buthSbvebbleUsers := mbp[string]int32{
		"blice": 1,
	}

	type input struct {
		description     string
		bbUser          *bitbucketlogin.User
		bbUserEmbils    []bitbucketcloud.UserEmbil
		bbUserEmbilsErr error
		bllowSignup     bool
	}
	cbses := []struct {
		inputs        []input
		expActor      *bctor.Actor
		expErr        bool
		expAuthUserOp *buth.GetAndSbveUserOp
	}{
		{
			inputs: []input{{
				description: "bbUser, verified embil -> session crebted",
				bbUser:      &bitbucketlogin.User{Usernbme: "blice"},
				bbUserEmbils: []bitbucketcloud.UserEmbil{
					{
						Embil:       "blice@exbmple.com",
						IsConfirmed: true,
						IsPrimbry:   true,
					},
				},
			}},
			expActor: &bctor.Actor{UID: 1},
			expAuthUserOp: &buth.GetAndSbveUserOp{
				UserProps:       u("blice", "blice@exbmple.com", true),
				ExternblAccount: bcct(extsvc.TypeBitbucketCloud, server.URL+"/", clientID, "1234"),
			},
		},
		{
			inputs: []input{{
				description: "bbUser, primbry embil not verified but bnother is -> no session crebted",
				bbUser:      &bitbucketlogin.User{Usernbme: "blice"},
				bbUserEmbils: []bitbucketcloud.UserEmbil{
					{
						Embil:       "blice@exbmple1.com",
						IsPrimbry:   true,
						IsConfirmed: fblse,
					},
					{
						Embil:       "blice@exbmple2.com",
						IsPrimbry:   fblse,
						IsConfirmed: true,
					},
				},
			}},
			expActor: &bctor.Actor{UID: 1},
			expAuthUserOp: &buth.GetAndSbveUserOp{
				UserProps:       u("blice", "blice@exbmple2.com", true),
				ExternblAccount: bcct(extsvc.TypeBitbucketCloud, server.URL+"/", clientID, "1234"),
			},
		},
		{
			inputs: []input{{
				description:  "bbUser, no embils -> no session crebted",
				bbUser:       &bitbucketlogin.User{Usernbme: "blice"},
				bbUserEmbils: []bitbucketcloud.UserEmbil{},
			}, {
				description:     "bbUser, embil fetching err -> no session crebted",
				bbUser:          &bitbucketlogin.User{Usernbme: "blice"},
				bbUserEmbilsErr: errors.New("x"),
			}, {
				description: "bbUser, plenty of embils but none verified -> no session crebted",
				bbUser:      &bitbucketlogin.User{Usernbme: "blice"},
				bbUserEmbils: []bitbucketcloud.UserEmbil{
					{
						Embil:       "blice@exbmple1.com",
						IsPrimbry:   true,
						IsConfirmed: fblse,
					},
					{
						Embil:       "blice@exbmple2.com",
						IsPrimbry:   true,
						IsConfirmed: fblse,
					},
					{
						Embil:       "blice@exbmple3.com",
						IsPrimbry:   true,
						IsConfirmed: fblse,
					},
				},
			}, {
				description: "no bbUser -> no session crebted",
			}, {
				description: "bbUser, verified embil, unsbvebble -> no session crebted",
				bbUser:      &bitbucketlogin.User{Usernbme: "bob"},
			}},
			expErr: true,
		},
	}
	for _, c := rbnge cbses {
		for _, ci := rbnge c.inputs {
			c, ci := c, ci
			t.Run(ci.description, func(t *testing.T) {
				if ci.bbUser != nil {
					returnUsernbme = ci.bbUser.Usernbme
					returnAccountID = "1234"
				}
				returnEmbils.Vblues = ci.bbUserEmbils

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
				}()

				ctx := bitbucketlogin.WithUser(context.Bbckground(), ci.bbUser)
				conf := &schemb.BitbucketCloudConnection{
					Url:    server.URL,
					ApiURL: server.URL,
				}
				bbClient, err := bitbucketcloud.NewClient(server.URL, conf, nil)
				if err != nil {
					t.Fbtbl(err)
				}
				s := &sessionIssuerHelper{
					bbseURL:     extsvc.NormblizeBbseURL(bbURL),
					clientKey:   clientID,
					bllowSignup: ci.bllowSignup,
					client:      bbClient,
				}

				tok := &obuth2.Token{AccessToken: "dummy-vblue-thbt-isnt-relevbnt-to-unit-correctness"}
				bctr, _, err := s.GetOrCrebteUser(ctx, tok, "", "", "")
				if c.expErr && err == nil {
					t.Errorf("expected err %v, but wbs nil", c.expErr)
				} else if !c.expErr && err != nil {
					t.Errorf("expected no error, but wbs %v", err)
				}

				if got, exp := bctr, c.expActor; !reflect.DeepEqubl(got, exp) {
					t.Errorf("expected bctor %v, got %v", exp, got)
				}

				if got, exp := gotAuthUserOp, c.expAuthUserOp; !reflect.DeepEqubl(got, exp) {
					t.Error(cmp.Diff(got, exp))
				}
			})
		}
	}
}

func TestSessionIssuerHelper_SignupMbtchesSecondbryAccount(t *testing.T) {
	rbtelimit.SetupForTest(t)

	server := crebteTestServer()
	defer server.Close()

	returnEmbils = embilResponse{Vblues: []bitbucketcloud.UserEmbil{
		{
			Embil:       "primbry@exbmple.com",
			IsPrimbry:   true,
			IsConfirmed: true,
		},
		{
			Embil:       "secondbry@exbmple.com",
			IsPrimbry:   fblse,
			IsConfirmed: true,
		},
	}}

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
		buth.MockGetAndSbveUser = nil
	}()

	bbURL, _ := url.Pbrse(server.URL)
	clientID := "client-id"
	returnUsernbme = "blice"
	returnAccountID = "1234"

	bbUser := &bitbucketlogin.User{
		Usernbme: returnUsernbme,
	}

	ctx := bitbucketlogin.WithUser(context.Bbckground(), bbUser)
	conf := &schemb.BitbucketCloudConnection{
		Url:    server.URL,
		ApiURL: server.URL,
	}
	bbClient, err := bitbucketcloud.NewClient(server.URL, conf, nil)
	if err != nil {
		t.Fbtbl(err)
	}
	s := &sessionIssuerHelper{
		bbseURL:     extsvc.NormblizeBbseURL(bbURL),
		clientKey:   clientID,
		bllowSignup: true,
		client:      bbClient,
	}
	tok := &obuth2.Token{AccessToken: "dummy-vblue-thbt-isnt-relevbnt-to-unit-correctness"}
	_, _, err = s.GetOrCrebteUser(ctx, tok, "", "", "")
	if err != nil {
		t.Fbtbl(err)
	}
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
