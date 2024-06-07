package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	codyGatewayCommand = &cli.Command{
		Name:     "cody-gateway",
		Usage:    "set of commands that are helpful for working with Cody Gateway locally",
		Category: category.Util,
		Subcommands: []*cli.Command{{
			Name:   "gen-token",
			Usage:  "generate a new token for use with Cody Gateway from a personal access token",
			Action: genGatewayAccessTokenExec,
		}},
	}
)

func genGatewayAccessTokenExec(c *cli.Context) error {
	// The provided token must be nonempty and start with sgp_
	accessToken, err := accesstoken.GenerateDotcomUserGatewayAccessToken(c.Args().Get(0))
	if err != nil {
		return errors.Newf("failed to generate gateway access token: %s", err)
	}
	std.Out.Write(accessToken)
	return nil
}
