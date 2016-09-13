package cli

import (
	"log"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/userauth"

	"context"
)

// cliContext and cliClient accesses the configured Sourcegraph endpoint
// with the configured credentials. They should be used for all CLI
// operations.
var cliContext context.Context
var cliClient *sourcegraph.Client

func init() {
	cli.CLI.InitFuncs = append(cli.CLI.InitFuncs, func() {
		skipGRPC := map[string]bool{
			// "src version" command does not need a context
			// at all.
			"version": true,
			// "src serve" creates its own (and its server would
			// not have started anyway by the time this client
			// would attempt to connect).
			"serve": true,
			// "src git-server" does not need a context at all.
			"git-server": true,
		}
		if cli.CLI.Active != nil && skipGRPC[cli.CLI.Active.Name] {
			return
		}

		cliContext = withCLICredentials(sourcegraph.WithGRPCEndpoint(context.Background(), endpoint.URLOrDefault()))

		var err error
		cliClient, err = sourcegraph.NewClientFromContext(cliContext)
		if err != nil {
			log.Fatalf("could not create client: %s", err)
		}
	})
}

var credentials CredentialOpts

func init() {
	cli.CLI.AddGroup("Client authentication", "", &credentials)
}

// CredentialOpts sets the authentication credentials to use when
// contacting the Sourcegraph server's API.
type CredentialOpts struct {
	AuthFile    string `long:"auth-file" description:"path to .src-auth" default:"$HOME/.src-auth" env:"SRC_AUTH_FILE"`
	AccessToken string `long:"token" description:"access token (OAuth2)" env:"SRC_TOKEN"`
	once        sync.Once
}

// WithCLICredentials sets the HTTP and gRPC credentials in the context.
func withCLICredentials(ctx context.Context) context.Context {
	credentials.once.Do(func() {
		if credentials.AuthFile != "" && credentials.AccessToken == "" { // AccessToken takes precedence over AuthFile
			userAuth, err := userauth.Read(credentials.AuthFile)
			if err != nil {
				log.Fatal(err)
			}

			if ua := userAuth[endpoint.URLOrDefault().String()]; ua != nil {
				credentials.AccessToken = ua.AccessToken
			}
		}
	})

	return sourcegraph.WithAccessToken(ctx, credentials.AccessToken)
}

// AuthFile returns the path of the file that stores authentications.
func authFile() string {
	return credentials.AuthFile
}
