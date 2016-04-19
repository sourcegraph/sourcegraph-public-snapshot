package cli

import (
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/client"
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
	ctx := client.Endpoint.NewContext(parent)
	ctx = client.WithCLICredentials(ctx)
	return ctx
}
