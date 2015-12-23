package sgx

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"sourcegraph.com/sourcegraph/grpccache"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

var ClientContextFuncs []func(context.Context) context.Context

func init() {
	sourcegraph.Cache = &grpccache.Cache{
		MaxSize: 150 * (1024 * 1024), // 150 MB
		KeyPart: func(ctx context.Context) string {
			// Separate caches based on the authorization level, to
			// avoid cross-user/client leakage.
			//
			// The authorization metadata can be in EITHER:
			//  (a) the client credentials, if the client was created
			//      normally. This is tried first.
			//  (b) the ctx's metadata, if the client was created by the
			//      remote services in a federated request. This is tried
			//      second in order to properly cache federated requests.
			//
			// NOTE: This is important code. If we don't get
			// the right auth data, we could leak sensitive data
			// across authentication boundaries.

			key := sourcegraph.GRPCEndpoint(ctx).String() + ":"

			if cred := sourcegraph.CredentialsFromContext(ctx); cred != nil {
				md, err := (oauth.TokenSource{TokenSource: cred}).GetRequestMetadata(ctx)
				if err != nil {
					log.Printf("Determining cache key with token auth failed: %s. Caching will not be performed.", err)
					// Use a random string to prevent caching.
					return "err#" + randstring.NewLen(64)
				}
				key += md["authorization"]
			} else if md, ok := metadata.FromContext(ctx); ok {
				// if ctx metadata contains auth token, use it in the cache key
				if authMD, ok := md["authorization"]; ok && len(authMD) > 0 {
					key += authMD[len(authMD)-1]
				}
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
	return sourcegraph.NewClientFromContext(cli.Ctx)
}

func init() {
	cli.CLI.InitFuncs = append(cli.CLI.InitFuncs, func() {
		// The "src version" command does not need a cli context at all.
		if cli.CLI.Active != nil && cli.CLI.Active.Name == "version" {
			return
		}
		cli.Ctx = WithClientContext(context.Background())
	})
}

// WithClientContext returns a copy of parent with client endpoint and
// auth information added.
func WithClientContext(parent context.Context) context.Context {
	// The "src serve" command is the only non-client command; it
	// must not have credentials set (because it is not a client
	// command).
	if cli.CLI.Active != nil && cli.CLI.Active.Name == "serve" {
		cli.Credentials.AuthFile = ""
	}
	ctx := WithClientContextUnauthed(parent)
	ctx, err := cli.Credentials.WithCredentials(ctx)
	if err != nil {
		log.Fatalf("Error constructing API client credentials: %s.", err)
	}
	return ctx
}

// WithClientContextUnauthed returns a copy of parent with client endpoint added.
func WithClientContextUnauthed(parent context.Context) context.Context {
	return cli.Endpoint.NewContext(parent)
}
