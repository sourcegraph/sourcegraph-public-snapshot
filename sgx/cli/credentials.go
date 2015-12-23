package cli

import (
	"sync"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"src.sourcegraph.com/sourcegraph/auth/userauth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

var Credentials CredentialOpts

func init() {
	CLI.AddGroup("Client authentication", "", &Credentials)
}

// CredentialOpts sets the authentication credentials to use when
// contacting the Sourcegraph server's API.
type CredentialOpts struct {
	AuthFile string `long:"auth-file" description:"path to .src-auth" default:"$HOME/.src-auth" env:"SRC_AUTH_FILE"`

	mu sync.RWMutex

	// AccessToken should be accessed via the GetAccessToken and SetAccessToken
	// methods which synchronize access to this value.
	AccessToken string `long:"token" description:"access token (OAuth2)" env:"SRC_TOKEN"`
}

// WithCredentials sets the HTTP and gRPC credentials in the context.
func (c *CredentialOpts) WithCredentials(ctx context.Context) (context.Context, error) {
	token := c.GetAccessToken()
	if token == "" && c.AuthFile != "" { // AccessToken takes precedence over AuthFile
		userAuth, err := userauth.Read(c.AuthFile)
		if err != nil {
			return nil, err
		}

		ua := userAuth[Endpoint.URLOrDefault().String()]
		if ua != nil {
			c.SetAccessToken(ua.AccessToken)
		}
	}

	token = c.GetAccessToken()
	if token != "" {
		ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: token}))
	}

	return ctx, nil
}

// GetAccessToken returns the currently set AccessToken.
//
// It synchronizes access to the token by acquiring a read lock.
func (c *CredentialOpts) GetAccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.AccessToken
}

// SetAccessToken sets a new access token.
//
// It synchronizes access to the token by acquiring a write lock.
func (c *CredentialOpts) SetAccessToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.AccessToken = token
}
