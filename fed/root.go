package fed

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/fed/discover"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Flags defines settings (in the form of CLI flags) for federation.
type Flags struct {
	// RootInstance is the base URL of the root server in the Sourcegraph
	// network. It is used for OAuth2 user authentication, default repo
	// graph data lookups, etc.
	RootURLStr string `long:"fed.root-url" description:"base URL of the root Sourcegraph server (used for OAuth2 user auth, open-source repo data, etc.)" default:"https://sourcegraph.com"`

	RootGRPCURLStr string `long:"fed.root-grpc-url" description:"gRPC Endpoint URL of the root Sourcegraph server (used for fetching remote repo data)" default:""`

	IsRoot bool `long:"fed.is-root" description:"configures this server to be a root server (if set, --fed.root-url setting is discarded)"`
}

// Config is the currently active federation config (as set by the CLI
// flags).
var Config Flags

// RootURL returns the parsed federation root URL. If the current
// process is configured as a federation root, it returns nil.
func (f *Flags) RootURL() *url.URL {
	if f.IsRoot || f.RootURLStr == "" {
		return nil
	}
	u, err := url.Parse(f.RootURLStr)
	if err != nil {
		log.Panicf("Failed to parse federation root URL: %s.", err)
	}
	if !u.IsAbs() {
		log.Panicf("Federation root URL must be absolute: %q.", u)
	}
	return u
}

// RootGRPCEndpoint returns the parsed federation root endpoint gRPC URL.
// If the current process is configured as a federation root, it returns nil.
// If the flag is not set, it will perform discovery on the RootURL and try to
// obtain the gRPC endpoint URL.
func (f *Flags) RootGRPCEndpoint() (*url.URL, error) {
	if f.IsRoot {
		return nil, nil
	}

	if f.RootGRPCURLStr == "" {
		var err error
		f.RootGRPCURLStr, err = discoverFedRootGRPCUrl(f.RootURLStr)
		if err != nil {
			return nil, err
		}
	}

	u, err := url.Parse(f.RootGRPCURLStr)
	return u, err
}

// AllowsClientRegistration is whether the server configured by f
// allows API clients to be registered with it (with client IDs,
// secrets, etc.).
func (f *Flags) AllowsClientRegistration() bool {
	return f.IsRoot
}

func discoverFedRootGRPCUrl(rootURLStr string) (string, error) {
	if rootURLStr == "" {
		return "", errors.New("cannot perform root gRPC endpoint discovery: fed.root-url is not set.")
	}

	ctx := context.Background()
	info, err := discover.Repo(ctx, rootURLStr)
	if err != nil {
		return "", fmt.Errorf("failed to discover federation root gRPC endpoint: %v", err)
	}

	ctx, err = info.NewContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to parse root gRPC endpoint: %v", err)
	}

	return sourcegraph.GRPCEndpoint(ctx).String(), err
}

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("Federation", "Federation", &Config)
	})
}
