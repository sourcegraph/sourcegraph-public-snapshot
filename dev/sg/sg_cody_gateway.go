package main

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var codyGatewayCommand = &cli.Command{
	Name:     "cody-gateway",
	Usage:    "set of commands that are helpful for working with Cody Gateway locally",
	Category: category.Util,
	Subcommands: []*cli.Command{{
		Name:        "gen-token",
		Usage:       "generate a new token for use with Cody Gateway",
		UsageText:   "sg cody-gateway gen-token [sg token]",
		Description: "generates a token with the given `sg token` - which should be prefixed with `sgp_`",
		Action:      genGatewayAccessTokenExec,
	}},
}

func genGatewayAccessTokenExec(c *cli.Context) error {
	out := std.NewOutput(os.Stderr, false)
	if c.NArg() == 0 {
		out.WriteWarningf("The first argument should be a Sourcegraph private access token - see --help for more information")
		return errors.New("missing private access token argument")
	}

	out.WriteNoticef("Generating new gateway access token ...")
	privateToken := c.Args().Get(0)
	if !strings.HasPrefix(privateToken, "sgp_") {
		out.WriteWarningf("Token must be prefixed with \"sgp_\"")
		return errors.New("invalid token: not prefixed with sgp_")
	}
	accessToken, err := accesstoken.GenerateDotcomUserGatewayAccessToken(privateToken)
	if err != nil {
		return errors.Newf("failed to generate gateway access token: %s", err)
	}
	std.Out.Write(accessToken)
	return nil
}
