package client

import (
	"log"
	"sync"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	sgxcli "sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/userauth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

var credentials CredentialOpts

func init() {
	sgxcli.CLI.AddGroup("Client authentication", "", &credentials)
}

// CredentialOpts sets the authentication credentials to use when
// contacting the Sourcegraph server's API.
type CredentialOpts struct {
	AuthFile    string `long:"auth-file" description:"path to .src-auth" default:"$HOME/.src-auth" env:"SRC_AUTH_FILE"`
	AccessToken string `long:"token" description:"access token (OAuth2)" env:"SRC_TOKEN"`
	once        sync.Once
}

// WithCLICredentials sets the HTTP and gRPC credentials in the context.
func WithCLICredentials(ctx context.Context) context.Context {
	credentials.once.Do(func() {
		if credentials.AuthFile != "" && credentials.AccessToken == "" { // AccessToken takes precedence over AuthFile
			userAuth, err := userauth.Read(credentials.AuthFile)
			if err != nil {
				log.Fatal(err)
			}

			if ua := userAuth[Endpoint.URLOrDefault().String()]; ua != nil {
				credentials.AccessToken = ua.AccessToken
			}
		}
	})

	return sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: credentials.AccessToken}))
}

// AuthFile returns the path of the file that stores authentications.
func AuthFile() string {
	return credentials.AuthFile
}
