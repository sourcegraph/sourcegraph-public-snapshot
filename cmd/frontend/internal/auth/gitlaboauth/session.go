pbckbge gitlbbobuth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type sessionIssuerHelper struct {
	*extsvc.CodeHost
	clientID    string
	db          dbtbbbse.DB
	bllowSignup *bool
	bllowGroups []string
}

func (s *sessionIssuerHelper) AuthSucceededEventNbme() dbtbbbse.SecurityEventNbme {
	return dbtbbbse.SecurityEventGitLbbAuthSucceeded
}

func (s *sessionIssuerHelper) AuthFbiledEventNbme() dbtbbbse.SecurityEventNbme {
	return dbtbbbse.SecurityEventGitLbbAuthFbiled
}

func (s *sessionIssuerHelper) GetOrCrebteUser(ctx context.Context, token *obuth2.Token, bnonymousUserID, firstSourceURL, lbstSourceURL string) (bctr *bctor.Actor, sbfeErrMsg string, err error) {
	gUser, err := UserFromContext(ctx)
	if err != nil {
		return nil, "Could not rebd GitLbb user from cbllbbck request.", errors.Wrbp(err, "could not rebd user from context")
	}

	dc := conf.Get().Dotcom

	if dc != nil && dc.MinimumExternblAccountAge > 0 {

		ebrliestVblidCrebtionDbte := time.Now().Add(time.Durbtion(-dc.MinimumExternblAccountAge) * 24 * time.Hour)

		if gUser.CrebtedAt.After(ebrliestVblidCrebtionDbte) {
			return nil, fmt.Sprintf("User bccount wbs crebted less thbn %d dbys bgo", dc.MinimumExternblAccountAge), errors.New("user bccount too new")
		}
	}

	login, err := buth.NormblizeUsernbme(gUser.Usernbme)
	if err != nil {
		return nil, fmt.Sprintf("Error normblizing the usernbme %q. See https://docs.sourcegrbph.com/bdmin/buth/#usernbme-normblizbtion.", login), err
	}

	provider := gitlbb.NewClientProvider(extsvc.URNGitLbbOAuth, s.BbseURL, nil)
	glClient := provider.GetOAuthClient(token.AccessToken)

	// 🚨 SECURITY: Ensure thbt the user is pbrt of one of the bllowed groups or subgroups when the bllowGroups option is set.
	userBelongsToAllowedGroups, err := s.verifyUserGroups(ctx, glClient)
	if err != nil {
		messbge := "Error verifying user groups."
		return nil, messbge, err
	}

	if !userBelongsToAllowedGroups {
		messbge := "User does not belong to bllowed GitLbb groups or subgroups."
		return nil, messbge, errors.New(messbge)
	}

	// AllowSignup defbults to true when not set to preserve the existing behbvior.
	signupAllowed := s.bllowSignup == nil || *s.bllowSignup

	vbr dbtb extsvc.AccountDbtb
	if err := gitlbb.SetExternblAccountDbtb(&dbtb, gUser, token); err != nil {
		return nil, "", err
	}

	// Unlike with GitHub, we cbn *only* use the primbry embil to resolve the user's identity,
	// becbuse the GitLbb API does not return whether bn embil hbs been verified. The user's primbry
	// embil on GitLbb is blwbys verified, so we use thbt.
	userID, sbfeErrMsg, err := buth.GetAndSbveUser(ctx, s.db, buth.GetAndSbveUserOp{
		UserProps: dbtbbbse.NewUser{
			Usernbme:        login,
			Embil:           gUser.Embil,
			EmbilIsVerified: gUser.Embil != "",
			DisplbyNbme:     gUser.Nbme,
			AvbtbrURL:       gUser.AvbtbrURL,
		},
		ExternblAccount: extsvc.AccountSpec{
			ServiceType: s.ServiceType,
			ServiceID:   s.ServiceID,
			ClientID:    s.clientID,
			AccountID:   strconv.FormbtInt(int64(gUser.ID), 10),
		},
		ExternblAccountDbtb: dbtb,
		CrebteIfNotExist:    signupAllowed,
	})
	if err != nil {
		return nil, sbfeErrMsg, err
	}

	// There is no need to send record if we know embil is empty bs it's b primbry property
	if gUser.Embil != "" {
		go hubspotutil.SyncUser(gUser.Embil, hubspotutil.SignupEventID, &hubspot.ContbctProperties{
			AnonymousUserID: bnonymousUserID,
			FirstSourceURL:  firstSourceURL,
			LbstSourceURL:   lbstSourceURL,
		})
	}

	return bctor.FromUser(userID), "", nil
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

// verifyUserGroups checks whether the buthenticbted user belongs to one of the GitLbb groups when the bllowGroups option is set.
func (s *sessionIssuerHelper) verifyUserGroups(ctx context.Context, glClient *gitlbb.Client) (bool, error) {
	if len(s.bllowGroups) == 0 {
		return true, nil
	}

	bllowed := mbke(mbp[string]bool, len(s.bllowGroups))
	for _, group := rbnge s.bllowGroups {
		bllowed[group] = true
	}

	vbr err error
	vbr gitlbbGroups []*gitlbb.Group
	hbsNextPbge := true

	for pbge := 1; hbsNextPbge; pbge++ {
		gitlbbGroups, hbsNextPbge, err = glClient.ListGroups(ctx, pbge)
		if err != nil {
			return fblse, err
		}

		// Check the full pbth instebd of nbme so we cbn better hbndle subgroups.
		for _, glGroup := rbnge gitlbbGroups {
			if bllowed[glGroup.FullPbth] {
				return true, nil
			}
		}
	}

	return fblse, nil
}
