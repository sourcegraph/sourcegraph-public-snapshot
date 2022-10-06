package main

import (
	"context"
	"os"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/urfave/cli/v2"
)

type srcInstance struct {
	AccessToken string `json:"access_token"`
	Endpoint    string `json:"endpoint"`
}

type srcSecrets struct {
	Instances map[string]srcInstance `json:"instances"`
}

var srcCommand = &cli.Command{
	Name:      "src",
	UsageText: "sg src [instance] [src-cli args]",
	Usage:     "Run src-cli on a given instance",
	Category:  CategoryCompany,
	Action: func(cmd *cli.Context) error {
		_, sc, err := getSrcSecret(cmd.Context, std.Out)
		if err != nil {
			return err
		}
		instance, ok := sc.Instances[cmd.Args().Get(0)]
		if !ok {
			std.Out.WriteFailuref("Instance not found, register one with 'sg src register-instance'")
			return errors.New("instance not found")
		}
		srcArgs := cmd.Args().Slice()[1:]
		c := usershell.Command(cmd.Context, append([]string{"src"}, srcArgs...)...)
		c = c.Env(map[string]string{
			"SRC_ACCESS_TOKEN": instance.AccessToken,
			"SRC_ENDPOINT":     instance.Endpoint,
		})
		return c.Run().Stream(os.Stdout)
	},
	Subcommands: []*cli.Command{
		{
			Name:      "instance",
			UsageText: "todo",
			Subcommands: []*cli.Command{
				{
					Name:    "register",
					Usage:   "",
					Aliases: []string{"update"},
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "name",
							Usage:    "Name for the new instance (ex: s2)",
							Required: true,
						},
						&cli.StringFlag{
							Name:     "endpoint",
							Usage:    "Endpoint for the new instance (ex: https://sourcegraph.sourcegraph.com)",
							Required: true,
						},
						&cli.StringFlag{
							Name:     "access-token",
							Usage:    "AccessToken for the new instance",
							Required: true,
						},
					},
					Action: func(cmd *cli.Context) error {
						store, sc, err := getSrcSecret(cmd.Context, std.Out)
						if err != nil {
							return err
						}
						sc.Instances[cmd.String("name")] = srcInstance{
							Endpoint:    cmd.String("endpoint"),
							AccessToken: cmd.String("access-token"),
						}
						if err := store.PutAndSave("src", sc); err != nil {
							return err
						}
						return nil
					},
				},
				{
					Name:  "list",
					Usage: "",
					Action: func(cmd *cli.Context) error {
						_, sc, err := getSrcSecret(cmd.Context, std.Out)
						if err != nil {
							return err
						}
						for name, instance := range sc.Instances {
							std.Out.WriteLine(output.Linef("", output.StyleBold, "|%-16s|%-32s|", "Name", "Endpoint"))
							std.Out.WriteLine(output.Linef("", output.StyleReset, "|%-16s|%-32s|", name, instance.Endpoint))
						}
						return nil
					},
				},
			},
		},
	},
}

func getSrcSecret(ctx context.Context, out *std.Output) (*secrets.Store, *srcSecrets, error) {
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	sc := srcSecrets{Instances: map[string]srcInstance{}}
	err = sec.Get("src", &sc)
	if err != nil && !errors.Is(err, secrets.ErrSecretNotFound) {
		return nil, nil, err
	}
	return sec, &sc, nil
}
