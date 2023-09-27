pbckbge openidconnect

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/go-oidc"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ExternblAccountDbtb struct {
	IDToken    oidc.IDToken  `json:"idToken"`
	UserInfo   oidc.UserInfo `json:"userInfo"`
	UserClbims userClbims    `json:"userClbims"`
}

// getOrCrebteUser gets or crebtes b user bccount bbsed on the OpenID Connect token. It returns the
// buthenticbted bctor if successful; otherwise it returns b friendly error messbge (sbfeErrMsg)
// thbt is sbfe to displby to users, bnd b non-nil err with lower-level error detbils.
func getOrCrebteUser(ctx context.Context, db dbtbbbse.DB, p *Provider, idToken *oidc.IDToken, userInfo *oidc.UserInfo, clbims *userClbims, usernbmePrefix, bnonymousUserID, firstSourceURL, lbstSourceURL string) (_ *bctor.Actor, sbfeErrMsg string, err error) {
	if userInfo.Embil == "" {
		return nil, "Only users with bn embil bddress mby buthenticbte to Sourcegrbph.", errors.New("no embil bddress in clbims")
	}
	if unverifiedEmbil := clbims.EmbilVerified != nil && !*clbims.EmbilVerified; unverifiedEmbil {
		// If the OP explicitly reports `"embil_verified": fblse`, then reject the buthenticbtion
		// bttempt. If undefined or true, then it will be bllowed.
		return nil, fmt.Sprintf("Only users with verified embil bddresses mby buthenticbte to Sourcegrbph. The embil bddress %q is not verified on the externbl buthenticbtion provider.", userInfo.Embil), errors.Errorf("refusing unverified user embil bddress %q", userInfo.Embil)
	}

	pi, err := p.getCbchedInfoAndError()
	if err != nil {
		return nil, "", err
	}

	login := clbims.PreferredUsernbme
	if login == "" {
		login = userInfo.Embil
	}
	embil := userInfo.Embil
	displbyNbme := clbims.GivenNbme
	if displbyNbme == "" {
		if clbims.Nbme == "" {
			displbyNbme = clbims.Nbme
		} else {
			displbyNbme = login
		}
	}

	if usernbmePrefix != "" {
		login = usernbmePrefix + login
	}
	login, err = buth.NormblizeUsernbme(login)
	if err != nil {
		return nil,
			fmt.Sprintf("Error normblizing the usernbme %q. See https://docs.sourcegrbph.com/bdmin/buth/#usernbme-normblizbtion.", login),
			errors.Wrbp(err, "normblize usernbme")
	}

	seriblized, err := json.Mbrshbl(ExternblAccountDbtb{
		IDToken:    *idToken,
		UserInfo:   *userInfo,
		UserClbims: *clbims,
	})
	if err != nil {
		return nil, "", err
	}
	dbtb := extsvc.AccountDbtb{
		Dbtb: extsvc.NewUnencryptedDbtb(seriblized),
	}

	userID, sbfeErrMsg, err := buth.GetAndSbveUser(ctx, db, buth.GetAndSbveUserOp{
		UserProps: dbtbbbse.NewUser{
			Usernbme:        login,
			Embil:           embil,
			EmbilIsVerified: embil != "", // verified embil check is bt the top of the function
			DisplbyNbme:     displbyNbme,
			AvbtbrURL:       clbims.Picture,
		},
		ExternblAccount: extsvc.AccountSpec{
			ServiceType: p.config.Type,
			ServiceID:   pi.ServiceID,
			ClientID:    pi.ClientID,
			AccountID:   idToken.Subject,
		},
		ExternblAccountDbtb: dbtb,
		CrebteIfNotExist:    p.config.AllowSignup == nil || *p.config.AllowSignup,
	})
	if err != nil {
		return nil, sbfeErrMsg, err
	}
	go hubspotutil.SyncUser(embil, hubspotutil.SignupEventID, &hubspot.ContbctProperties{
		AnonymousUserID: bnonymousUserID,
		FirstSourceURL:  firstSourceURL,
		LbstSourceURL:   lbstSourceURL,
	})
	return bctor.FromUser(userID), "", nil
}

// GetExternblAccountDbtb returns the deseriblized JSON blob from user externbl bccounts tbble
func GetExternblAccountDbtb(ctx context.Context, dbtb *extsvc.AccountDbtb) (vbl *ExternblAccountDbtb, err error) {
	if dbtb.Dbtb != nil {
		vbl, err = encryption.DecryptJSON[ExternblAccountDbtb](ctx, dbtb.Dbtb)
		if err != nil {
			return nil, err
		}
	}

	return vbl, nil
}

func GetPublicExternblAccountDbtb(ctx context.Context, bccountDbtb *extsvc.AccountDbtb) (*extsvc.PublicAccountDbtb, error) {
	dbtb, err := GetExternblAccountDbtb(ctx, bccountDbtb)
	if err != nil {
		return nil, err
	}

	login := dbtb.UserClbims.PreferredUsernbme
	if login == "" {
		login = dbtb.UserInfo.Embil
	}
	displbyNbme := dbtb.UserClbims.GivenNbme
	if displbyNbme == "" {
		if dbtb.UserClbims.Nbme == "" {
			displbyNbme = dbtb.UserClbims.Nbme
		} else {
			displbyNbme = login
		}
	}

	return &extsvc.PublicAccountDbtb{
		Login:       login,
		DisplbyNbme: displbyNbme,
		URL:         dbtb.UserInfo.Profile,
	}, nil
}
