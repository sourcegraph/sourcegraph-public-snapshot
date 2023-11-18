package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/run"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func main() {
	sanitycheck.Pass()
	cfg := &config.Config{}
	cfg.Load()

	env.Lock()

	logging.Init() //nolint:staticcheck // Deprecated, but logs unmigrated to sourcegraph/log look really bad without this.
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()

	logger := log.Scoped("executor")

	runner := &util.RealCmdRunner{}

	makeActionHandler := func(handler func(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, config *config.Config) error) func(*cli.Context) error {
		return func(ctx *cli.Context) error {
			return handler(ctx, runner, logger, cfg)
		}
	}

	app := &cli.App{
		Version: version.Version(),
		// TODO: More info, link to docs, some inline documentation etc.
		Description:    "The Sourcegraph untrusted jobs runner. See https://docs.sourcegraph.com/admin/executors to learn more about setup, how it works and how to configure features that depend on it.",
		Name:           "executor",
		Usage:          "The Sourcegraph untrusted jobs runner.",
		DefaultCommand: "run",
		CommandNotFound: func(ctx *cli.Context, s string) {
			fmt.Printf("Unknown command %s. Use %s help to learn more.\n", s, ctx.App.HelpName)
			os.Exit(1)
		},
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Runs the executor. Connects to the job queue and processes jobs.",
				// Also show the env vars supported.
				CustomHelpTemplate: cli.CommandHelpTemplate + env.HelpString(),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "verify",
						Usage:    "Run validation checks to make sure the environment is set up correctly before starting to dequeue jobs.",
						Required: false,
					},
				},
				Action: makeActionHandler(run.Run),
			},
			{
				Name:   "validate",
				Usage:  "Validate the environment is set up correctly.",
				Action: makeActionHandler(run.Validate),
			},
			{
				Name:  "install",
				Usage: "Install components required to run executors.",
				Subcommands: []*cli.Command{
					{
						Name:  "ignite",
						Usage: "Installs ignite required for executor VMs. Firecracker only.",
						Flags: []cli.Flag{
							&cli.PathFlag{
								Name:        "bin-dir",
								Usage:       "Set the bin directory used to install ignite to. Must be in the PATH.",
								DefaultText: "/usr/local/bin",
								Required:    false,
							},
						},
						Action: makeActionHandler(run.InstallIgnite),
					},
					{
						Name:   "image",
						Usage:  "Ensures required runtime images are pulled and imported properly. Firecracker only.",
						Action: makeActionHandler(run.InstallImage),
					},
					{
						Name:   "cni",
						Usage:  "Installs CNI plugins required for executor VMs. Firecracker only.",
						Action: makeActionHandler(run.InstallCNI),
					},
					{
						Name:  "src-cli",
						Usage: "Installs src-cli at a supported version.",
						Flags: []cli.Flag{
							&cli.PathFlag{
								Name:        "bin-dir",
								Usage:       "Set the bin directory used to install src-cli to. Must be in the PATH.",
								DefaultText: "/usr/local/bin",
								Required:    false,
							},
						},
						Action: makeActionHandler(run.InstallSrc),
					},
					{
						Name:  "iptables-rules",
						Usage: "Installs iptables rules required for maximum isolation of executor VMs. Firecracker only.",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:     "recreate-chain",
								Usage:    "Force recreate the CNI_ADMIN iptables chain.",
								Required: false,
							},
						},
						Action: makeActionHandler(run.InstallIPTablesRules),
					},
					{
						Name:   "all",
						Usage:  "Runs all installers listed above.",
						Action: makeActionHandler(run.InstallAll),
					},
				},
			},
			{
				Name:  "test-vm",
				Usage: "Spawns a test VM with the parameters configured through the environment and prints a command to connect to it.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "repo",
						Usage: "Provide a repo name to clone the repository at HEAD into the VM. Optional.",

						Required: false,
					},
					&cli.StringFlag{
						Name:  "revision",
						Usage: "Provide a revision to check out when using --repo. Required when using --repo.",

						Required: false,
					},
					&cli.BoolFlag{
						Name:     "name-only",
						Usage:    "Only print the vm name on stdout. Can be used to call ignite attach programmatically.",
						Required: false,
					},
				},
				Action: makeActionHandler(run.TestVM),
			},
		},
	}

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
