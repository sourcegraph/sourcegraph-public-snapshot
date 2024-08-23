package accountsv1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

// NewClient constructs a new SAMS Accounts client, pointed to the supplied SAMS host.
// e.g. "https://accounts.sourcegraph.com".
//
// Users should prefer to use the top-level 'sams.NewAccountsV1' constructor instead.
func NewClient(samsHost string, tokenSource oauth2.TokenSource) *Client {
	// Canonicalize the host so we only need to check if it ends in a slash or not once.
	samsHost = strings.ToLower(samsHost)
	samsHost = strings.TrimSuffix(samsHost, "/")

	return &Client{
		host:        samsHost,
		tokenSource: tokenSource,
	}
}

// Client is a wrapper around SAMS primitive REST-based Accounts API. Most
// likely, you want to use the more service-to-service "Clients" API instead.
//
// This API is needed when using SAMS to identify users, but not perform
// authorization checks. e.g. the caller will handle its own authorization
// checks based on the identity of the SAMS user. (The returned User.Sub, the
// SAMS account external ID.)
type Client struct {
	host        string
	tokenSource oauth2.TokenSource
}

// GetUser returns the basic user profile of the calling user. (Who owns the
// underlying token or TokenSource the client is using for authentication.)
//
// If the supplied token is invalid, malformed, or expired, the error will contain
// "unexpected status 401".
func (c *Client) GetUser(ctx context.Context) (*User, error) {
	url := fmt.Sprintf("%s/api/v1/user", c.host)

	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, errors.Wrap(err, "getting token")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil /* body */)
	if err != nil {
		return nil, errors.Wrap(err, "creating SAMS user details request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Add("User-Agent", "sourcegraph-accounts-sdk-go/1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fetching user details")
	}
	if resp.Body == nil {
		return nil, errors.New("no response body")
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading response body")
	}
	if err = resp.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "closing response body")
	}
	if resp.StatusCode != http.StatusOK {
		unexpectedRespErr := errors.Errorf(
			"unexpected status %d (response body: %s)",
			resp.StatusCode, string(bodyBytes))
		return nil, unexpectedRespErr
	}

	var user User
	if err = json.Unmarshal(bodyBytes, &user); err != nil {
		return nil, errors.Wrap(err, "unmarshalling response")
	}

	return &user, nil
}
