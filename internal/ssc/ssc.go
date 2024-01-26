package ssc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// FeatureFlagUseSSCDevInstance, if set, will send requests to the SSC dev environment for dogfooding.
//
// IMPORTANT: If this is set, calls to Client.LookupDotcomUserSAMSAccountID return the SAMS account ID
// for the SAMS dev environment. (Since SSC will be using that for authentication.)
const FeatureFlagUseSSCDevInstance = "ssc.use-dev-environment"

const (
	samsProdHostname = "accounts.sourcegraph.com"
	samsDevHostname  = "accounts.sgdev.org"
)

// Client is the interface for making requests to the Self-Service Cody backend.
// This uses a REST API exposed from the service, and not the GraphQL API (which is
// instead used for user-facing frontend queries.)
type Client interface {
	// LookupDotcomUserSAMSAccountID returns the SAMS external account ID for the given dotcom user ID.
	// Will return "" if the user has no SAMS identity associated with their dotcom user account.
	//
	// This will honor the FeatureFlagUseSSCDevInstance. So if set, the returned SAMS account ID
	// will be from the SAMS dev environment rather than production..
	LookupDotcomUserSAMSAccountID(ctx context.Context, db database.DB, dotcomUserID int32) (string, error)

	// FetchSubscriptionBySAMSAccountID will return the SSC subscription information associated with
	// the given SAMS user account. Will return (nil, nil) if the user ID is valid, but has no
	// SSC subscription information associated with their account.
	//
	// In the future, this behavior will change to _ALWAYS_ return subscription data, even for
	// users that have not converted. But for now, the `cody` package needs to know if the user
	// has setup their Cody Pro subscription on the SSC backend or not.
	FetchSubscriptionBySAMSAccountID(ctx context.Context, samsAccountID string) (*Subscription, error)
}

func NewClient() Client {
	sgconf := conf.Get().SiteConfig()
	return &client{
		baseURL: sgconf.SscApiBaseUrl, // [sic] generated code
	}
}

type client struct {
	baseURL     string
	secretToken string
}

// createSAMSToken creates a new SAMS access token for contacting the SSC backend.
func (c *client) createSAMSToken(ctx context.Context) (string, error) {
	// The target SSC environment needs to match the SAMS environment being used.
	useSSCDevInstance := featureflag.FromContext(ctx).GetBoolOr(FeatureFlagUseSSCDevInstance, false)

	var sams *schema.OpenIDConnectAuthProvider
	for _, provider := range conf.Get().AuthProviders {
		oidcInfo := provider.Openidconnect
		if oidcInfo == nil {
			continue
		}
		if useSSCDevInstance {
			if strings.Contains(oidcInfo.Issuer, samsDevHostname) {
				sams = oidcInfo
				break
			}
		} else {
			if strings.Contains(oidcInfo.Issuer, samsProdHostname) {
				sams = oidcInfo
				break
			}
		}
	}
	if sams == nil {
		return "", errors.New("no appropriate SAMS auth provider found")
	}

	// BUG: We shouldn't create a new token for every request, and instead use the proper
	// libraries to cache and refresh automatically. (But that would require that the
	// SSC client is long-lived, meaning we need to fix that in the cody package first.)
	clientCreds := &clientcredentials.Config{
		ClientID:     sams.ClientID,
		ClientSecret: sams.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth/token", sams.Issuer),

		// IMPORTANT: The SAMS client must be capable of issuing tokens with the
		// client.ssc scope.
		Scopes: []string{"client.ssc"},
	}
	token, err := clientCreds.Token(ctx)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (c *client) sendRequest(ctx context.Context, method string, url string, body *Subscription) (code *int, err error) {
	samsToken, err := c.createSAMSToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "creating SAMS token")
	}

	// Issue the request.
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", samsToken))

	resp, err := httpcli.UncachedExternalDoer.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return &resp.StatusCode, err
		}

		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
	}

	return &resp.StatusCode, nil

}

// LookupDotcomUserSAMSAccountID attempts to lookup the dotcom user's SAMS account ID from their attached
// identities. It's possible the user hasn't yet attached a SAMS identity.
func (c *client) LookupDotcomUserSAMSAccountID(ctx context.Context, db database.DB, dotcomUserID int32) (string, error) {
	// Fetch all of the user's OpenID Connect identities. We expect a user can have several of these,
	// such as identities from both the SAMS-dev and SAMS-prod instances.
	oidcIdentities, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		UserID:      dotcomUserID,
		ServiceType: "openidconnect",
		LimitOffset: &database.LimitOffset{
			Limit: 10,
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "listing external accounts")
	}
	if len(oidcIdentities) == 0 {
		return "", nil
	}

	// Loop through the available identities and pick out the SAMS identities.
	var (
		samsProdAccountID string
		samsDevAccountID  string
	)
	for _, identity := range oidcIdentities {
		if strings.Contains(identity.URL, samsProdHostname) {
			samsProdAccountID = identity.AccountID
		} else if strings.Contains(identity.URL, samsDevHostname) {
			samsDevAccountID = identity.AccountID
		}
	}

	useSSCDevInstance := featureflag.FromContext(ctx).GetBoolOr(FeatureFlagUseSSCDevInstance, false)
	if useSSCDevInstance {
		return samsDevAccountID, nil
	}
	return samsProdAccountID, nil
}

// FetchSubscriptionBySAMSAccountID returns the user's Cody subscription for the sams_account_id.
// It returns nil, nil if the user does not have a Cody Pro subscription.
func (c *client) FetchSubscriptionBySAMSAccountID(ctx context.Context, samsAccountID string) (*Subscription, error) {
	if c.baseURL == "" {
		return nil, errors.New("SSC base url is not set")
	}

	if c.secretToken == "" {
		return nil, errors.New("SSC secret token is not set")
	}

	var subscription Subscription
	url := fmt.Sprintf("%s/rest/svc/subscription/%s", c.baseURL, samsAccountID)
	code, err := c.sendRequest(ctx, http.MethodGet, url, &subscription)
	if err != nil {
		return nil, err
	}

	subscription.Status = SubscriptionStatus(strings.ToUpper(subscription.StatusRaw))

	if *code == http.StatusOK {
		return &subscription, nil
	}

	// 204 response indicates that the user does not have a Cody Pro subscription
	if *code == http.StatusNoContent {
		return nil, nil
	}

	return nil, errors.Errorf("unexpected status code %d while fetching user subscription from SSC", *code)
}
