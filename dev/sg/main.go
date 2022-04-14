package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	if os.Args[len(os.Args)-1] == "--generate-bash-completion" {
		batchCompletionMode = true
	}

	if err := sg.RunContext(context.Background(), os.Args); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

var (
	BuildCommit = "dev"

	// configFile is the path to use with sgconf.Get - it must not be used before flag
	// initialization.
	configFile string
	// configOverwriteFile is the path to use with sgconf.Get - it must not be used before
	// flag initialization.
	configOverwriteFile string

	// Global verbose mode
	verbose bool

	// postInitHooks is useful for doing anything that requires flags to be set beforehand,
	// e.g. generating help text based on parsed config, and are called before any command
	// Action is executed. These should run quickly and must fail gracefully.
	//
	// Commands can register postInitHooks in an 'init()' function that appends to this
	// slice.
	postInitHooks []func(cmd *cli.Context)

	// batchCompletionMode determines if we are in bash completion mode. In this mode,
	// sg should respond quickly, so most setup tasks (e.g. postInitHooks) are skipped.
	//
	// Do not run complicated tasks, etc. in Before or After hooks when in this mode.
	batchCompletionMode bool
)

// sg is the main sg CLI application.
var sg = &cli.App{
	Usage:       "The Sourcegraph developer tool!",
	Description: "Learn more: https://docs.sourcegraph.com/dev/background-information/sg",
	Version:     BuildCommit,
	Compiled:    time.Now(),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "toggle verbose mode",
			Aliases:     []string{"v"},
			EnvVars:     []string{"SG_VERBOSE"},
			Value:       false,
			Destination: &verbose,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "load sg configuration from `file`",
			EnvVars:     []string{"SG_CONFIG"},
			TakesFile:   true,
			Value:       sgconf.DefaultFile,
			Destination: &configFile,
		},
		&cli.StringFlag{
			Name:        "overwrite",
			Aliases:     []string{"o"},
			Usage:       "load sg configuration from `file` that is gitignored and can be used to, for example, add credentials",
			EnvVars:     []string{"SG_OVERWRITE"},
			TakesFile:   true,
			Value:       sgconf.DefaultOverwriteFile,
			Destination: &configOverwriteFile,
		},
		&cli.BoolFlag{
			Name:    "skip-auto-update",
			Usage:   "prevent sg from automatically updating itself",
			EnvVars: []string{"SG_SKIP_AUTO_UPDATE"},
			Value:   BuildCommit == "dev", // Default to skip in dev, otherwise don't
		},
	},
	Before: func(cmd *cli.Context) error {
		if batchCompletionMode {
			// All other setup pertains to running commands - to keep completions fast,
			// we skip all other setup.
			return nil
		}

		if verbose {
			stdout.Out.SetVerbose()
		}

		// Validate configuration flags, which is required for sgconf.Get to work everywhere else.
		if configFile == "" {
			return errors.Newf("--config must not be empty")
		}
		if configOverwriteFile == "" {
			return errors.Newf("--overwrite must not be empty")
		}

		// Set up access to secrets
		secretsStore, err := loadSecrets()
		if err != nil {
			writeWarningLinef("failed to open secrets: %s", err)
		} else {
			cmd.Context = secrets.WithContext(cmd.Context, secretsStore)
		}

		// We always try to set this, since we often want to watch files, start commands, etc...
		if err := setMaxOpenFiles(); err != nil {
			writeWarningLinef("Failed to set max open files: %s", err)
		}

		// Check for updates, unless we are running update manually.
		if cmd.Args().First() != "update" {
			err := checkSgVersionAndUpdate(cmd.Context, cmd.Bool("skip-auto-update"))
			if err != nil {
				writeWarningLinef("update check: %s", err)
				// Do not exit here, so we don't break user flow when they want to
				// run `sg` but updating fails
			}
		}

		// Call registered hooks last
		for _, hook := range postInitHooks {
			hook(cmd)
		}

		return nil
	},
	Commands: []*cli.Command{
		// Common dev tasks
		startCommand,
		runCommand,
		ciCommand,
		testCommand,
		lintCommand,
		generateCommand,
		dbCommand,
		migrationCommand,

		// Dev environment
		doctorCommand,
		secretCommand,
		setupCommand,

		// Company
		teammateCommand,
		rfcCommand,
		liveCommand,
		opsCommand,
		auditCommand,

		// Util
		helpCommand,
		versionCommand,
		updateCommand,
		installCommand,
		funkyLogoCommand,
	},

	EnableBashCompletion:   true,
	UseShortOptionHandling: true,

	HideVersion:     true,
	HideHelpCommand: true,
}

func loadSecrets() (*secrets.Store, error) {
	homePath, err := root.GetSGHomePath()
	if err != nil {
		return nil, err
	}
	fp := filepath.Join(homePath, secrets.DefaultFile)
	return secrets.LoadFromFile(fp)
}
