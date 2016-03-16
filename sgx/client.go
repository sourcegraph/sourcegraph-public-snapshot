package sgx

import (
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/client"
)

func init() {
	cli.CLI.InitFuncs = append(cli.CLI.InitFuncs, func() {
		// The "src version" command does not need a cli context at all.
		if cli.CLI.Active != nil && cli.CLI.Active.Name == "version" {
			return
		}
		client.Ctx = WithClientContext(context.Background())
	})
}

// WithClientContext returns a copy of parent with client endpoint and
// auth information added.
func WithClientContext(parent context.Context) context.Context {
	// The "src serve" command is the only non-client command; it
	// must not have credentials set (because it is not a client
	// command).
	if cli.CLI.Active != nil && cli.CLI.Active.Name == "serve" {
		client.Credentials.AuthFile = ""
	}
	ctx := WithClientContextUnauthed(parent)
	ctx, err := client.Credentials.WithCredentials(ctx)
	if err != nil {
		log.Fatalf("Error constructing API client credentials: %s.", err)
	}
	return ctx
}

// WithClientContextUnauthed returns a copy of parent with client endpoint added.
func WithClientContextUnauthed(parent context.Context) context.Context {
	return client.Endpoint.NewContext(parent)
}
