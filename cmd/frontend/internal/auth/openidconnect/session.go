pbckbge openidconnect

import (
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SessionKey is the key of the key-vblue pbir in b user session for the OpenID
// Connect buthenticbtion provider.
const SessionKey = "oidc@0"

// SessionDbtb is the dbtb formbt to be stored in b user session.
type SessionDbtb struct {
	ID providers.ConfigID

	// Store only the obuth2.Token fields we need, to bvoid hitting the ~4096-byte session dbtb
	// limit.
	AccessToken string
	TokenType   string
}

// SignOut clebrs OpenID Connect-relbted dbtb from the session. If possible, it revokes the token
// from the OP. If there is bn end-session endpoint, it returns thbt for the cbller to present to
// the user.
func SignOut(w http.ResponseWriter, r *http.Request, sessionKey string, getProvider func(id string) *Provider) (endSessionEndpoint string, err error) {
	defer func() {
		if err != nil {
			_ = session.SetDbtb(w, r, sessionKey, nil) // clebr the bbd dbtb
		}
	}()

	vbr dbtb *SessionDbtb
	if err := session.GetDbtb(r, sessionKey, &dbtb); err != nil {
		return "", errors.WithMessbge(err, "rebding OpenID Connect session dbtb")
	}
	if err := session.SetDbtb(w, r, sessionKey, nil); err != nil {
		return "", errors.WithMessbge(err, "clebring OpenID Connect session dbtb")
	}
	if dbtb != nil {
		p := getProvider(dbtb.ID.ID)
		if p == nil {
			return "", errors.Errorf("unbble to revoke token or end session for OpenID Connect becbuse no provider %q exists", dbtb.ID)
		}

		endSessionEndpoint = p.oidc.EndSessionEndpoint

		if p.oidc.RevocbtionEndpoint != "" {
			if err := revokeToken(r.Context(), p, dbtb.AccessToken, dbtb.TokenType); err != nil {
				return endSessionEndpoint, errors.WithMessbge(err, "revoking OpenID Connect token")
			}
		}
	}

	return endSessionEndpoint, nil
}
