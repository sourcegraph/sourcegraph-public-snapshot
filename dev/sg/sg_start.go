package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/conc/pool"
	sgrun "github.com/sourcegraph/run"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/exit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func init() {
	postInitHooks = append(postInitHooks,
		func(cmd *cli.Context) {
			// Create 'sg start' help text after flag (and config) initialization
			startCommand.Description = constructStartCmdLongHelp()
		},
		func(cmd *cli.Context) {
			ctx, cancel := context.WithCancel(cmd.Context)
			interrupt.Register(func() {
				cancel()
				// TODO wait for stuff properly.
				time.Sleep(1 * time.Second)
			})
			cmd.Context = ctx
		},
	)
}

const devPrivateDefaultBranch = "master"

var (
	debugStartServices cli.StringSlice
	infoStartServices  cli.StringSlice
	warnStartServices  cli.StringSlice
	errorStartServices cli.StringSlice
	critStartServices  cli.StringSlice
	exceptServices     cli.StringSlice
	onlyServices       cli.StringSlice

	startCommand = &cli.Command{
		Name:      "start",
		ArgsUsage: "[commandset]",
		Usage:     "🌟 Starts the given commandset. Without a commandset it starts the default Sourcegraph dev environment",
		UsageText: `
# Run default environment, Sourcegraph enterprise:
sg start

# List available environments (defined under 'commandSets' in 'sg.config.yaml'):
sg start -help

# Run the enterprise environment with code-intel enabled:
sg start enterprise-codeintel

# Run the environment for Batch Changes development:
sg start batches

# Override the logger levels for specific services
sg start --debug=gitserver --error=enterprise-worker,enterprise-frontend enterprise

# View configuration for a commandset
sg start -describe single-program
`,
		Category: category.Dev,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "describe",
				Usage: "Print details about the selected commandset",
			},
			&cli.BoolFlag{
				Name:  "sgtail",
				Usage: "Connects to running sgtail instance",
			},

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
			&cli.StringSliceFlag{
				Name:        "except",
				Usage:       "List of services of the specified command set to NOT start",
				Destination: &exceptServices,
			},
			&cli.StringSliceFlag{
				Name:        "only",
				Usage:       "List of services of the specified command set to start. Commands NOT in this list will NOT be started.",
				Destination: &onlyServices,
			},
		},
		BashComplete: completions.CompleteOptions(func() (options []string) {
			config, _ := getConfig()
			if config == nil {
				return
			}
			for name := range config.Commandsets {
				options = append(options, name)
			}
			return
		}),
		Action: startExec,
	}
)

func constructStartCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, `Use this to start your Sourcegraph environment!`)

	config, err := getConfig()
	if err != nil {
		out.Write([]byte("\n"))
		std.NewOutput(&out, false).WriteWarningf(err.Error())
		return out.String()
	}

	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "Available comamndsets in `%s`:\n", configFile)

	var names []string
	for name := range config.Commandsets {
		switch name {
		case "enterprise-codeintel":
			names = append(names, fmt.Sprintf("%s 🧠", name))
		case "batches":
			names = append(names, fmt.Sprintf("%s 🦡", name))
		default:
			names = append(names, name)
		}
	}
	sort.Strings(names)
	fmt.Fprint(&out, "\n* "+strings.Join(names, "\n* "))

	return out.String()
}

var sgOnce sync.Once

func startExec(ctx *cli.Context) error {
	config, err := getConfig()
	if err != nil {
		return err
	}

	args := ctx.Args().Slice()
	if len(args) > 2 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	if len(args) != 1 {
		if config.DefaultCommandset != "" {
			args = append(args, config.DefaultCommandset)
		} else {
			std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: No commandset specified and no 'defaultCommandset' specified in sg.config.yaml\n"))
			return flag.ErrHelp
		}
	}

	pid, exists, err := run.PidExistsWithArgs(os.Args[1:])
	if err != nil {
		std.Out.WriteAlertf("Could not check if 'sg %s' is already running with the same arguments. Process: %d", strings.Join(os.Args[1:], " "), pid)
		return errors.Wrap(err, "Failed to check if sg is already running with the same arguments or not.")
	}
	if exists {
		std.Out.WriteAlertf("Found 'sg %s' already running with the same arguments. Process: %d", strings.Join(os.Args[1:], " "), pid)
		return errors.New("no concurrent sg start with same arguments allowed")
	}

	if ctx.Bool("sgtail") {
		if err := run.OpenUnixSocket(); err != nil {
			return errors.Wrapf(err, "Did you forget to run sgtail first?")
		}
	}

	commandset := args[0]
	set, ok := config.Commandsets[commandset]
	if !ok {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: commandset %q not found :(", commandset))
		return flag.ErrHelp
	}

	if ctx.Bool("describe") {
		out, err := yaml.Marshal(set)
		if err != nil {
			return err
		}

		return std.Out.WriteMarkdown(fmt.Sprintf("# %s\n\n```yaml\n%s\n```\n\n", commandset, string(out)))
	}

	// If the commandset requires the dev-private repository to be cloned, we
	// check that it's at the right location here.
	if set.RequiresDevPrivate && !NoDevPrivateCheck {
		repoRoot, err := root.RepositoryRoot()
		if err != nil {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "Failed to determine repository root location: %s", err))
			return exit.NewEmptyExitErr(1)
		}

		devPrivatePath := filepath.Join(repoRoot, "..", "dev-private")
		exists, err := pathExists(devPrivatePath)
		if err != nil {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "Failed to check whether dev-private repository exists: %s", err))
			return exit.NewEmptyExitErr(1)
		}
		if !exists {
			std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: dev-private repository not found!"))
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "It's expected to exist at: %s", devPrivatePath))
			std.Out.WriteLine(output.Styled(output.StyleWarning, "See the documentation for how to get set up: https://docs.sourcegraph.com/dev/setup/quickstart#run-sg-setup"))

			std.Out.Write("")
			overwritePath := filepath.Join(repoRoot, "sg.config.overwrite.yaml")
			std.Out.WriteLine(output.Styledf(output.StylePending, "If you know what you're doing and want disable the check, add the following to %s:", overwritePath))
			std.Out.Write("")
			std.Out.Write(fmt.Sprintf(`  commandsets:
    %s:
      requiresDevPrivate: false
`, set.Name))
			std.Out.Write("")

			return exit.NewEmptyExitErr(1)
		}

		// dev-private exists, let's see if there are any changes
		update := std.Out.Pending(output.Styled(output.StylePending, "Checking for dev-private changes..."))
		shouldUpdate, err := shouldUpdateDevPrivate(ctx.Context, devPrivatePath, devPrivateDefaultBranch)
		if shouldUpdate {
			update.WriteLine(output.Line(output.EmojiInfo, output.StyleSuggestion, "We found some changes in dev-private that you're missing out on! If you want the new changes, 'cd ../dev-private' and then do a 'git stash' and a 'git pull'!"))
		}
		if err != nil {
			update.Close()
			std.Out.WriteWarningf("WARNING: Encountered some trouble while checking if there are remote changes in dev-private!")
			std.Out.Write("")
			std.Out.Write(err.Error())
			std.Out.Write("")
		} else {
			update.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Done checking dev-private changes"))
		}
	}

	return startCommandSet(ctx.Context, set, config)
}

func shouldUpdateDevPrivate(ctx context.Context, path, branch string) (bool, error) {
	// git fetch so that we check whether there are any remote changes
	if err := sgrun.Bash(ctx, fmt.Sprintf("git fetch origin %s", branch)).Dir(path).Run().Wait(); err != nil {
		return false, err
	}
	// Now we check if there are any changes. If the output is empty, we're not missing out on anything.
	outputStr, err := sgrun.Bash(ctx, fmt.Sprintf("git diff --shortstat origin/%s", branch)).Dir(path).Run().String()
	if err != nil {
		return false, err
	}
	return len(outputStr) > 0, err

}

func startCommandSet(ctx context.Context, set *sgconf.Commandset, conf *sgconf.Config) error {
	if err := runChecksWithName(ctx, set.Checks); err != nil {
		return err
	}

	exceptList := exceptServices.Value()
	exceptSet := make(map[string]interface{}, len(exceptList))
	for _, svc := range exceptList {
		exceptSet[svc] = struct{}{}
	}

	onlyList := onlyServices.Value()
	onlySet := make(map[string]interface{}, len(onlyList))
	for _, svc := range onlyList {
		onlySet[svc] = struct{}{}
	}

	cmds := make([]run.Command, 0, len(set.Commands))
	for _, name := range set.Commands {
		cmd, ok := conf.Commands[name]
		if !ok {
			return errors.Errorf("command %q not found in commandset %q", name, set.Name)
		}

		if _, excluded := exceptSet[name]; excluded {
			std.Out.WriteLine(output.Styledf(output.StylePending, "Skipping command %s since it's in --except.", cmd.Name))
			continue
		}

		// No --only specified, just add command
		if len(onlySet) == 0 {
			cmds = append(cmds, cmd)
		} else {
			if _, inSet := onlySet[name]; inSet {
				cmds = append(cmds, cmd)
			} else {
				std.Out.WriteLine(output.Styledf(output.StylePending, "Skipping command %s since it's not included in --only.", cmd.Name))
			}
		}

	}

	bcmds := make([]run.BazelCommand, 0, len(set.BazelCommands))
	for _, name := range set.BazelCommands {
		bcmd, ok := conf.BazelCommands[name]
		if !ok {
			return errors.Errorf("command %q not found in commandset %q", name, set.Name)
		}

		bcmds = append(bcmds, bcmd)
	}
	if len(cmds) == 0 && len(bcmds) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "WARNING: no commands to run"))
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

	// First we build everything once, to ensure all binaries are present.
	if err := run.BazelBuild(ctx, bcmds...); err != nil {
		return err
	}

	p := pool.New().WithContext(ctx).WithCancelOnError()
	p.Go(func(ctx context.Context) error {
		return run.Commands(ctx, env, verbose, cmds...)
	})
	p.Go(func(ctx context.Context) error {
		return run.BazelCommands(ctx, env, verbose, bcmds...)
	})

	return p.Wait()
}

// logLevelOverrides builds a map of commands -> log level that should be overridden in the environment.
func logLevelOverrides() map[string]string {
	levelServices := make(map[string][]string)
	levelServices["debug"] = debugStartServices.Value()
	levelServices["info"] = infoStartServices.Value()
	levelServices["warn"] = warnStartServices.Value()
	levelServices["error"] = errorStartServices.Value()
	levelServices["crit"] = critStartServices.Value()

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
		std.Out.WriteLine(output.Styledf(output.StylePending, "Setting log level: %s for command %s.", level, cmd.Name))
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
