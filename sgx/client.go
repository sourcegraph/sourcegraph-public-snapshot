package sgx

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/credentials"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/grpccache"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

var ClientContextFuncs []func(context.Context) context.Context

func init() {
	cli.CLI.AddGroup("Client API endpoint", "", &Endpoints)
	cli.CLI.AddGroup("Client authentication", "", &Credentials)
}

var Endpoints conf.EndpointOpts

// CredentialOpts sets the authentication credentials to use when
// contacting the Sourcegraph server's API.
type CredentialOpts struct {
	AccessToken string `long:"token" description:"access token (OAuth2)" env:"SRC_TOKEN"`
	AuthFile    string `long:"auth-file" description:"path to .src-auth" default:"$HOME/.src-auth" env:"SRC_AUTH_FILE"`
}

var Credentials CredentialOpts

// WithCredentials sets the HTTP and gRPC credentials in the context.
func (c *CredentialOpts) WithCredentials(ctx context.Context) (context.Context, error) {
	if c.AccessToken == "" && c.AuthFile != "" { // AccessToken takes precedence over AuthFile
		userAuth, err := readUserAuth()
		if err != nil {
			return nil, err
		}

		// Prefer explicitly specified endpoint, then auth file default endpoint,
		// then fallback default. For this reason, we use Endpoint not EndpointURL
		// (which provides the fallback default) here.
		ua := userAuth[Endpoints.Endpoint]
		if Endpoints.Endpoint == "" {
			if ua == nil {
				var ep string
				ep, ua = userAuth.getDefault()
				if ep != "" {
					Endpoints.Endpoint = ep
				}
			}
			if ua == nil {
				ua = userAuth[Endpoints.EndpointURL().String()]
			}
		}
		if ua != nil {
			c.AccessToken = ua.AccessToken
		}
	}

	if c.AccessToken != "" {
		ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: c.AccessToken}))
	}

	return ctx, nil
}

func init() {
	sourcegraph.Cache = &grpccache.Cache{
		MaxSize: 150 * (1024 * 1024), // 150 MB
		KeyPart: func(ctx context.Context) string {
			// Separate caches based on the authorization level, to
			// avoid cross-user/client leakage.
			//
			// The authorization metadata can be in EITHER (a) the
			// ctx's metadata (if the client was created by the remote
			// services in a federated request) or (b) in the client
			// credentials (if the client was created normally). But
			// the auth data from (b) is what'll actually be used--(a)
			// is just a remnant from the original server handler. So,
			// use (b). NOTE: This is important code. If we don't get
			// the right auth data, we could leak sensitive data
			// across authentication boundaries.

			key := sourcegraph.GRPCEndpoint(ctx).String() + ":"

			if cred := sourcegraph.CredentialsFromContext(ctx); cred != nil {
				md, err := (credentials.TokenSource{TokenSource: cred}).GetRequestMetadata(ctx)
				if err != nil {
					log.Printf("Determining cache key with token auth failed: %s. Caching will not be performed.", err)
					// Use a random string to prevent caching.
					return "err#" + randstring.NewLen(64)
				}
				key += md["authorization"]
			}

			s := sha256.Sum256([]byte(key))
			return base64.StdEncoding.EncodeToString(s[:])
		},
		Log: os.Getenv("CACHE_LOG") != "",
	}
}

// Client returns a Sourcegraph API client configured to use the
// specified endpoints and authentication info.
func Client() *sourcegraph.Client {
	return sourcegraph.NewClientFromContext(cliCtx)
}

// cliCtx is a context that accesses the configured
// Sourcegraph endpoint with the configured credentials.
var cliCtx context.Context

func init() {
	cli.CLI.InitFuncs = append(cli.CLI.InitFuncs, func() {
		// The "src version" command does not need a cli context at all.
		if cli.CLI.Active != nil && cli.CLI.Active.Name == "version" {
			return
		}
		cliCtx = context.Background()
		// The "src serve" command is the only non-client command; it
		// must not have credentials set (because it is not a client
		// command) or a default discovery endpoint (because it is
		// what creates the endpoint, and the check would occur before
		// the server could start).
		if cli.CLI.Active != nil && cli.CLI.Active.Name == "serve" {
			Credentials.AuthFile = ""
		}
		cliCtx = WithClientContext(cliCtx)
	})
}

// WithClientContext returns a copy of parent with client endpoint and
// auth information added.
func WithClientContext(parent context.Context) context.Context {
	ctx, err := Credentials.WithCredentials(parent)
	if err != nil {
		log.Fatalf("Error constructing API client credentials: %s.", err)
	}
	ctx, err = Endpoints.WithEndpoints(ctx)
	if err != nil {
		log.Fatalf("Error constructing API client endpoints: %s.", err)
	}
	return ctx
}
