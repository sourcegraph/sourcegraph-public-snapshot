package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/background"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	// Do not add initialization here, do all setup in sg.Before - this is a necessary
	// workaround because we don't have control over the bash completion flag, which is
	// part of urfave/cli internals.
	if os.Args[len(os.Args)-1] == "--generate-bash-completion" {
		bashCompletionsMode = true
	}

	if err := sg.RunContext(context.Background(), os.Args); err != nil {
		// We want to prefer an already-initialized std.Out no matter what happens,
		// because that can be configured (e.g. with '--disable-output-detection'). Only
		// if something went horribly wrong and std.Out is not yet initialized should we
		// attempt an initialization here.
		if std.Out == nil {
			std.Out = std.NewOutput(os.Stdout, false)
		}
		std.Out.WriteFailuref(err.Error())
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

	// bashCompletionsMode determines if we are in bash completion mode. In this mode,
	// sg should respond quickly, so most setup tasks (e.g. postInitHooks) are skipped.
	//
	// Do not run complicated tasks, etc. in Before or After hooks when in this mode.
	bashCompletionsMode bool
)

const sgBugReportTemplate = "https://github.com/sourcegraph/sourcegraph/issues/new?template=sg_bug.md"

// sg is the main sg CLI application.
//
//go:generate go run . help -full -output ./doc/dev/background-information/sg/reference.md
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
			Aliases:     []string{"c"},
			EnvVars:     []string{"SG_CONFIG"},
			TakesFile:   true,
			Value:       sgconf.DefaultFile,
			Destination: &configFile,
		},
		&cli.StringFlag{
			Name:        "overwrite",
			Usage:       "load sg configuration from `file` that is gitignored and can be used to, for example, add credentials",
			Aliases:     []string{"o"},
			EnvVars:     []string{"SG_OVERWRITE"},
			TakesFile:   true,
			Value:       sgconf.DefaultOverwriteFile,
			Destination: &configOverwriteFile,
		},
		&cli.BoolFlag{
			Name:    "skip-auto-update",
			Usage:   "prevent sg from automatically updating itself",
			EnvVars: []string{"SG_SKIP_AUTO_UPDATE"},
			Value:   BuildCommit == "dev", // Default to skip in dev
		},
		&cli.BoolFlag{
			Name:    "disable-analytics",
			Usage:   "disable event logging (logged to '~/.sourcegraph/events')",
			EnvVars: []string{"SG_DISABLE_ANALYTICS"},
			Value:   BuildCommit == "dev", // Default to skip in dev
		},
		&cli.BoolFlag{
			Name:        "disable-output-detection",
			Usage:       "use fixed output configuration instead of detecting terminal capabilities",
			EnvVars:     []string{"SG_DISABLE_OUTPUT_DETECTION"},
			Destination: &std.DisableOutputDetection,
		},
	},
	Before: func(cmd *cli.Context) (err error) {
		if bashCompletionsMode {
			// All other setup pertains to running commands - to keep completions fast,
			// we skip all other setup.
			return nil
		}

		var (
			start            = time.Now()
			disableAnalytics = cmd.Bool("disable-analytics")
		)

		// Let sg components register pre-interrupt hooks
		interrupt.Listen()

		// Configure global output
		std.Out = std.NewOutput(cmd.App.Writer, verbose)

		// Initialize context
		cmd.Context = background.Context(cmd.Context)
		cmd.Context, err = usershell.Context(cmd.Context)
		if err != nil {
			std.Out.WriteWarningf("Unable to infer user shell context: " + err.Error())
		}

		// Set up analytics and hooks for each command.
		if !disableAnalytics {
			cmd.Context = analytics.WithContext(cmd.Context, cmd.App.Version)
			addAnalyticsHooks(start, []string{"sg"}, cmd.App.Commands)

			// Lots of setup happens in Before - we want to make sure anything that
			// happens here is tracked. We set this up here after setting up output and
			// some initial safe setup.
			defer func() {
				if p := recover(); p != nil {
					std.Out.WriteWarningf("Encountered panic - please open an issue with the command output:\n\t%s",
						sgBugReportTemplate)
					message := fmt.Sprintf("%v:\n%s", p, getRelevantStack())
					err = cli.NewExitError(message, 1)

					event := analytics.LogEvent(cmd.Context, "sg_before", nil, start, "panic")
					event.Properties["error_details"] = err.Error()
					analytics.Persist(cmd.Context, "sg", cmd.FlagNames())
				}
			}()
		}

		// Configure logger, for commands that use components that use loggers
		os.Setenv("SRC_DEVELOPMENT", "true")
		os.Setenv("SRC_LOG_FORMAT", "console")
		liblog := log.Init(log.Resource{Name: "sg"})
		interrupt.Register(func() { _ = liblog.Sync() })

		// Add autosuggestion hooks to commands with subcommands but no action
		addSuggestionHooks(cmd.App.Commands)

		// Add feedback subcommand to all commands and subcommands
		addFeedbackFlags(cmd.App.Commands)

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
			std.Out.WriteWarningf("failed to open secrets: %s", err)
		} else {
			cmd.Context = secrets.WithContext(cmd.Context, secretsStore)
		}

		// We always try to set this, since we often want to watch files, start commands, etc...
		if err := setMaxOpenFiles(); err != nil {
			std.Out.WriteWarningf("Failed to set max open files: %s", err)
		}

		// Check for updates, unless we are running update manually.
		if cmd.Args().First() != "update" {
			background.Run(cmd.Context, func(ctx context.Context, out *std.Output) {
				err := checkSgVersionAndUpdate(ctx, out, cmd.Bool("skip-auto-update"))
				if err != nil {
					out.WriteWarningf("update check: %s", err)
				}
			}, verbose)
		}

		// Call registered hooks last
		for _, hook := range postInitHooks {
			hook(cmd)
		}

		return nil
	},
	After: func(cmd *cli.Context) error {
		if !bashCompletionsMode {
			// Wait for background jobs to finish up, iff not in autocomplete mode
			background.Wait(cmd.Context, std.Out)
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
		insightsCommand,

		// Dev environment
		doctorCommand,
		secretCommand,
		setupCommand,

		// Company
		teammateCommand,
		rfcCommand,
		adrCommand,
		liveCommand,
		opsCommand,
		auditCommand,
		analyticsCommand,

		// Util
		helpCommand,
		feedbackCommand,
		versionCommand,
		updateCommand,
		installCommand,
		funkyLogoCommand,
	},
	ExitErrHandler: func(cmd *cli.Context, err error) {
		if err == nil {
			return
		}

		// Show help text only
		if errors.Is(err, flag.ErrHelp) {
			cli.ShowSubcommandHelpAndExit(cmd, 1)
		}

		// Render error
		errMsg := err.Error()
		if errMsg != "" {
			std.Out.WriteFailuref(errMsg)
		}

		// Determine exit code
		if exitErr, ok := err.(cli.ExitCoder); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	},

	CommandNotFound: suggestCommands,

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
