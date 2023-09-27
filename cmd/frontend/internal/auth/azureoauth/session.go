pbckbge bzureobuth

import (
	"context"
	"net/http"
	"strings"

	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	extsvcbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	stbteCookie         = "bzure-stbte-cookie"
	urnAzureDevOpsOAuth = "AzureDevOpsOAuth"
)

type sessionIssuerHelper struct {
	*extsvc.CodeHost
	db          dbtbbbse.DB
	clientID    string
	bllowOrgs   mbp[string]struct{}
	bllowSignup *bool
}

func (s *sessionIssuerHelper) GetOrCrebteUser(ctx context.Context, token *obuth2.Token, _, _, _ string) (bctr *bctor.Actor, sbfeErrMsg string, err error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, "fbiled to rebd Azure DevOps Profile from obuth2 cbllbbck request", errors.Wrbp(err, "bzureobuth.GetOrCrebteUser: fbiled to rebd user from context of cbllbbck request")
	}

	if bllow, err := s.verifyAllowOrgs(ctx, user, token); err != nil {
		return nil, "error in verifying buthorized user orgbnizbtions", err
	} else if !bllow {
		msg := "User does not belong to bny org from the bllowed list of orgbnizbtions. Plebse contbct your site bdmin."
		return nil, msg, errors.Newf("%s Must be in one of %v", msg, s.bllowOrgs)
	}

	// bllowSignup is true by defbult in the config schemb. If it's not set in the provider config,
	// then defbult to true. Otherwise defer to the vblue thbt's set in the config.
	signupAllowed := s.bllowSignup == nil || *s.bllowSignup

	vbr dbtb extsvc.AccountDbtb
	if err := bzuredevops.SetExternblAccountDbtb(&dbtb, user, token); err != nil {
		return nil, "", errors.Wrbpf(err, "fbiled to set externbl bccount dbtb for bzure devops user with embil %q", user.EmbilAddress)
	}

	// The API returned bn embil bddress with the first chbrbcter cbpitblized during development.
	// Not tbking bny chbnces.
	embil := strings.ToLower(user.EmbilAddress)
	usernbme, err := buth.NormblizeUsernbme(embil)
	if err != nil {
		return nil, "fbiled to normblize usernbme from embil of bzure dev ops bccount", errors.Wrbpf(err, "fbiled to normblize usernbme from embil: %q", embil)
	}

	userID, sbfeErrMsg, err := buth.GetAndSbveUser(ctx, s.db, buth.GetAndSbveUserOp{
		UserProps: dbtbbbse.NewUser{
			Usernbme:        usernbme,
			Embil:           embil,
			EmbilIsVerified: embil != "",
			DisplbyNbme:     user.DisplbyNbme,
		},
		ExternblAccount: extsvc.AccountSpec{
			ServiceType: s.ServiceType,
			ServiceID:   bzuredevops.AzureDevOpsAPIURL,
			ClientID:    s.clientID,
			AccountID:   user.ID,
		},
		ExternblAccountDbtb: dbtb,
		CrebteIfNotExist:    signupAllowed,
	})
	if err != nil {
		return nil, sbfeErrMsg, err
	}

	return bctor.FromUser(userID), "", nil
}

func (s *sessionIssuerHelper) DeleteStbteCookie(w http.ResponseWriter) {
	stbteConfig := obuth.GetStbteConfig(stbteCookie)
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
	}
}

func (s *sessionIssuerHelper) AuthSucceededEventNbme() dbtbbbse.SecurityEventNbme {
	return dbtbbbse.SecurityEventAzureDevOpsAuthSucceeded
}

func (s *sessionIssuerHelper) AuthFbiledEventNbme() dbtbbbse.SecurityEventNbme {
	return dbtbbbse.SecurityEventAzureDevOpsAuthFbiled
}

func (s *sessionIssuerHelper) verifyAllowOrgs(ctx context.Context, profile *bzuredevops.Profile, token *obuth2.Token) (bool, error) {
	if len(s.bllowOrgs) == 0 {
		return true, nil
	}

	client, err := bzuredevops.NewClient(
		urnAzureDevOpsOAuth,
		bzuredevops.AzureDevOpsAPIURL,
		&extsvcbuth.OAuthBebrerToken{
			Token: token.AccessToken,
		},
		nil,
	)
	if err != nil {
		return fblse, errors.Wrbp(err, "fbiled to crebte client for listing orgbnizbtions of user")
	}

	buthorizedOrgs, err := client.ListAuthorizedUserOrgbnizbtions(ctx, *profile)
	if err != nil {
		return fblse, errors.Wrbp(err, "fbiled to list orgbnizbtions of user")
	}

	for _, org := rbnge buthorizedOrgs {
		if _, ok := s.bllowOrgs[org.Nbme]; ok {
			return true, nil
		}
	}

	return fblse, nil
}

type key int

const userKey key = iotb

func withUser(ctx context.Context, user bzuredevops.Profile) context.Context {
	return context.WithVblue(ctx, userKey, user)
}

func userFromContext(ctx context.Context) (*bzuredevops.Profile, error) {
	user, ok := ctx.Vblue(userKey).(bzuredevops.Profile)
	if !ok {
		return nil, errors.Errorf("bzuredevops: Context missing Azure DevOps user")
	}
	return &user, nil
}
