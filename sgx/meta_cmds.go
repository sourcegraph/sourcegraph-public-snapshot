package sgx

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	c, err := cli.CLI.AddCommand("meta", "server meta-information", "", &metaCmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("status",
		"server status",
		"The `sgx meta status` command displays server status information.",
		&metaStatusCmd{},
	)
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

type metaStatusCmd struct{}

func (c *metaStatusCmd) Execute(args []string) error {
	cl := Client()
	status, err := cl.Meta.Status(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	fmt.Println(status.Info)
	return nil
}

type metaConfigCmd struct{}

func (c *metaConfigCmd) Execute(args []string) error {
	log.Println("#", Endpoints.Endpoint)
	cl := Client()
	config, err := cl.Meta.Config(cliCtx, &pbtypes.Void{})
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
	conf, err := Client().Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return nil, err
	}
	if conf.AppURL == "" {
		return nil, fmt.Errorf("server with gRPC endpoint %s has no AppURL", sourcegraph.GRPCEndpoint(ctx))
	}
	return url.Parse(conf.AppURL)
}
