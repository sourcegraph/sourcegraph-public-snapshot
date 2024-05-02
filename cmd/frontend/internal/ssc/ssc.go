package ssc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client is the interface for making requests to the Self-Service Cody backend.
// This uses a REST API exposed from the service, and not the GraphQL API (which is
// instead used for user-facing frontend queries.)
type Client interface {
	// FetchSubscriptionBySAMSAccountID will return the SSC subscription information associated with
	// the given SAMS user account. Will return (nil, nil) if the user ID is valid, but has no
	// SSC subscription information associated with their account. Or an error if the SAMS account
	// ID is unrecognized.
	//
	// In the future, this behavior will change to _ALWAYS_ return subscription data, even for
	// users that have not converted. But for now, the `cody` package needs to know if the user
	// has setup their Cody Pro subscription on the SSC backend or not.
	FetchSubscriptionBySAMSAccountID(ctx context.Context, samsAccountID string) (*Subscription, error)
}

type client struct {
	baseURL         string
	samsTokenSource oauth2.TokenSource
}

// Validate inspects the client configuration and ensures it is valid.
func (c *client) Validate() error {
	if c.baseURL == "" {
		return errors.New("no SCC base URL provided")
	}
	if c.samsTokenSource == nil {
		return errors.New("no SAMS token source available")
	}
	return nil
}

// sendRequest issues an HTTP request to SSC. If supplied, the response will be unmarshalled into outBody as JSON.
func (c *client) sendRequest(ctx context.Context, method string, url string, outBody *Subscription) (*int, error) {
	// Build and send the request.
	req, err := http.NewRequestWithContext(ctx, method, url, nil /* body */)
	if err != nil {
		return nil, err
	}

	// Create the OAuth2 HTTP client. This will handle setting up the HTTP headers based on the token source.
	// (And potentially issue a new token as needed.)
	httpClient := oauth2.NewClient(ctx, c.samsTokenSource)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "calling SSC")
	}
	defer resp.Body.Close()

	if outBody != nil && resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return &resp.StatusCode, errors.Wrap(err, "reading response")
		}

		err = json.Unmarshal(bodyBytes, outBody)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshalling response")
		}
	}

	return &resp.StatusCode, nil
}

func (c *client) FetchSubscriptionBySAMSAccountID(ctx context.Context, samsAccountID string) (*Subscription, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	var subscription Subscription
	url := fmt.Sprintf("%s/rest/svc/subscription/%s", c.baseURL, samsAccountID)

	tr, traceCtx := trace.New(ctx, "sccSendRequest", attribute.String("url", url))
	code, err := c.sendRequest(traceCtx, http.MethodGet, url, &subscription)
	if err != nil {
		tr.EndWithErr(&err)
		return nil, err
	}
	tr.End()

	subscription.Status = SubscriptionStatus(strings.ToUpper(subscription.StatusRaw))

	switch *code {
	case http.StatusOK:
		// User has an SSC subscription.
		return &subscription, nil
	case http.StatusNoContent:
		// User is valid, but does not have an SSC subscription.
		return nil, nil
	case http.StatusNotFound:
		// User is not found on SSC. This is not a valid state, but we should handle it gracefully.
		return nil, nil
	default:
		return nil, errors.Errorf("unexpected status code %d", *code)
	}
}

// getSSCBaseURL returns the base URL for the SSC backend's REST API for
// service-to-service requests.
func getSSCBaseURL() string {
	config := conf.Get()

	// Prefer the newer "dotcom.codyProConfig.sscBackendOrigin" config setting if available.
	// This allows for local development (not hard-coding the https scheme).
	if dotcomConfig := config.Dotcom; dotcomConfig != nil {
		if codyProConfig := dotcomConfig.CodyProConfig; codyProConfig != nil {
			return fmt.Sprintf("%s/cody/api", codyProConfig.SscBackendOrigin)
		}
	}

	// Fall back to original logic, using the "ssc.apiBaseUrl" setting.
	// (To be removed when the codyProConfig changes are in production.)
	siteConfig := config.SiteConfig()
	baseURL := siteConfig.SscApiBaseUrl
	if baseURL == "" {
		baseURL = "https://accounts.sourcegraph.com/cody/api"
	}

	return baseURL
}

// GetSAMSServiceID returns the ServiceID of the currently registered SAMS identity provider.
// This is found in the site configuration, and must match the auth.providers configuration
// exactly.
func GetSAMSServiceID() string {
	config := conf.Get()

	// Prefer the newer "dotcom.codyProConfig.samsBackendOrigin" config setting if available.
	// This allows for local development (not hard-coding the https scheme).
	if dotcomConfig := config.Dotcom; dotcomConfig != nil {
		if codyProConfig := dotcomConfig.CodyProConfig; codyProConfig != nil {
			return codyProConfig.SamsBackendOrigin
		}
	}

	// Fallback to the original logic, using the "ssc.samsHostName" setting.
	// (To be removed when the codyProConfig changes are in production.)
	sgconf := config.SiteConfig()
	if sgconf.SscSamsHostName == "" {
		// If unset, default to the production hostname.
		return "https://accounts.sourcegraph.com"
	}
	return fmt.Sprintf("https://%s", sgconf.SscSamsHostName)
}

// NewClient returns a new SSC API client. It is important to avoid creating new
// API clients if possible, so that it can reuse SAMS access tokens when making
// requests to SSC. (Otherwise every request would need to create a new token,
// adding unnecessary latency.)
//
// If no SAMS authorization provider is configured, this function will not panic,
// but instead will return an error on every call.
func NewClient() (Client, error) {
	// Fetch the SAMS configuration data.
	var samsConfig *clientcredentials.Config
	for _, provider := range conf.Get().AuthProviders {
		oidcInfo := provider.Openidconnect
		if oidcInfo == nil {
			continue
		}

		if oidcInfo.Issuer == GetSAMSServiceID() {
			samsConfig = &clientcredentials.Config{
				ClientID:     oidcInfo.ClientID,
				ClientSecret: oidcInfo.ClientSecret,
				TokenURL:     fmt.Sprintf("%s/oauth/token", oidcInfo.Issuer),
				Scopes:       []string{"client.ssc"},
			}
			break
		}
	}

	if samsConfig == nil {
		return &client{}, errors.New("no SAMS authorization provider configured")
	}

	// We want this tokenSource to be long lived, so we benefit from reusing existing
	// SAMS tokens if repeated requests are made within the token's lifetime. (Under
	// the hood it returns an oauth2.ReuseTokenSource.)
	tokenSource := samsConfig.TokenSource(context.Background())
	return &client{
		baseURL:         getSSCBaseURL(),
		samsTokenSource: tokenSource,
	}, nil
}
