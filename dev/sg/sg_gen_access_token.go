package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	genGatewayAccessTokenCommand = &cli.Command{
		Name: "gen_gateway_access_token",
		Usage: "Generate a dotcom user gateway access token from a personal access token. " +
			"Useful for testing Cody Gateway locally.",
		Action:   genGatewayAccessTokenExec,
		Category: category.Util,
	}
)

func genGatewayAccessTokenExec(c *cli.Context) error {
	accessToken, err := accesstoken.GenerateDotcomUserGatewayAccessToken(c.Args().Get(0))
	if err != nil {
		return errors.Newf("failed to generate gateway access token: %s", err)
	}
	std.Out.Write(accessToken)
	return nil
}
