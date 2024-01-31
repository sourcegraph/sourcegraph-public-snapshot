package ssc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SAMSProdHostname is the hostname for the SAMS production instance.
const SAMSProdHostname = "accounts.sourcegraph.com"

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
	baseURL string

	// httpClient to use for making requests to SSC. It will have a transport configured
	// to automatically issue or reuse SAMS access tokens as needed.
	httpClient *http.Client
}

// Validate inspects the client configuration and ensures it is valid.
func (c *client) Validate() error {
	if c.baseURL == "" {
		return errors.New("no SCC base URL provided")
	}
	if c.httpClient == nil {
		return errors.New("no HTTP client found")
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

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "calling SSC")
	}
	defer resp.Body.Close()

	if outBody != nil {
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
	default:
		return nil, errors.Errorf("unexpected status code %d", *code)
	}
}

// NewClient returns a new SSC API client. It is important to avoid creating new
// API clients if possible, so that it can reuse SAMS access tokens when making
// requests to SSC. (Otherwise every request would need to create a new token,
// adding unnecessary latency.)
//
// If no SAMS authorization provider is configured, this function will not panic,
// but instead will return an error on every call.
func NewClient() Client {
	sgconf := conf.Get().SiteConfig()

	// Fetch the SAMS configuration data.
	var samsConfig *clientcredentials.Config
	for _, provider := range conf.Get().AuthProviders {
		oidcInfo := provider.Openidconnect
		if oidcInfo == nil {
			continue
		}

		if strings.Contains(oidcInfo.Issuer, SAMSProdHostname) {
			samsConfig = &clientcredentials.Config{
				ClientID:     oidcInfo.ClientID,
				ClientSecret: oidcInfo.ClientSecret,
				TokenURL:     fmt.Sprintf("%s/oauth/token", oidcInfo.Issuer),
				Scopes:       []string{"client.ssc"},
			}
			break
		}
	}

	// Create a long-lived HTTP client for all dotcom<->SSC requests, using
	// the samsConfig credentials as needed.
	httpClient := samsConfig.Client(context.Background())

	return &client{
		baseURL:    sgconf.SscApiBaseUrl, // [sic] generated code
		httpClient: httpClient,
	}
}
