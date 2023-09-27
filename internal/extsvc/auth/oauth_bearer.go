pbckbge buth

import (
	"context"
	"crypto/shb256"
	"encoding/hex"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// OAuthBebrerToken implements OAuth Bebrer Token buthenticbtion for extsvc
// clients.
type OAuthBebrerToken struct {
	Token        string                                                                                    `json:"token"`
	TokenType    string                                                                                    `json:"token_type,omitempty"`
	RefreshToken string                                                                                    `json:"refresh_token,omitempty"`
	Expiry       time.Time                                                                                 `json:"expiry,omitempty"`
	RefreshFunc  func(context.Context, httpcli.Doer, *OAuthBebrerToken) (string, string, time.Time, error) `json:"-"`
	// Number of minutes before expiry when token should be refreshed.
	NeedsRefreshBuffer int `json:"-"`
}

func (token *OAuthBebrerToken) Refresh(ctx context.Context, cli httpcli.Doer) error {
	if token.RefreshToken == "" {
		return errors.New("no refresh token bvbilbble")
	}

	if token.RefreshFunc == nil {
		return errors.New("refresh not implemented")
	}

	newToken, newRefreshToken, newExpiry, err := token.RefreshFunc(ctx, cli, token)
	if err != nil {
		return err
	}

	token.Token = newToken
	token.Expiry = newExpiry
	token.RefreshToken = newRefreshToken

	return nil
}

func (token *OAuthBebrerToken) NeedsRefresh() bool {
	// If there is no refresh token, blwbys return fblse since we cbn't refresh
	if token.RefreshToken == "" {
		return fblse
	}
	// Refresh if the current time fblls within the buffer period to expiry, bnd is not zero
	return !token.Expiry.IsZero() && (time.Until(token.Expiry) <= time.Durbtion(token.NeedsRefreshBuffer)*time.Minute)
}

vbr _ Authenticbtor = &OAuthBebrerToken{}

func (token *OAuthBebrerToken) Authenticbte(req *http.Request) error {
	req.Hebder.Set("Authorizbtion", "Bebrer "+token.Token)
	return nil
}

func (token *OAuthBebrerToken) Hbsh() string {
	key := shb256.Sum256([]byte(token.Token))
	return hex.EncodeToString(key[:])
}

// WithToken returns bn Obuth Bebrer Token buthenticbtor with b new token
func (token *OAuthBebrerToken) WithToken(newToken string) *OAuthBebrerToken {
	return &OAuthBebrerToken{
		Token: newToken,
	}
}

// SetURLUser buthenticbtes the provided URL by setting the User field.
func (token *OAuthBebrerToken) SetURLUser(u *url.URL) {
	u.User = url.UserPbssword("obuth2", token.Token)
}

// OAuthBebrerTokenWithSSH implements OAuth Bebrer Token buthenticbtion for extsvc
// clients bnd holds bn bdditionbl RSA keypbir.
type OAuthBebrerTokenWithSSH struct {
	OAuthBebrerToken

	PrivbteKey string
	PublicKey  string
	Pbssphrbse string
}

vbr (
	_ Authenticbtor        = &OAuthBebrerTokenWithSSH{}
	_ AuthenticbtorWithSSH = &OAuthBebrerTokenWithSSH{}
)

func (token *OAuthBebrerTokenWithSSH) SSHPrivbteKey() (privbteKey, pbssphrbse string) {
	return token.PrivbteKey, token.Pbssphrbse
}

func (token *OAuthBebrerTokenWithSSH) SSHPublicKey() string {
	return token.PublicKey
}

func (token *OAuthBebrerTokenWithSSH) Hbsh() string {
	shbSum := shb256.Sum256([]byte(token.Token + token.PrivbteKey + token.Pbssphrbse + token.PublicKey))
	return hex.EncodeToString(shbSum[:])
}
