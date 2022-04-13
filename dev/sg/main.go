package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func main() {
	if err := loadSecrets(); err != nil {
		writeWarningLinef("failed to open secrets: %s", err)
	}
	ctx := secrets.WithContext(context.Background(), secretsStore)

	if err := sg.RunContext(ctx, os.Args); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

const (
	defaultConfigFile          = "sg.config.yaml"
	defaultConfigOverwriteFile = "sg.config.overwrite.yaml"
	defaultSecretsFile         = "sg.secrets.json"
)

var (
	BuildCommit = "dev"

	// globalConf is the global config. If a command needs to access it, it *must* call
	// `parseConf` before.
	globalConf *Config

	// secretsStore is instantiated when sg gets run.
	secretsStore *secrets.Store

	// Note that these values are only available after the main sg CLI app has been run.
	configFlag          string
	overwriteConfigFlag string

	// Global verbose mode
	verbose bool

	// postInitHooks is useful for doing anything that requires flags to be set beforehand,
	// e.g. generating help text based on parsed config
	postInitHooks []cli.ActionFunc
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
			Usage:       "load sg configuration from `file`",
			EnvVars:     []string{"SG_CONFIG"},
			TakesFile:   true,
			Value:       defaultConfigFile,
			Destination: &configFlag,
		},
		&cli.StringFlag{
			Name:        "overwrite",
			Usage:       "load sg configuration from `file` that is gitignored and can be used to, for example, add credentials",
			EnvVars:     []string{"SG_OVERWRITE"},
			TakesFile:   true,
			Value:       defaultConfigOverwriteFile,
			Destination: &overwriteConfigFlag,
		},
		&cli.BoolFlag{
			Name:    "skip-auto-update",
			Usage:   "prevent sg from automatically updating itself",
			EnvVars: []string{"SG_SKIP_AUTO_UPDATE"},
			Value:   BuildCommit == "dev", // Default to skip in dev, otherwise don't
		},
	},
	Before: func(cmd *cli.Context) error {
		if verbose {
			stdout.Out.SetVerbose()
		}

		// We always try to set this, since we
		// often want to watch files, start commands, etc...
		if err := setMaxOpenFiles(); err != nil {
			writeWarningLinef("Failed to set max open files: %s", err)
		}

		if cmd.Args().First() != "update" {
			// If we're not running "sg update ...", we want to check the version first
			err := checkSgVersionAndUpdate(cmd.Context, cmd.Bool("skip-auto-update"))
			if err != nil {
				writeWarningLinef("update check: %s", err)
				// Do not exit here, so we don't break user flow when they want to
				// run `sg` but updating fails
			}
		}

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

func loadSecrets() error {
	homePath, err := root.GetSGHomePath()
	if err != nil {
		return err
	}
	fp := filepath.Join(homePath, defaultSecretsFile)
	secretsStore, err = secrets.LoadFile(fp)
	return err
}

// parseConf parses the config file and the optional overwrite file.
// Iear the conf has already been parsed it's a noop.
func parseConf(confFile, overwriteFile string) (bool, output.FancyLine) {
	if globalConf != nil {
		return true, output.FancyLine{}
	}

	// Try to determine root of repository, so we can look for config there
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return false, output.Linef("", output.StyleWarning, "Failed to determine repository root location: %s", err)
	}

	// If the configFlag/overwriteConfigFlag flags have their default value, we
	// take the value as relative to the root of the repository.
	if confFile == defaultConfigFile {
		confFile = filepath.Join(repoRoot, confFile)
	}

	if overwriteFile == defaultConfigOverwriteFile {
		overwriteFile = filepath.Join(repoRoot, overwriteFile)
	}

	globalConf, err = ParseConfigFile(confFile)
	if err != nil {
		return false, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as configuration file:%s\n%s", output.StyleBold, confFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
	}

	if ok, _ := fileExists(overwriteFile); ok {
		overwriteConf, err := ParseConfigFile(overwriteFile)
		if err != nil {
			return false, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as overwrites configuration file:%s\n%s", output.StyleBold, overwriteFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
		}
		globalConf.Merge(overwriteConf)
	}

	return true, output.FancyLine{}
}
