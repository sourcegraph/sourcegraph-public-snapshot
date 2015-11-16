package fed

import (
	"log"
	"net/url"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Flags defines settings (in the form of CLI flags) for federation.
type Flags struct {
	// RootURLStr is the base URL of the root server in the Sourcegraph
	// network. It is used for OAuth2 user authentication, default repo
	// graph data lookups, etc.
	RootURLStr string `long:"fed.root-url" description:"base URL of the root Sourcegraph server (used for OAuth2 user auth, open-source repo data, etc.)" default:"https://sourcegraph.com"`

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

// AllowsClientRegistration is whether the server configured by f
// allows API clients to be registered with it (with client IDs,
// secrets, etc.).
func (f *Flags) AllowsClientRegistration() bool {
	return f.IsRoot
}

// NewRemoteContext returns a copy of context that accesses remote
// services on the federation root.
func (f *Flags) NewRemoteContext(ctx context.Context) context.Context {
	return NewRemoteContext(ctx, f.RootURL())
}

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("Federation", "Federation", &Config)
	})
}
