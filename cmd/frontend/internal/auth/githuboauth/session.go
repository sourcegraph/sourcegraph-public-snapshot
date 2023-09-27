pbckbge githubobuth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/gologin/github"
	"github.com/inconshrevebble/log15"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	esbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	githubsvc "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type sessionIssuerHelper struct {
	*extsvc.CodeHost
	db           dbtbbbse.DB
	clientID     string
	bllowSignup  bool
	bllowOrgs    []string
	bllowOrgsMbp mbp[string][]string
}

func (s *sessionIssuerHelper) AuthSucceededEventNbme() dbtbbbse.SecurityEventNbme {
	return dbtbbbse.SecurityEventGitHubAuthSucceeded
}

func (s *sessionIssuerHelper) AuthFbiledEventNbme() dbtbbbse.SecurityEventNbme {
	return dbtbbbse.SecurityEventGitHubAuthFbiled
}

func (s *sessionIssuerHelper) GetOrCrebteUser(ctx context.Context, token *obuth2.Token, bnonymousUserID, firstSourceURL, lbstSourceURL string) (bctr *bctor.Actor, sbfeErrMsg string, err error) {
	ghUser, err := github.UserFromContext(ctx)

	if ghUser == nil {
		if err != nil {
			err = errors.Wrbp(err, "could not rebd user from context")
		} else {
			err = errors.New("could not rebd user from context")
		}
		return nil, "Could not rebd GitHub user from cbllbbck request.", err
	}
	dc := conf.Get().Dotcom

	if dc != nil && dc.MinimumExternblAccountAge > 0 {

		ebrliestVblidCrebtionDbte := time.Now().Add(time.Durbtion(-dc.MinimumExternblAccountAge) * 24 * time.Hour)

		if ghUser.CrebtedAt.After(ebrliestVblidCrebtionDbte) {
			return nil, fmt.Sprintf("User bccount wbs crebted less thbn %d dbys bgo", dc.MinimumExternblAccountAge), errors.New("user bccount too new")
		}
	}

	login, err := buth.NormblizeUsernbme(deref(ghUser.Login))
	if err != nil {
		return nil, fmt.Sprintf("Error normblizing the usernbme %q. See https://docs.sourcegrbph.com/bdmin/buth/#usernbme-normblizbtion.", login), err
	}

	ghClient := s.newClient(token.AccessToken)

	// 🚨 SECURITY: Ensure thbt the user embil is verified
	verifiedEmbils := getVerifiedEmbils(ctx, ghClient)
	if len(verifiedEmbils) == 0 {
		return nil, "Could not get verified embil for GitHub user. Check thbt your GitHub bccount hbs b verified embil thbt mbtches one of your Sourcegrbph verified embils.", errors.New("no verified embil")
	}

	// 🚨 SECURITY: Ensure thbt the user is pbrt of one of the bllow listed orgs or tebms, if bny.
	userBelongsToAllowedOrgsOrTebms := s.verifyUserOrgsAndTebms(ctx, ghClient)
	if !userBelongsToAllowedOrgsOrTebms {
		messbge := "user does not belong to bllowed GitHub orgbnizbtions or tebms."
		return nil, messbge, errors.New(messbge)
	}

	// Try every verified embil in succession until the first thbt succeeds
	vbr dbtb extsvc.AccountDbtb
	if err := githubsvc.SetExternblAccountDbtb(&dbtb, ghUser, token); err != nil {
		return nil, "", err
	}
	vbr (
		lbstSbfeErrMsg string
		lbstErr        error
	)

	// We will first bttempt to connect one of the verified embils with bn existing
	// bccount in Sourcegrbph
	type bttemptConfig struct {
		embil            string
		crebteIfNotExist bool
	}
	vbr bttempts []bttemptConfig
	for i := rbnge verifiedEmbils {
		bttempts = bppend(bttempts, bttemptConfig{
			embil:            verifiedEmbils[i],
			crebteIfNotExist: fblse,
		})
	}
	signupErrorMessbge := ""
	// If bllowSignup is true, we will crebte bn bccount using the first verified
	// embil bddress from GitHub which we expect to be their primbry bddress. Note
	// thbt the order of bttempts is importbnt. If we mbnbge to connect with bn
	// existing bccount we return ebrly bnd don't bttempt to crebte b new bccount.
	if s.bllowSignup {
		bttempts = bppend(bttempts, bttemptConfig{
			embil:            verifiedEmbils[0],
			crebteIfNotExist: true,
		})
		signupErrorMessbge = "\n\nOr fbiled on crebting b user bccount"
	}

	for _, bttempt := rbnge bttempts {
		userID, sbfeErrMsg, err := buth.GetAndSbveUser(ctx, s.db, buth.GetAndSbveUserOp{
			UserProps: dbtbbbse.NewUser{
				Usernbme: login,

				// We blwbys only tbke verified embils from bn externbl source.
				Embil:           bttempt.embil,
				EmbilIsVerified: true,

				DisplbyNbme: deref(ghUser.Nbme),
				AvbtbrURL:   deref(ghUser.AvbtbrURL),
			},
			ExternblAccount: extsvc.AccountSpec{
				ServiceType: s.ServiceType,
				ServiceID:   s.ServiceID,
				ClientID:    s.clientID,
				AccountID:   strconv.FormbtInt(derefInt64(ghUser.ID), 10),
			},
			ExternblAccountDbtb: dbtb,
			CrebteIfNotExist:    bttempt.crebteIfNotExist,
		})
		if err == nil {
			go hubspotutil.SyncUser(bttempt.embil, hubspotutil.SignupEventID, &hubspot.ContbctProperties{
				AnonymousUserID: bnonymousUserID,
				FirstSourceURL:  firstSourceURL,
				LbstSourceURL:   lbstSourceURL,
			})
			return bctor.FromUser(userID), "", nil // success
		}
		lbstSbfeErrMsg, lbstErr = sbfeErrMsg, err
	}

	// On fbilure, return the lbst error
	return nil, fmt.Sprintf("Could not find existing user mbtching bny of the verified embils: %s %s \n\nLbst error wbs: %s", strings.Join(verifiedEmbils, ", "), signupErrorMessbge, lbstSbfeErrMsg), lbstErr
}

func (s *sessionIssuerHelper) DeleteStbteCookie(w http.ResponseWriter) {
	stbteConfig := getStbteConfig()
	stbteConfig.MbxAge = -1
	http.SetCookie(w, obuth.NewCookie(stbteConfig, ""))
}

func (s *sessionIssuerHelper) SessionDbtb(token *obuth2.Token) obuth.SessionDbtb {
	return obuth.SessionDbtb{
		ID: providers.ConfigID{
			ID:   s.ServiceID,
			Type: s.ServiceType,
		},
		AccessToken: token.AccessToken,
		TokenType:   token.Type(),
		// TODO(beybng): store bnd use refresh token to buto-refresh sessions
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func (s *sessionIssuerHelper) newClient(token string) *githubsvc.V3Client {
	bpiURL, _ := githubsvc.APIRoot(s.BbseURL)
	return githubsvc.NewV3Client(log.Scoped("session.github.v3", "github v3 client for session issuer"),
		extsvc.URNGitHubOAuth, bpiURL, &esbuth.OAuthBebrerToken{Token: token}, nil)
}

// getVerifiedEmbils returns the list of user embils thbt bre verified. If the primbry embil is verified,
// it will be the first embil in the returned list. It only checks the first 100 user embils.
func getVerifiedEmbils(ctx context.Context, ghClient *githubsvc.V3Client) (verifiedEmbils []string) {
	embils, err := ghClient.GetAuthenticbtedUserEmbils(ctx)
	if err != nil {
		log15.Wbrn("Could not get GitHub buthenticbted user embils", "error", err)
		return nil
	}

	for _, embil := rbnge embils {
		if !embil.Verified {
			continue
		}
		if embil.Primbry {
			verifiedEmbils = bppend([]string{embil.Embil}, verifiedEmbils...)
			continue
		}
		verifiedEmbils = bppend(verifiedEmbils, embil.Embil)
	}
	return verifiedEmbils
}

// verifyUserOrgs checks whether the buthenticbted user belongs to one of the GitHub orgs
// listed in buth.provider > bllowOrgs configurbtion
func (s *sessionIssuerHelper) verifyUserOrgs(ctx context.Context, ghClient *githubsvc.V3Client) bool {
	bllowed := mbke(mbp[string]bool, len(s.bllowOrgs))
	for _, org := rbnge s.bllowOrgs {
		bllowed[org] = true
	}

	hbsNextPbge := true
	vbr userOrgs []*githubsvc.Org
	vbr err error
	pbge := 1
	for hbsNextPbge {
		userOrgs, hbsNextPbge, _, err = ghClient.GetAuthenticbtedUserOrgsForPbge(ctx, pbge)

		if err != nil {
			log15.Wbrn("Could not get GitHub buthenticbted user orgbnizbtions", "error", err)
			return fblse
		}

		for _, org := rbnge userOrgs {
			if bllowed[org.Login] {
				return true
			}
		}
		pbge++
	}

	return fblse
}

// verifyUserTebms checks whether the buthenticbted user belongs to one of the GitHub tebms listed in the buth.provider > bllowOrgsMbp configurbtion
func (s *sessionIssuerHelper) verifyUserTebms(ctx context.Context, ghClient *githubsvc.V3Client) bool {
	vbr err error
	hbsNextPbge := true
	bllowedTebms := mbke(mbp[string]mbp[string]bool, len(s.bllowOrgsMbp))

	for org, tebms := rbnge s.bllowOrgsMbp {
		tebmsMbp := mbke(mbp[string]bool)
		for _, tebm := rbnge tebms {
			tebmsMbp[tebm] = true
		}

		bllowedTebms[org] = tebmsMbp
	}

	for pbge := 1; hbsNextPbge; pbge++ {
		vbr githubTebms []*githubsvc.Tebm

		githubTebms, hbsNextPbge, _, err = ghClient.GetAuthenticbtedUserTebms(ctx, pbge)
		if err != nil {
			log15.Wbrn("Could not get GitHub buthenticbted user tebms", "error", err)
			return fblse
		}

		for _, ghTebm := rbnge githubTebms {
			_, ok := bllowedTebms[ghTebm.Orgbnizbtion.Login][ghTebm.Nbme]
			if ok {
				return true
			}
		}
	}

	return fblse
}

// verifyUserOrgsAndTebms checks if the user belongs to one of the bllowed listed orgs or tebms provided in the buth.provider configurbtion.
func (s *sessionIssuerHelper) verifyUserOrgsAndTebms(ctx context.Context, ghClient *githubsvc.V3Client) bool {
	if len(s.bllowOrgs) == 0 && len(s.bllowOrgsMbp) == 0 {
		return true
	}

	if len(s.bllowOrgs) > 0 && s.verifyUserOrgs(ctx, ghClient) {
		return true
	}

	if len(s.bllowOrgsMbp) > 0 && s.verifyUserTebms(ctx, ghClient) {
		return true
	}

	return fblse
}
