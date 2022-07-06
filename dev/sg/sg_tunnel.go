package main

import (
	"fmt"
	"net/url"
	"os/exec"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var tunnelCommand = &cli.Command{
	Name:     "tunnel",
	Usage:    "Setup a tunnel to forward requests from the internet to your local instance",
	Category: CategoryDev,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "url",
			Aliases: []string{"u"},
			Usage:   "URL to forward request to. (default: https://sourcegraph.test:3443)",
			Value:   "https://sourcegraph.test:3443",
		},
	},
	Action: func(cmd *cli.Context) error {
		np, err := exec.LookPath("ngrok")
		if err != nil {
			std.Out.WriteLine(output.Styled(output.StyleFailure, "'sg tunnel' requires ngrok to be installed"))
			return err
		}

		u, err := url.Parse(cmd.String("url"))
		if err != nil {
			return err
		}

		args := []string{
			"http",
			fmt.Sprintf("--host-header=%s", u.Host),
			cmd.String("url"),
		}

		c := exec.Command(np, args...)
		err = run.InteractiveInRoot(c)
		if err != nil {
			return err
		}
		return nil
	},
}
