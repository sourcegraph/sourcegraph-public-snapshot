package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func init() {
	postInitHooks = append(postInitHooks, func(cmd *cli.Context) {
		// Create 'sg start' help text after flag (and config) initialization
		startCommand.Description = constructStartCmdLongHelp()
	})
}

var (
	debugStartServices cli.StringSlice
	infoStartServices  cli.StringSlice
	warnStartServices  cli.StringSlice
	errorStartServices cli.StringSlice
	critStartServices  cli.StringSlice

	startCommand = &cli.Command{
		Name:      "start",
		ArgsUsage: "[commandset]",
		Usage:     "🌟 Starts the given commandset. Without a commandset it starts the default Sourcegraph dev environment",
		Category:  CategoryDev,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "Services to set at debug log level.",
				Destination: &debugStartServices,
			},
			&cli.StringSliceFlag{
				Name:        "info",
				Aliases:     []string{"i"},
				Usage:       "Services to set at info log level.",
				Destination: &infoStartServices,
			},
			&cli.StringSliceFlag{
				Name:        "warn",
				Aliases:     []string{"w"},
				Usage:       "Services to set at warn log level.",
				Destination: &warnStartServices,
			},
			&cli.StringSliceFlag{
				Name:        "error",
				Aliases:     []string{"e"},
				Usage:       "Services to set at info error level.",
				Destination: &errorStartServices,
			},
			&cli.StringSliceFlag{
				Name:        "crit",
				Aliases:     []string{"c"},
				Usage:       "Services to set at info crit level.",
				Destination: &critStartServices,
			},
		},
		BashComplete: completeOptions(func() (options []string) {
			config, _ := sgconf.Get(configFile, configOverwriteFile)
			if config == nil {
				return
			}
			for name := range config.Commandsets {
				options = append(options, name)
			}
			return
		}),
		Action: execAdapter(startExec),
	}
)

func constructStartCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, `Runs the given commandset.

If no commandset is specified, it starts the commandset with the name 'default'.

Use this to start your Sourcegraph environment!
`)

	config, err := sgconf.Get(configFile, configOverwriteFile)
	if err != nil {
		out.Write([]byte("\n"))
		output.NewOutput(&out, output.OutputOpts{}).WriteLine(newWarningLinef(err.Error()))
		return out.String()
	}

	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE COMMANDSETS IN %s%s%s:\n", output.StyleBold, configFile, output.StyleReset)

	var names []string
	for name := range config.Commandsets {
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

	return out.String()
}

func startExec(ctx context.Context, args []string) error {
	config, err := sgconf.Get(configFile, configOverwriteFile)
	if err != nil {
		writeWarningLinef(err.Error())
		os.Exit(1)
	}

	if len(args) > 2 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	if len(args) != 1 {
		if config.DefaultCommandset != "" {
			args = append(args, config.DefaultCommandset)
		} else {
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: No commandset specified and no 'defaultCommandset' specified in sg.config.yaml\n"))
			return flag.ErrHelp
		}
	}

	set, ok := config.Commandsets[args[0]]
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
			stdout.Out.WriteLine(output.Line("", output.StyleWarning, "If you're not a Sourcegraph teammate you probably want to run: sg start oss"))
			stdout.Out.WriteLine(output.Line("", output.StyleWarning, "If you're a Sourcegraph teammate, see the documentation for how to clone it: https://docs.sourcegraph.com/dev/getting-started/quickstart_2_clone_repository"))

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

	return startCommandSet(ctx, set, config)
}

func startCommandSet(ctx context.Context, set *sgconf.Commandset, conf *sgconf.Config) error {
	if err := runChecksWithName(ctx, set.Checks); err != nil {
		return err
	}

	cmds := make([]run.Command, 0, len(set.Commands))
	for _, name := range set.Commands {
		cmd, ok := conf.Commands[name]
		if !ok {
			return errors.Errorf("command %q not found in commandset %q", name, set.Name)
		}

		cmds = append(cmds, cmd)
	}

	if len(cmds) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "WARNING: no commands to run"))
		return nil
	}

	levelOverrides := logLevelOverrides()
	for _, cmd := range cmds {
		enrichWithLogLevels(&cmd, levelOverrides)
	}

	env := conf.Env
	for k, v := range set.Env {
		env[k] = v
	}

	return run.Commands(ctx, env, verbose, cmds...)
}

// logLevelOverrides builds a map of commands -> log level that should be overridden in the environment.
func logLevelOverrides() map[string]string {
	levelServices := make(map[string][]string)
	levelServices["debug"] = parseCsvs(debugStartServices.Value())
	levelServices["info"] = parseCsvs(infoStartServices.Value())
	levelServices["warn"] = parseCsvs(warnStartServices.Value())
	levelServices["error"] = parseCsvs(errorStartServices.Value())
	levelServices["crit"] = parseCsvs(critStartServices.Value())

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

func parseCsvs(inputs []string) []string {
	var allValues []string
	for _, i := range inputs {
		values := parseCsv(i)
		allValues = append(allValues, values...)
	}
	return allValues
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
