package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	startFlagSet       = flag.NewFlagSet("sg start", flag.ExitOnError)
	debugStartServices = startFlagSet.String("debug", "", "Comma separated list of services to set at debug log level.")
	infoStartServices  = startFlagSet.String("info", "", "Comma separated list of services to set at info log level.")
	warnStartServices  = startFlagSet.String("warn", "", "Comma separated list of services to set at warn log level.")
	errorStartServices = startFlagSet.String("error", "", "Comma separated list of services to set at error log level.")
	critStartServices  = startFlagSet.String("crit", "", "Comma separated list of services to set at crit log level.")

	startCommand = &ffcli.Command{
		Name:       "start",
		ShortUsage: "sg start [commandset]",
		ShortHelp:  "🌟Starts the given commandset. Without a commandset it starts the default Sourcegraph dev environment.",
		LongHelp:   constructStartCmdLongHelp(),

		FlagSet: startFlagSet,
		Exec:    startExec,
	}

	// run-set is the deprecated older version of `start`
	runSetFlagSet = flag.NewFlagSet("sg run-set", flag.ExitOnError)
	runSetCommand = &ffcli.Command{
		Name:       "run-set",
		ShortUsage: "sg run-set <commandset>",
		ShortHelp:  "DEPRECATED. Use 'sg start' instead. Run the given commandset.",
		FlagSet:    runSetFlagSet,
		Exec:       runSetExec,
	}
)

func constructStartCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, `Runs the given commandset.

If no commandset is specified, it starts the commandset with the name 'default'.

Use this to start your Sourcegraph environment!
`)

	// Attempt to parse config to list available commands, but don't fail on
	// error, because we should never error when the user wants --help output.
	_, _ = parseConf(*configFlag, *overwriteConfigFlag)

	if globalConf != nil {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE COMMANDSETS IN %s%s%s\n", output.StyleBold, *configFlag, output.StyleReset)

		var names []string
		for name := range globalConf.Commandsets {
			switch name {
			case "enterprise-codeintel":
				names = append(names, fmt.Sprintf("  %s 🧠", name))
			case "batches":
				names = append(names, fmt.Sprintf("  %s 🦡", name))
			default:
				names = append(names, fmt.Sprintf("  %s", name))
			}
		}
		sort.Strings(names)
		fmt.Fprint(&out, strings.Join(names, "\n"))
	}

	return out.String()
}

func startExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		stdout.Out.WriteLine(errLine)
		os.Exit(1)
	}

	if len(args) > 2 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	if len(args) != 1 {
		if globalConf.DefaultCommandset != "" {
			args = append(args, globalConf.DefaultCommandset)
		} else {
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: No commandset specified and no 'defaultCommandset' specified in sg.config.yaml\n"))
			return flag.ErrHelp
		}
	}

	set, ok := globalConf.Commandsets[args[0]]
	if !ok {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: commandset %q not found :(", args[0]))
		return flag.ErrHelp
	}

	// If the commandset requires the dev-private repository to be cloned, we
	// check that it's at the right location here.
	if set.RequiresDevPrivate {
		repoRoot, err := root.RepositoryRoot()
		if err != nil {
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "Failed to determine repository root location: %s", err))
			os.Exit(1)
		}

		devPrivatePath := filepath.Join(repoRoot, "..", "dev-private")
		exists, err := pathExists(devPrivatePath)
		if err != nil {
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "Failed to check whether dev-private repository exists: %s", err))
			os.Exit(1)
		}
		if !exists {
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: dev-private repository not found!"))
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "It's expected to exist at: %s", devPrivatePath))
			stdout.Out.WriteLine(output.Line("", output.StyleWarning, "If you're not a Sourcegraph employee you probably want to run: sg start oss"))
			stdout.Out.WriteLine(output.Line("", output.StyleWarning, "If you're a Sourcegraph employee, see the documentation for how to clone it: https://docs.sourcegraph.com/dev/getting-started/quickstart_2_clone_repository"))

			stdout.Out.Write("")
			overwritePath := filepath.Join(repoRoot, "sg.config.overwrite.yaml")
			stdout.Out.WriteLine(output.Linef("", output.StylePending, "If you know what you're doing and want disable the check, add the following to %s:", overwritePath))
			stdout.Out.Write("")
			stdout.Out.Write(fmt.Sprintf(`  commandsets:
    %s:
      requiresDevPrivate: false
`, set.Name))
			stdout.Out.Write("")

			os.Exit(1)
		}
	}

	var checks []run.Check
	for _, name := range set.Checks {
		check, ok := globalConf.Checks[name]
		if !ok {
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "WARNING: check %s not found in config", name))
			continue
		}
		checks = append(checks, check)
	}

	ok, err := run.Checks(ctx, globalConf.Env, checks...)
	if err != nil {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: checks could not be run: %s", err))
	}

	if !ok {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: checks did not pass, aborting start of commandset %s", set.Name))
		return nil
	}

	cmds := make([]run.Command, 0, len(set.Commands))
	for _, name := range set.Commands {
		cmd, ok := globalConf.Commands[name]
		if !ok {
			return errors.Errorf("command %q not found in commandset %q", name, args[0])
		}

		cmds = append(cmds, cmd)
	}

	if len(cmds) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "WARNING: no commands to run"))
	}

	levelOverrides := logLevelOverrides()
	for _, cmd := range cmds {
		enrichWithLogLevels(&cmd, levelOverrides)
	}

	env := globalConf.Env
	for k, v := range set.Env {
		env[k] = v
	}

	return run.Commands(ctx, env, *verboseFlag, cmds...)
}

// logLevelOverrides builds a map of commands -> log level that should be overridden in the environment.
func logLevelOverrides() map[string]string {
	levelServices := make(map[string][]string)
	levelServices["debug"] = parseCsv(*debugStartServices)
	levelServices["info"] = parseCsv(*infoStartServices)
	levelServices["warn"] = parseCsv(*warnStartServices)
	levelServices["error"] = parseCsv(*errorStartServices)
	levelServices["crit"] = parseCsv(*critStartServices)

	overrides := make(map[string]string)
	for level, services := range levelServices {
		for _, service := range services {
			overrides[service] = level
		}
	}

	return overrides
}

// enrichWithLogLevels will add any logger level overrides to a given command if they have been specified.
func enrichWithLogLevels(cmd *run.Command, overrides map[string]string) {
	logLevelVariable := "SRC_LOG_LEVEL"

	if level, ok := overrides[cmd.Name]; ok {
		stdout.Out.WriteLine(output.Linef("", output.StylePending, "Setting log level: %s for command %s.", level, cmd.Name))
		if cmd.Env == nil {
			cmd.Env = make(map[string]string, 1)
			cmd.Env[logLevelVariable] = level
		}
		cmd.Env[logLevelVariable] = level
	}
}

// parseCsv takes an input comma seperated string and returns a list of tokens each trimmed for whitespace
func parseCsv(input string) []string {
	tokens := strings.Split(input, ",")
	results := make([]string, 0, len(tokens))
	for _, token := range tokens {
		results = append(results, strings.TrimSpace(token))
	}
	return results
}

var deprecationStyle = output.CombineStyles(output.Fg256Color(255), output.Bg256Color(124))

func runSetExec(ctx context.Context, args []string) error {
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, " _______________________________________________________________________ "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "/         `sg run-set` is deprecated - use `sg start` instead!          \\"))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "!                                                                       !"))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "!         Run `sg start -help` for usage information.                   !"))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "\\_______________________________________________________________________/"))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               !  !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               !  !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               L_ !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                              / _)!                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                             / /__L                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                       _____/ (____)                                     "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                              (____)                                     "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                       _____  (____)                                     "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                            \\_(____)                                     "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               !  !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               !  !                                      "))
	stdout.Out.WriteLine(output.Linef("", deprecationStyle, "                               \\__/                                      "))
	return startExec(ctx, args)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
