package main

import (
	"context"
	"net/url"
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
					Name:      "register",
					Usage:     "Register (or edit an existing) Sourcegraph instance to target with src-cli",
					UsageText: "sg src instance register [name] [endpoint] [access_token]",
					Action: func(cmd *cli.Context) error {
						store, sc, err := getSrcSecret(cmd.Context, std.Out)
						if err != nil {
							return errors.Wrap(err, "failed to read existing instances")
						}
						if cmd.Args().Len() < 3 {
							return errors.Newf("not enough arguments, want %d got %d", 3, cmd.Args().Len())
						}

						name := cmd.Args().First()
						endpoint := cmd.Args().Slice()[1]
						endpointUrl, err := url.Parse(endpoint)
						if err != nil {
							return errors.Wrapf(err, "cannot parse [endpoint]")
						}
						if endpointUrl.Scheme != "http" && endpointUrl.Scheme != "https" {
							return errors.New("cannot parse [endpoint], scheme must be http or https")
						}

						accessToken := cmd.Args().Slice()[2]

						sc.Instances[name] = srcInstance{
							Endpoint:    endpoint,
							AccessToken: accessToken,
						}
						if err := store.PutAndSave("src", sc); err != nil {
							return errors.Wrap(err, "failed to save instance")
						}
						return nil
					},
				},
				{
					Name:  "list",
					Usage: "List registered instances for src-cli",
					Action: func(cmd *cli.Context) error {
						_, sc, err := getSrcSecret(cmd.Context, std.Out)
						if err != nil {
							return err
						}
						std.Out.WriteLine(output.Linef("", output.StyleBold, "| %-16s| %-32s|", "Name", "Endpoint"))
						for name, instance := range sc.Instances {
							std.Out.WriteLine(output.Linef("", output.StyleReset, "| %-16s| %-32s|", name, instance.Endpoint))
						}
						return nil
					},
				},
			},
		},
	},
}

// getSrcSecret retrieves src-cli secrets from the context secrets store
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
