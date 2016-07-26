package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	c, err := cli.CLI.AddCommand("meta", "server meta-information", "", &metaCmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("config",
		"server config",
		"The `sgx meta config` command displays server config information.",
		&metaConfigCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type metaCmd struct{}

func (c *metaCmd) Execute(args []string) error { return nil }

type metaConfigCmd struct{}

func (c *metaConfigCmd) Execute(args []string) error {
	log.Println("#", endpoint.URLOrDefault())
	cl := cliClient
	config, err := cl.Meta.Config(cliContext, &pbtypes.Void{})
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// getRemoteAppURL returns the parsed AppURL of the remote Sourcegraph
// server configured in ctx.
func getRemoteAppURL(ctx context.Context) (*url.URL, error) {
	conf, err := cliClient.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return nil, err
	}
	if conf.AppURL == "" {
		return nil, fmt.Errorf("server with gRPC endpoint %s has no AppURL", sourcegraph.GRPCEndpoint(ctx))
	}
	return url.Parse(conf.AppURL)
}
