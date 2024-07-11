package ssc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/oauthtoken"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// APIProxyHandler is an HTTP handler that essentially proxies API requests from the
// current Sourcegraph instance to the SSC backend, but exchanging the credentials
// of the calling Sourcegraph user with an access token for their SAMS identity.
//
// This way we can transparently serve an HTTP API from Sourcegraph.com, when in
// actuality it is processed by a different service all together. (And thereby
// allowing us to decouple Cody Pro-specifc functionality from the Sourcegraph
// Enterprise instance.)
//
// The API Proxy simply handles the credential exchange and verification. But
// will send along the HTTP request. (Even if the URL and method aren't supported,
// and would serve a 404/405 response, etc.)
type APIProxyHandler struct {
	CodyProConfig *schema.CodyProConfig
	DB            database.DB
	Logger        log.Logger

	// URLPrefix of where the handler is served. e.g. ".api/ssc/proxy/". This
	// will be replaced with the SSC-specific URL prefix ()"cody/api/v1/").
	URLPrefix string

	// SAMSOAuthContext is the metadata necessary for contacting SAMS. Used
	// when we notice a Sourcegraph account's SAMS identity has an expired
	// access token.
	SAMSOAuthContext *oauthutil.OAuthContext
}

var _ http.Handler = (*APIProxyHandler)(nil)

// GetSAMSOAuthContext returns the OAuthContext object to describe the SAMS
// IdP registered to the current Sourcegraph instance. (As identified by
// `GETSAMSServiceID()`)
func GetSAMSOAuthContext() (*oauthutil.OAuthContext, error) {
	for _, provider := range conf.Get().AuthProviders {
		oidcInfo := provider.Openidconnect
		if oidcInfo == nil {
			continue
		}
		if oidcInfo.Issuer == GetSAMSServiceID() {
			oauthCtx := oauthutil.OAuthContext{
				ClientID:     oidcInfo.ClientID,
				ClientSecret: oidcInfo.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  fmt.Sprintf("%s/oauth/authorize", oidcInfo.Issuer),
					TokenURL: fmt.Sprintf("%s/oauth/token", oidcInfo.Issuer),
				},
			}
			return &oauthCtx, nil
		}
	}

	return nil, errors.New("no SAMS configuration found")
}

// getUserIDFromRequest extracts the Sourcegraph User ID from the incoming request,
// or returns an error suitable for sending to the end user.
func (p *APIProxyHandler) getUserIDFromContext(ctx context.Context) (int32, error) {
	callingActor := actor.FromContext(ctx)
	if callingActor == nil || !callingActor.IsAuthenticated() {
		p.Logger.Warn("rejecting request made by unauthenticated Sourcegraph user")
		return 0, errors.New("no credentials available")
	}
	if callingActor.IsInternal() || callingActor.SourcegraphOperator {
		p.Logger.Warn("rejecting request made by internal service / Sourcegraph Operator")
		return 0, errors.New("request not made on behalf of a user")
	}
	return callingActor.UID, nil
}

// buildProxyRequest converts the incoming HTTP request into what will be sent to the SSC backend.
func (p *APIProxyHandler) buildProxyRequest(sourceReq *http.Request, token string) (*http.Request, error) {
	// For simplicity, read the full request body before sending the proxy request.
	var bodyReader io.Reader
	if sourceReq.Body != nil {
		bodyBytes, err := io.ReadAll(sourceReq.Body)
		if err != nil {
			return nil, errors.Wrap(err, "reading source request body")
		}
		bodyReader = strings.NewReader(string(bodyBytes))
	}

	// Construct the remapped URL on the SSC backend.
	// For example:
	// Source: ".api/ssc/proxy/" + "teams/current/members"
	// Proxy :    "cody/api/v1/" + "teams/current/members"
	sourceURLPath := strings.TrimPrefix(sourceReq.URL.Path, p.URLPrefix)
	sourceURLPath = "/" + sourceURLPath // Force the path to be rooted.
	// nosemgrep: resolving the supplied path, to concatenate with the SSC API URL prefix below.
	sourceURLPath = path.Clean(sourceURLPath)

	sscURLPath, err := url.JoinPath(
		p.CodyProConfig.SscBackendOrigin,
		"cody/api/v1",
		sourceURLPath)
	if err != nil {
		return nil, errors.Wrap(err, "building proxy URL")
	}

	sscURL, err := url.Parse(sscURLPath)
	if err != nil {
		return nil, errors.Wrap(err, "parsing generated proxy URL")
	}
	sscURL.RawQuery = sourceReq.URL.Query().Encode()

	p.Logger.Info("Building proxy request", log.String("method", sourceReq.Method), log.String("url", sscURL.String()))
	proxyReq, err := http.NewRequest(sourceReq.Method, sscURL.String(), bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	proxyReq.Header.Add("Authorization", "Bearer "+token)
	return proxyReq, nil
}

// getSAMSCredentialsForUser fetches the SAMS identity for the the given Sourcegraph user ID, and
// decrypts the OAuth token stored within.
func (p *APIProxyHandler) getSAMSCredentialsForUser(ctx context.Context, userID int32) (
	*extsvc.Account, *oauth2.Token, error) {
	// NOTE: It's possible for a user to have multiple SAMS identities attached to the same Sourcegraph
	// user account. The underlying implementation provides a stable result sorting by ID, so we
	// just return the first SAMS identity found.
	//
	// BUG: This needs to be reconciled with the logic in internal/ssc/subscription.go. Since when it
	// comes to returning the user's Cody Pro subscription, we take *ALL* SAMS identities into account.
	// So in order to provide a consistent view of a user's Cody Pro subscription data, we need to
	// ensure that for the ~80 or so users in this situation can safely use their "first" SAMS ID when
	// they are sorted lexographically.
	extAccounts, err := p.DB.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		UserID:      userID,
		ServiceType: "openidconnect",
		// We expect the SAMS backend origin to match the registered identity provider,
		// e.g. "https://accounts.sourcegraph.com" or "http://localhost:9992".
		ServiceID:      p.CodyProConfig.SamsBackendOrigin,
		ExcludeExpired: true,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "listing user external accounts")
	}
	switch len(extAccounts) {
	case 0:
		return nil, nil, errors.New("user does not have a SAMS identity")
	case 1:
		// Expected, AOK
	default:
		// Exceptional case. The user has multiple identities, and we will just take
		// the first. (Which may be confusing the user in some circumstances.)
		p.Logger.Warn("user has multiple SAMS identities", log.Int32("uid", userID))
	}

	// Load the specific external account (SAMS identity).
	samsIdentity, err := p.DB.UserExternalAccounts().Get(ctx, extAccounts[0].ID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting user SAMS identity")
	}

	// Decrypt and unmarshall as an OAuth token.
	token, err := encryption.DecryptJSON[oauth2.Token](ctx, samsIdentity.AuthData)
	if err != nil {
		return nil, nil, errors.Wrap(err, "decrypting/unmarshalling SAMS auth data")
	}
	return samsIdentity, token, nil
}

// tryRefreshSAMSCredentials attempts to refresh the user's SAMS credentials by
// exchanging the OAuth refresh token we have on file for a new access/refresh token.
//
// Upon success, the new tokens will be persisted in the database.
func (p *APIProxyHandler) tryRefreshSAMSCredentials(
	ctx context.Context, samsIdent *extsvc.Account, currentToken *oauth2.Token) (string, error) {
	if samsIdent == nil || currentToken == nil {
		return "", errors.New("current identity or current token not provided")
	}

	externalAccountID := samsIdent.ID // ID of the external identity, not the user ID.
	refreshFn := oauthtoken.GetAccountRefreshAndStoreOAuthTokenFunc(
		p.DB.UserExternalAccounts(), externalAccountID, p.SAMSOAuthContext)

	// Perform the refresh.
	userBearerToken := auth.OAuthBearerToken{
		Token:        currentToken.AccessToken,
		RefreshToken: currentToken.RefreshToken,
		Expiry:       currentToken.Expiry,
	}
	newToken, _ /* newRefreshToken */, newExpiry, err := refreshFn(
		ctx, httpcli.UncachedExternalDoer, &userBearerToken)
	if err != nil {
		return "", errors.Wrap(err, "refreshing SAMS token")
	}

	p.Logger.Info("refresh user's SAMS token", log.Time("new expiration", newExpiry))
	return newToken, nil
}

func (p *APIProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p.Logger.Info("proxying SSC API request", log.String("url", r.URL.String()))

	// Confirm the proxy is configured correctly.
	if p.CodyProConfig == nil || p.SAMSOAuthContext == nil || p.SAMSOAuthContext.ClientID == "" {
		http.Error(w, "proxy not configured", http.StatusServiceUnavailable)
		return
	}

	sgUserID, err := p.getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Lookup the user's SAMS credentials.
	samsIdentity, samsToken, err := p.getSAMSCredentialsForUser(ctx, sgUserID)
	if err != nil {
		// Here we assume that the function will only fail because of an IO problem.
		// And not that a user simply doesn't have a SAMS identity. (Since for dotcom
		// that is guaranteed to be the case.)
		p.Logger.Error("getting SAMS credentials for user", log.Int32("uid", sgUserID), log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// SAMS credentials have a shorter lifetime than the Sourcegraph session.
	// So it's likely that even though the user is still logged into sourcegraph.com,
	// the request to the SSC backend will fail because the OAuth credentials
	// associated with their SAMS login has expired.
	//
	// If we detect this, we first try to refresh the user's SAMS access token. But
	// that may also fail (if the underlying SAMS refresh token has also expired). So
	// the frontend must expect this situation via a 401 response, and force the user
	// to reauthenticate. (Which would then pick up a new SAMS auth token.)
	accessToken := samsToken.AccessToken
	if samsToken.Expiry.Before(time.Now()) {
		p.Logger.Warn("the user's SAMS token has expired", log.Time("expiry", samsToken.Expiry))

		newToken, err := p.tryRefreshSAMSCredentials(ctx, samsIdentity, samsToken)
		if err != nil {
			p.Logger.Error("error trying to refresh the user's SAMS credentials", log.Error(err))

			// Just fail here since there is nothing we can do. We know the token is invalid,
			// and we were unable to create a new one.
			http.Error(w, "Sourcegraph Accounts identity has expired", http.StatusUnauthorized)
			return
		}
		accessToken = newToken
	}

	// Copy the incoming request and send it to the SSC backend.
	proxyRequest, err := p.buildProxyRequest(r, accessToken)
	if err != nil {
		p.Logger.Error("building SSC proxy request", log.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	client := http.DefaultClient
	proxyResponse, err := client.Do(proxyRequest)
	if err != nil {
		p.Logger.Error("sending SSC proxy request", log.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var bodyBytes []byte
	if bodyReader := proxyResponse.Body; bodyReader != nil {
		bodyBytes, err = io.ReadAll(bodyReader)
		if err != nil {
			p.Logger.Error("reading SSC response", log.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if err = bodyReader.Close(); err != nil {
			p.Logger.Error("closing SSC response body", log.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// For any SSC 5xx errors, surface that as a 502 from this Sourcegraph instance,
	// to make it clear that the underlying error didn't happen "here".
	if proxyResponse.StatusCode >= 500 {
		p.Logger.Error(
			"received 5xx response from SSC backend",
			log.String("responseBody", string(bodyBytes)))
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	// Success! Serve the proxied response.
	p.Logger.Debug("serving proxied response from the SSC backend",
		log.Int("code", proxyResponse.StatusCode),
		log.Int("bodySize", len(bodyBytes)))
	w.WriteHeader(proxyResponse.StatusCode)
	if _, err = w.Write(bodyBytes); err != nil {
		p.Logger.Error("writing proxied response body", log.Error(err))
	}
}
