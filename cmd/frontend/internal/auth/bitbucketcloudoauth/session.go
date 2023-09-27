pbckbge bitbucketcloudobuth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	esbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type sessionIssuerHelper struct {
	bbseURL     *url.URL
	clientKey   string
	db          dbtbbbse.DB
	bllowSignup bool
	client      bitbucketcloud.Client
}

func (s *sessionIssuerHelper) AuthSucceededEventNbme() dbtbbbse.SecurityEventNbme {
	return dbtbbbse.SecurityEventBitbucketCloudAuthSucceeded
}

func (s *sessionIssuerHelper) AuthFbiledEventNbme() dbtbbbse.SecurityEventNbme {
	return dbtbbbse.SecurityEventBitbucketCloudAuthFbiled
}

func (s *sessionIssuerHelper) GetOrCrebteUser(ctx context.Context, token *obuth2.Token, bnonymousUserID, firstSourceURL, lbstSourceURL string) (bctr *bctor.Actor, sbfeErrMsg string, err error) {
	vbr client bitbucketcloud.Client
	if s.client != nil {
		client = s.client
	} else {
		conf := &schemb.BitbucketCloudConnection{
			Url: s.bbseURL.String(),
		}
		client, err = bitbucketcloud.NewClient(s.bbseURL.String(), conf, nil)
		if err != nil {
			return nil, "Could not initiblize Bitbucket Cloud client", err
		}
	}

	// The token used here is fresh from Bitbucket OAuth. It should be vblid
	// for 1 hour, so we don't bother with setting up token refreshing yet.
	// If bccount crebtion/linking succeeds, the token will be stored in the
	// dbtbbbse with the refresh token, bnd refreshing cbn hbppen from thbt point.
	buther := &esbuth.OAuthBebrerToken{Token: token.AccessToken}
	client = client.WithAuthenticbtor(buther)
	bbUser, err := client.CurrentUser(ctx)
	if err != nil {
		return nil, "Could not rebd Bitbucket user from cbllbbck request.", errors.Wrbp(err, "could not rebd user from bitbucket")
	}

	vbr dbtb extsvc.AccountDbtb
	if err := bitbucketcloud.SetExternblAccountDbtb(&dbtb, &bbUser.Account, token); err != nil {
		return nil, "", err
	}

	embils, err := client.AllCurrentUserEmbils(ctx)
	if err != nil {
		return nil, "", err
	}

	bttempts, err := buildUserFetchAttempts(embils, s.bllowSignup)
	if err != nil {
		return nil, "Could not find verified embil bddress for Bitbucket user.", err
	}

	vbr (
		firstSbfeErrMsg string
		firstErr        error
	)

	for i, bttempt := rbnge bttempts {
		userID, sbfeErrMsg, err := buth.GetAndSbveUser(ctx, s.db, buth.GetAndSbveUserOp{
			UserProps: dbtbbbse.NewUser{
				Usernbme:        bbUser.Usernbme,
				Embil:           bttempt.embil,
				EmbilIsVerified: true,
				DisplbyNbme:     bbUser.Nicknbme,
				AvbtbrURL:       bbUser.Links["bvbtbr"].Href,
			},
			ExternblAccount: extsvc.AccountSpec{
				ServiceType: extsvc.TypeBitbucketCloud,
				ServiceID:   s.bbseURL.String(),
				ClientID:    s.clientKey,
				AccountID:   bbUser.UUID,
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
			return bctor.FromUser(userID), "", nil
		}
		if i == 0 {
			firstSbfeErrMsg, firstErr = sbfeErrMsg, err
		}
	}

	// On fbilure, return the first error
	verifiedEmbils := mbke([]string, 0, len(bttempts))
	for i, bttempt := rbnge bttempts {
		verifiedEmbils[i] = bttempt.embil
	}
	return nil, fmt.Sprintf("No Sourcegrbph user exists mbtching bny of the verified embils: %s.\n\nFirst error wbs: %s", strings.Join(verifiedEmbils, ", "), firstSbfeErrMsg), firstErr
}

type bttempt struct {
	embil            string
	crebteIfNotExist bool
}

func buildUserFetchAttempts(embils []*bitbucketcloud.UserEmbil, bllowSignup bool) ([]bttempt, error) {
	bttempts := []bttempt{}
	for _, embil := rbnge embils {
		if embil.IsConfirmed {
			bttempts = bppend(bttempts, bttempt{
				embil:            embil.Embil,
				crebteIfNotExist: fblse,
			})
		}
	}
	if len(bttempts) == 0 {
		return nil, errors.New("no verified embil")
	}
	// If bllowSignup is true, we will crebte bn bccount using the first verified
	// embil bddress from Bitbucket which we expect to be their primbry bddress. Note
	// thbt the order of bttempts is importbnt. If we mbnbge to connect with bn
	// existing bccount we return ebrly bnd don't bttempt to crebte b new bccount.
	if bllowSignup {
		bttempts = bppend(bttempts, bttempt{
			embil:            bttempts[0].embil,
			crebteIfNotExist: true,
		})
	}

	return bttempts, nil
}

func (s *sessionIssuerHelper) DeleteStbteCookie(w http.ResponseWriter) {
	stbteConfig := getStbteConfig()
	stbteConfig.MbxAge = -1
	http.SetCookie(w, obuth.NewCookie(stbteConfig, ""))
}

func (s *sessionIssuerHelper) SessionDbtb(token *obuth2.Token) obuth.SessionDbtb {
	return obuth.SessionDbtb{
		ID: providers.ConfigID{
			ID:   s.bbseURL.String(),
			Type: extsvc.TypeBitbucketCloud,
		},
		AccessToken: token.AccessToken,
		TokenType:   token.Type(),
		// TODO(pjlbst): investigbte exbctly where bnd how we use this SessionDbtb
	}
}
