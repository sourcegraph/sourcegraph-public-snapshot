package cli

import (
	"log"
	"time"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

	"bytes"

	srclib "sourcegraph.com/sourcegraph/srclib/cli"
)

// globalOpt contains global options.
var globalOpt struct {
	Config     func(s string) error `long:"config" description:"INI config file" no-ini:"true" env:"SRC_CONFIG"`
	Verbose    bool                 `long:"verbose" short:"v" description:"show verbose output (same as --log-level=dbug)" env:"SRC_VERBOSE"`
	VerbosePkg string               `long:"verbose-pkg" description:"if set, only log output from specified package" env:"SRC_VERBOSE_PKG"`
	LogLevel   string               `long:"log-level" description:"upper log level to restrict log output to (dbug, dbug-dev, info, warn, error, crit)" default:"info" env:"SRC_LOG_LEVEL"`

	Trace          []string      `long:"log-trace" description:"comma separated list of trace logs to show. Options: all,gRPC,HTTP,build,github" default:"HTTP" env:"SRC_LOG_TRACE"`
	TraceThreshold time.Duration `long:"log-trace-threshold" description:"Show traces that take longer than this." env:"SRC_LOG_TRACE_THRESHOLD"`
}

func init() {
	cli.CLI.LongDescription = "src runs and manages a Sourcegraph instance."
	cli.CLI.AddGroup("Global options", "", &globalOpt)

	cli.CLI.InitFuncs = append(cli.CLI.InitFuncs, func() {
		srclib.GlobalOpt.Verbose = globalOpt.Verbose
	})
}

func init() {
	srclib.CacheLocalRepo = false
}

func Main() error {
	log.SetFlags(0)
	log.SetPrefix("")

	// Add our own custom help group to each command.
	if err := addHelpGroups(cli.CLI.Command); err != nil {
		return err
	}

	for _, f := range cli.PostInit {
		f()
	}

	globalOpt.Config = func(s string) error {
		return flags.NewIniParser(cli.CLI).ParseFile(s)
	}

	_, err := cli.CLI.Parse()
	printErrorHelp(err)
	return err
}

// addHelpGroups adds help groups to the given command and all of it's sub
// commands, recursively.
func addHelpGroups(cmd *flags.Command) error {
	// Determine whether or not we should register the default help group.
	register := true
	for _, name := range cli.CustomHelpCmds {
		if cmd.Name == name {
			register = false
			break
		}
	}

	if register {
		// Build the group.
		var help struct {
			ShowHelp func() error `short:"h" long:"help" description:"Show this help message"`
		}
		help.ShowHelp = func() error {
			var b bytes.Buffer
			cli.CLI.WriteHelp(&b)
			return &flags.Error{
				Type:    flags.ErrHelp,
				Message: b.String(),
			}
		}

		// Add the group to the command.
		_, err := cmd.AddGroup("Help Options", "", &help)
		if err != nil {
			return err
		}
	}

	// Do the same for each sub command, recursively.
	for _, subCmd := range cmd.Commands() {
		if err := addHelpGroups(subCmd); err != nil {
			return err
		}
	}
	return nil
}
