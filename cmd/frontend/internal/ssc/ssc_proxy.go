package ssc

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/session"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SSCAPIProxy is an HTTP handler that essentially proxies API requests from the
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
}

var _ http.Handler = (*APIProxyHandler)(nil)

// getUserIDFromRequest extracts the Sourcegraph User ID from the incomming request,
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

// buildProxyRequest converts the incomming HTTP request into what will be sent to the SSC backend.
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

// getSAMSCredentialsForUser fetches the OAuth token from the user's SAMS external identity.
func (p *APIProxyHandler) getSAMSCredentialsForUser(ctx context.Context, userID int32) (*oauth2.Token, error) {
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
		return nil, errors.Wrap(err, "listing user external accounts")
	}
	switch len(extAccounts) {
	case 0:
		return nil, errors.New("user does not have a SAMS identity")
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
		return nil, errors.Wrap(err, "getting user SAMS identity")
	}

	// Decrypt and unmarshall as an OAuth token.
	token, err := encryption.DecryptJSON[oauth2.Token](ctx, samsIdentity.AuthData)
	if err != nil {
		return nil, errors.Wrap(err, "decrypting/unmarshalling SAMS auth data")
	}
	return token, nil
}

func (p *APIProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var info struct {
		Actor *actor.Actor `json:"actor"`
	}
	if err := session.GetData(r, "actor", &info); err != nil {
		p.Logger.Error("failed to get actor from session", log.Error(err))
	}
	// TODO: we should set actor to context in another way,
	// see: https://github.com/sourcegraph/sourcegraph/blob/24c0b99b65161297c41b9836bbd80f7811daae20/cmd/frontend/internal/httpapi/httpapi.go#L217-L229
	ctx := actor.WithActor(r.Context(), info.Actor)

	p.Logger.Info("proxying SSC API request", log.String("url", r.URL.String()))

	sgUserID, err := p.getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Lookup the user's SAMS credentials.
	samsToken, err := p.getSAMSCredentialsForUser(ctx, sgUserID)
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
	// The frontend needs to expect this 401 response, and force the user to
	// reauthenticate. (Which would then pick up a new SAMS auth token.) Or
	// we need to have a background process that will periodically refresh the
	// user's SAMS credentials, so that the refresh and access token for the
	// user's SAMS account are sufficiently fresh.
	if samsToken.Expiry.Before(time.Now()) {
		p.Logger.Warn("the user's SAMS token has expired", log.Time("expiry", samsToken.Expiry))
		http.Error(w, "Sourcegraph Accounts identity has expired", http.StatusUnauthorized)
		return
	}

	// Copy the incoming request and send it to the SSC backend.
	proxyRequest, err := p.buildProxyRequest(r, samsToken.AccessToken)
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
