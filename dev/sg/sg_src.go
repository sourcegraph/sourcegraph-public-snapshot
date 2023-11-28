package main

import (
	"context"
	"net/url"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type srcInstance struct {
	AccessToken string `json:"access_token"`
	Endpoint    string `json:"endpoint"`
}

type srcSecrets struct {
	Current   string                 `json:"current"`
	Instances map[string]srcInstance `json:"instances"`
}

var srcInstanceCommand = &cli.Command{
	Name:      "src-instance",
	UsageText: "sg src-instance [command]",
	Usage:     "Interact with Sourcegraph instances that 'sg src' will use",
	Category:  category.Dev,
	Subcommands: []*cli.Command{
		{
			Name:      "register",
			Usage:     "Register (or edit an existing) Sourcegraph instance to target with src-cli",
			UsageText: "sg src instance register [name] [endpoint]",
			Action: func(cmd *cli.Context) error {
				store, sc, err := getSrcInstances(cmd.Context)
				if err != nil {
					return errors.Wrap(err, "failed to read existing instances")
				}
				if cmd.Args().Len() != 2 {
					return errors.Newf("not enough arguments, want %d got %d", 2, cmd.Args().Len())
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

				accessToken, err := std.Out.PromptPasswordf(
					os.Stdin,
					`Please enter the access token for Sourcegraph instance named %s (%s):`,
					name,
					endpoint,
				)
				if err != nil {
					return errors.Wrapf(err, "failed to read access token")
				}

				sc.Instances[name] = srcInstance{
					Endpoint:    endpoint,
					AccessToken: accessToken,
				}
				if err := store.PutAndSave("src", sc); err != nil {
					return errors.Wrap(err, "failed to save instance")
				}
				std.Out.WriteSuccessf("src instance %s added", name)
				std.Out.WriteSuggestionf("Run 'sg src-instance use %s' to switch to that instance for 'sg src'", name)
				return nil
			},
		},
		{
			Name:  "use",
			Usage: "Set current src-cli instance to use with 'sg src'",
			Action: func(cmd *cli.Context) error {
				store, sc, err := getSrcInstances(cmd.Context)
				if err != nil {
					return err
				}
				name := cmd.Args().First()
				instance, ok := sc.Instances[name]
				if !ok {
					std.Out.WriteFailuref("Instance not found, register one with 'sg src-instance register'")
					return errors.New("instance not found")
				}
				sc.Current = name
				if err := store.PutAndSave("src", sc); err != nil {
					return errors.Wrap(err, "failed to save instance")
				}
				std.Out.WriteSuccessf("Switched to %s (%s)", name, instance.Endpoint)
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "List registered instances for src-cli",
			Action: func(cmd *cli.Context) error {
				_, sc, err := getSrcInstances(cmd.Context)
				if err != nil {
					return err
				}
				std.Out.WriteLine(output.Linef("", output.StyleReset, "| %-16s| %-32s|", "Name", "Endpoint"))
				for name, instance := range sc.Instances {
					if name == sc.Current {
						std.Out.WriteLine(output.Linef("", output.StyleSuccess, "| %-16s| %-32s|", name, instance.Endpoint))
					} else {
						std.Out.WriteLine(output.Linef("", output.StyleReset, "| %-16s| %-32s|", name, instance.Endpoint))
					}
				}
				return nil
			},
		},
	},
}

var srcCommand = &cli.Command{
	Name:      "src",
	UsageText: "sg src [src-cli args]\nsg src help # get src-cli help",
	Usage:     "Run src-cli on a given instance defined with 'sg src-instance'",
	Category:  category.Dev,
	Action: func(cmd *cli.Context) error {
		_, sc, err := getSrcInstances(cmd.Context)
		if err != nil {
			return err
		}
		instanceName := sc.Current
		if instanceName == "" {
			std.Out.WriteFailuref("Instance not found, register one with 'sg src-instance register'")
			return errors.New("set an instance with sg src-instance use [instance-name]")
		}
		instance, ok := sc.Instances[instanceName]
		if !ok {
			std.Out.WriteFailuref("Instance not found, register one with 'sg src-instance register'")
			return errors.New("instance not found")
		}

		c := exec.CommandContext(cmd.Context, "src", cmd.Args().Slice()...)
		c.Env = append(c.Environ(), "SRC_ACCESS_TOKEN="+instance.AccessToken, "SRC_ENDPOINT="+instance.Endpoint)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return c.Run()
	},
}

// getSrcInstances retrieves src instances configuration from the secrets store
func getSrcInstances(ctx context.Context) (*secrets.Store, *srcSecrets, error) {
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
