package openidconnect

import (
	"net/http"
	"strings"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/session"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SessionKey is the key of the key-value pair in a user session for the OpenID
// Connect authentication provider.
const SessionKey = "oidc@0"

// SessionData is the data format to be stored in a user session.
type SessionData struct {
	ID providers.ConfigID

	// Store only the oauth2.Token fields we need, to avoid hitting the ~4096-byte session data
	// limit.
	AccessToken string
	TokenType   string
}

// SignOut clears OpenID Connect-related data from the session. If possible, it revokes the token
// from the OP. If there is an end-session endpoint, it returns that for the caller to present to
// the user.
func SignOut(w http.ResponseWriter, r *http.Request, sessionKey string, getProvider func(id string) *Provider) (endSessionEndpoint string, err error) {
	defer func() {
		if err != nil {
			_ = session.SetData(w, r, sessionKey, nil) // clear the bad data
		}
	}()

	var data *SessionData
	if err := session.GetData(r, sessionKey, &data); err != nil {
		return "", errors.WithMessage(err, "reading OpenID Connect session data")
	}
	if err := session.SetData(w, r, sessionKey, nil); err != nil {
		return "", errors.WithMessage(err, "clearing OpenID Connect session data")
	}
	if data == nil {
		return "", nil
	}

	p, oidcClient, safeErrMsg, err := GetProviderAndClient(r.Context(), data.ID.ID, GetProvider)
	if err != nil {
		return "", errors.Newf("unable to revoke token or end session for OpenID Connect because failed to get OpenID Connect provider: %s", safeErrMsg)
	}

	endSessionEndpoint = oidcClient.EndSessionEndpoint
	if oidcClient.RevocationEndpoint != "" {
		if err := revokeToken(r.Context(), p, oidcClient.RevocationEndpoint, data.AccessToken, data.TokenType); err != nil {
			return endSessionEndpoint, errors.Wrap(err, "revoking OpenID Connect token")
		}
	}

	// If we are in dotcom mode and this is a SAMS provider, we try to revoke the
	// session on the SAMS instance as well.
	if !dotcom.SourcegraphDotComMode() || !strings.HasPrefix(p.config.ClientID, "sams_cid_") {
		return endSessionEndpoint, nil
	}

	// We only need to do something if the SAMS user session cookie is present.
	sessionCookie, err := r.Cookie("accounts_session_v2")
	if err != nil || sessionCookie.Value == "" {
		return endSessionEndpoint, nil
	}

	// NOTE: It is absolutely true that it is not ideal to have to create a new SAMS
	// client upon every sign-out, but since logic here is a low-frequent and
	// dotcom-specific operation, we can live with it, to avoid cascading
	// refactorings that doesn't really do any useful in enterprise environment.
	connConfig := sams.ConnConfig{
		ExternalURL: p.config.Issuer,
	}
	samsClient, err := sams.NewClientV1(
		sams.ClientV1Config{
			ConnConfig: connConfig,
			TokenSource: sams.ClientCredentialsTokenSource(
				connConfig,
				p.config.ClientID,
				p.config.ClientSecret,
				[]scopes.Scope{
					"sams::session::read",
					"sams::session::write",
				},
			),
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "creating SAMS client")
	}

	samsSession, err := samsClient.Sessions().GetSessionByID(r.Context(), sessionCookie.Value)
	if err != nil {
		return "", errors.Wrap(err, "getting SAMS session")
	}
	if samsSession.User == nil {
		// The session is already invalidated on the SAMS instance.
		return endSessionEndpoint, nil
	}

	err = samsClient.Sessions().SignOutSession(r.Context(), sessionCookie.Value, samsSession.GetUser().GetId())
	if err != nil {
		return "", errors.Wrap(err, "signing out SAMS session")
	}
	return endSessionEndpoint, nil
}
