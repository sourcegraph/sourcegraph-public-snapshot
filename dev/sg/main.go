package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/Masterminds/semver"
	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/migration"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/squash"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var BuildCommit string = "dev"

var out *output.Output = stdout.Out

var (
	runFlagSet = flag.NewFlagSet("sg run", flag.ExitOnError)
	runCommand = &ffcli.Command{
		Name:       "run",
		ShortUsage: "sg run <command>...",
		ShortHelp:  "Run the given commands.",
		LongHelp:   constructRunCmdLongHelp(),
		FlagSet:    runFlagSet,
		Exec:       runExec,
	}

	runSetFlagSet = flag.NewFlagSet("sg run-set", flag.ExitOnError)
	runSetCommand = &ffcli.Command{
		Name:       "run-set",
		ShortUsage: "sg run-set <commandset>",
		ShortHelp:  "DEPRECATED. Use 'sg start' instead. Run the given commandset.",
		FlagSet:    runSetFlagSet,
		Exec:       runSetExec,
	}

	startFlagSet       = flag.NewFlagSet("sg start", flag.ExitOnError)
	debugStartServices = startFlagSet.String("debug", "", "Comma separated list of services to set at debug log level.")
	infoStartServices  = startFlagSet.String("info", "", "Comma separated list of services to set at info log level.")
	warnStartServices  = startFlagSet.String("warn", "", "Comma separated list of services to set at warn log level.")
	errorStartServices = startFlagSet.String("error", "", "Comma separated list of services to set at error log level.")
	critStartServices  = startFlagSet.String("crit", "", "Comma separated list of services to set at crit log level.")
	startCommand       = &ffcli.Command{
		Name:       "start",
		ShortUsage: "sg start [commandset]",
		ShortHelp:  "🌟Starts the given commandset. Without a commandset it starts the default Sourcegraph dev environment.",
		LongHelp:   constructStartCmdLongHelp(),

		FlagSet: startFlagSet,
		Exec:    startExec,
	}

	testFlagSet = flag.NewFlagSet("sg test", flag.ExitOnError)
	testCommand = &ffcli.Command{
		Name:       "test",
		ShortUsage: "sg test <testsuite>",
		ShortHelp:  "Run the given test suite.",
		LongHelp:   "Run the given test suite.",
		FlagSet:    testFlagSet,
		Exec:       testExec,
	}

	doctorFlagSet = flag.NewFlagSet("sg doctor", flag.ExitOnError)
	doctorCommand = &ffcli.Command{
		Name:       "doctor",
		ShortUsage: "sg doctor",
		ShortHelp:  "Run the checks defined in the sg config file.",
		LongHelp: `Run the checks defined in the sg config file to make sure your system is healthy.

See the "checks:" in the configuration file.`,
		FlagSet: doctorFlagSet,
		Exec:    doctorExec,
	}

	liveFlagSet = flag.NewFlagSet("sg live", flag.ExitOnError)
	liveCommand = &ffcli.Command{
		Name:       "live",
		ShortUsage: "sg live <environment>",
		ShortHelp:  "Reports which version of Sourcegraph is currently live in the given environment",
		LongHelp:   constructLiveCmdLongHelp(),
		FlagSet:    liveFlagSet,
		Exec:       liveExec,
	}

	migrationAddFlagSet          = flag.NewFlagSet("sg migration add", flag.ExitOnError)
	migrationAddDatabaseNameFlag = migrationAddFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance.")
	migrationAddCommand          = &ffcli.Command{
		Name:       "add",
		ShortUsage: fmt.Sprintf("sg migration add [-db=%s] <name>", db.DefaultDatabase.Name),
		ShortHelp:  "Add a new migration file",
		FlagSet:    migrationAddFlagSet,
		Exec:       migrationAddExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationUpFlagSet          = flag.NewFlagSet("sg migration up", flag.ExitOnError)
	migrationUpDatabaseNameFlag = migrationUpFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance.")
	migrationUpNFlag            = migrationUpFlagSet.Int("n", 1, "How many migrations to apply.")
	migrationUpCommand          = &ffcli.Command{
		Name:       "up",
		ShortUsage: fmt.Sprintf("sg migration up [-db=%s] [-n]", db.DefaultDatabase.Name),
		ShortHelp:  "Run up migration files",
		FlagSet:    migrationUpFlagSet,
		Exec:       migrationUpExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationDownFlagSet          = flag.NewFlagSet("sg migration down", flag.ExitOnError)
	migrationDownDatabaseNameFlag = migrationDownFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance.")
	migrationDownNFlag            = migrationDownFlagSet.Int("n", 1, "How many migrations to apply.")
	migrationDownCommand          = &ffcli.Command{
		Name:       "down",
		ShortUsage: fmt.Sprintf("sg migration down [-db=%s] [-n=1]", db.DefaultDatabase.Name),
		ShortHelp:  "Run down migration files",
		FlagSet:    migrationDownFlagSet,
		Exec:       migrationDownExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationSquashFlagSet          = flag.NewFlagSet("sg migration squash", flag.ExitOnError)
	migrationSquashDatabaseNameFlag = migrationSquashFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance")
	migrationSquashCommand          = &ffcli.Command{
		Name:       "squash",
		ShortUsage: fmt.Sprintf("sg migration squash [-db=%s] <current-release>", db.DefaultDatabase.Name),
		ShortHelp:  "Collapse migration files from historic releases together",
		FlagSet:    migrationSquashFlagSet,
		Exec:       migrationSquashExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationFixupFlagSet          = flag.NewFlagSet("sg migration fixup", flag.ExitOnError)
	migrationFixupDatabaseNameFlag = migrationFixupFlagSet.String("db", "all", "The target database instance (or 'all' for all databases)")
	migrationFixupMainNameFlag     = migrationFixupFlagSet.String("main", "main", "The branch/revision to compare with")
	migrationFixupRunFlag          = migrationFixupFlagSet.Bool("run", true, "Run the migrations in your local database")
	migrationFixupCommand          = &ffcli.Command{
		Name:       "fixup",
		ShortUsage: fmt.Sprintf("sg migration fixup [-db=%s] [-main=%s] [-run=true]", "all", "main"),
		ShortHelp:  "Find and fix any conflicting migration names from rebasing on main. Also properly migrates your local database",
		FlagSet:    migrationFixupFlagSet,
		Exec:       migrationFixupExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationFlagSet = flag.NewFlagSet("sg migration", flag.ExitOnError)
	migrationCommand = &ffcli.Command{
		Name:       "migration",
		ShortUsage: "sg migration <command>",
		ShortHelp:  "Modifies and runs database migrations",
		FlagSet:    migrationFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			migrationAddCommand,
			migrationUpCommand,
			migrationDownCommand,
			migrationSquashCommand,
			migrationFixupCommand,
		},
	}

	rfcFlagSet = flag.NewFlagSet("sg rfc", flag.ExitOnError)
	rfcCommand = &ffcli.Command{
		Name:       "rfc",
		ShortUsage: "sg rfc [list|search|open]",
		ShortHelp:  "Run the given RFC command to manage RFCs.",
		LongHelp:   `List, search and open Sourcegraph RFCs`,
		FlagSet:    rfcFlagSet,
		Exec:       rfcExec,
	}

	funkyLogoFlagSet = flag.NewFlagSet("sg logo", flag.ExitOnError)
	funkLogoCommand  = &ffcli.Command{
		Name:       "logo",
		ShortUsage: "sg logo",
		ShortHelp:  "Print the sg logo",
		FlagSet:    funkyLogoFlagSet,
		Exec:       logoExec,
	}
)

const (
	defaultConfigFile          = "sg.config.yaml"
	defaultConfigOverwriteFile = "sg.config.overwrite.yaml"
)

var (
	rootFlagSet         = flag.NewFlagSet("sg", flag.ExitOnError)
	configFlag          = rootFlagSet.String("config", defaultConfigFile, "configuration file")
	overwriteConfigFlag = rootFlagSet.String("overwrite", defaultConfigOverwriteFile, "configuration overwrites file that is gitignored and can be used to, for example, add credentials")

	rootCommand = &ffcli.Command{
		ShortUsage: "sg [flags] <subcommand>",
		FlagSet:    rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			runCommand,
			runSetCommand,
			startCommand,
			testCommand,
			doctorCommand,
			liveCommand,
			migrationCommand,
			rfcCommand,
			funkLogoCommand,
		},
	}
)

func setMaxOpenFiles() error {
	const maxOpenFiles = 10000

	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return err
	}

	if rLimit.Cur < maxOpenFiles {
		rLimit.Cur = maxOpenFiles

		// This may not succeed, see https://github.com/golang/go/issues/30401
		return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	}

	return nil
}

func checkSgVersion() {
	_, err := root.RepositoryRoot()
	if err != nil {
		// Ignore the error, because we only want to check the version if we're
		// in sourcegraph/sourcegraph
		return
	}

	if BuildCommit == "dev" {
		// If `sg` was built with a dirty `./dev/sg` directory it's a dev build
		// and we don't need to display this message.
		return
	}

	out, err := run.GitCmd("rev-list", fmt.Sprintf("%s..HEAD", BuildCommit), "./dev/sg")
	if err != nil {
		fmt.Printf("error getting new commits in ./dev/sg: %s\n", err)
		os.Exit(1)
	}

	out = strings.TrimSpace(out)
	if out != "" {
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "--------------------------------------------------------------------------"))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "HEY! New version of sg available. Run `./dev/sg/install.sh` to install it."))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "--------------------------------------------------------------------------"))
	}
}

func main() {
	if err := rootCommand.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	checkSgVersion()

	// We always try to set this, since we often want to watch files, start commands, etc.
	if err := setMaxOpenFiles(); err != nil {
		fmt.Printf("failed to set max open files: %s\n", err)
		os.Exit(1)
	}

	if err := rootCommand.Run(context.Background()); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

// globalConf is the global config. If a command needs to access it, it *must* call
// `parseConf` before.
var globalConf *Config

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
		return false, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as configuration file:%s\n%s\n", output.StyleBold, confFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
	}

	if ok, _ := fileExists(overwriteFile); ok {
		overwriteConf, err := ParseConfigFile(overwriteFile)
		if err != nil {
			return false, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as overwrites configuration file:%s\n%s\n", output.StyleBold, overwriteFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
		}
		globalConf.Merge(overwriteConf)
	}

	return true, output.FancyLine{}
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

// enrichWithLogLevels will add any logger level overrides to a given command if they have been specified.
func enrichWithLogLevels(cmd *run.Command, overrides map[string]string) {
	logLevelVariable := "SRC_LOG_LEVEL"

	if level, ok := overrides[cmd.Name]; ok {
		out.WriteLine(output.Linef("", output.StylePending, "Setting log level: %s for command %s.", level, cmd.Name))
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

func testExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		out.WriteLine(errLine)
		os.Exit(1)
	}

	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No test suite specified\n"))
		return flag.ErrHelp
	}

	cmd, ok := globalConf.Tests[args[0]]
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: test suite %q not found :(\n", args[0]))
		return flag.ErrHelp
	}

	return run.Test(ctx, cmd, args[1:], globalConf.Env)
}

func startExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		out.WriteLine(errLine)
		os.Exit(1)
	}

	if len(args) > 2 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	if len(args) == 0 {
		args = append(args, "default")
	}

	set, ok := globalConf.Commandsets[args[0]]
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: commandset %q not found :(\n", args[0]))
		return flag.ErrHelp
	}

	var checks []run.Check
	for _, name := range set.Checks {
		check, ok := globalConf.Checks[name]
		if !ok {
			out.WriteLine(output.Linef("", output.StyleWarning, "WARNING: check %s not found in config\n", name))
			continue
		}
		checks = append(checks, check)
	}

	ok, err := run.Checks(ctx, globalConf.Env, checks...)
	if err != nil {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: checks could not be run: %s\n", err))
	}

	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: checks did not pass, aborting start of commandset %s\n", set.Name))
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

	levelOverrides := logLevelOverrides()
	for _, cmd := range cmds {
		enrichWithLogLevels(&cmd, levelOverrides)
	}

	env := globalConf.Env
	for k, v := range set.Env {
		env[k] = v
	}

	return run.Commands(ctx, env, cmds...)
}

func runExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		out.WriteLine(errLine)
		os.Exit(1)
	}

	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No command specified\n"))
		return flag.ErrHelp
	}

	var cmds []run.Command
	for _, arg := range args {
		cmd, ok := globalConf.Commands[arg]
		if !ok {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: command %q not found :(\n", arg))
			return flag.ErrHelp
		}
		cmds = append(cmds, cmd)
	}

	return run.Commands(ctx, globalConf.Env, cmds...)
}

func doctorExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		out.WriteLine(errLine)
		os.Exit(1)
	}

	var checks []run.Check
	for _, c := range globalConf.Checks {
		checks = append(checks, c)
	}
	_, err := run.Checks(ctx, globalConf.Env, checks...)
	return err
}

func liveExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No environment specified\n"))
		return flag.ErrHelp
	}

	if len(args) != 1 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	e, ok := getEnvironment(args[0])
	if !ok {
		if customURL, err := url.Parse(args[0]); err == nil {
			e = environment{Name: customURL.Host, URL: customURL.String()}
		} else {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: environment %q not found, or is not a valid URL :(\n", args[0]))
			return flag.ErrHelp
		}
	}

	return printDeployedVersion(e)
}

func migrationAddExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No migration name specified\n"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	var (
		databaseName  = *migrationAddDatabaseNameFlag
		migrationName = args[0]
		database, ok  = db.DatabaseByName(databaseName)
	)
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", databaseName))
		return flag.ErrHelp
	}

	upFile, downFile, err := migration.RunAdd(database, migrationName)
	if err != nil {
		return err
	}

	block := out.Block(output.Linef("", output.StyleBold, "Migration files created"))
	block.Writef("Up migration: %s", upFile)
	block.Writef("Down migration: %s", downFile)
	block.Close()

	return nil
}

func migrationUpExec(ctx context.Context, args []string) error {
	if len(args) != 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	var (
		databaseName = *migrationUpDatabaseNameFlag
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", databaseName))
		return flag.ErrHelp
	}

	var n *int
	migrationUpFlagSet.Visit(func(f *flag.Flag) {
		if f.Name == "n" {
			n = migrationUpNFlag
		}
	})

	// Only pass the value of n here if the user actually set it
	// We have to do the dance above because the flags package
	// requires you to define a default value for each flag.
	return migration.RunUp(database, n)
}

func migrationDownExec(ctx context.Context, args []string) error {
	if len(args) != 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	var (
		databaseName = *migrationDownDatabaseNameFlag
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", databaseName))
		return flag.ErrHelp
	}

	return migration.RunDown(database, migrationDownNFlag)
}

// minimumMigrationSquashDistance is the minimum number of releases a migration is guaranteed to exist
// as a non-squashed file.
//
// A squash distance of 1 will allow one minor downgrade.
// A squash distance of 2 will allow two minor downgrades.
// etc
const minimumMigrationSquashDistance = 2

func migrationSquashExec(ctx context.Context, args []string) (err error) {
	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No current-version specified\n"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	var (
		databaseName  = *migrationSquashDatabaseNameFlag
		migrationName = args[0]
		database, ok  = db.DatabaseByName(databaseName)
	)
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", databaseName))
		return flag.ErrHelp
	}

	currentVersion, err := semver.NewVersion(migrationName)
	if err != nil {
		return err
	}

	// Get the last migration that existed in the version _before_ `minimumMigrationSquashDistance` releases ago
	commit := fmt.Sprintf("v%d.%d.0", currentVersion.Major(), currentVersion.Minor()-minimumMigrationSquashDistance-1)
	out.Writef("Squashing migration files defined up through %s", commit)

	return squash.Run(database, commit)
}

func constructRunCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "  Runs the given command. If given a whitespace-separated list of commands it runs the set of commands.\n")

	// Attempt to parse config to list available commands, but don't fail on
	// error, because we should never error when the user wants --help output.
	_, _ = parseConf(*configFlag, *overwriteConfigFlag)

	if globalConf != nil {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE COMMANDS IN %s%s%s\n", output.StyleBold, *configFlag, output.StyleReset)

		for name := range globalConf.Commands {
			fmt.Fprintf(&out, "  %s\n", name)
		}
	}

	return out.String()
}

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
		fmt.Fprintf(&out, strings.Join(names, "\n"))
	}

	return out.String()
}

func migrationFixupExec(ctx context.Context, args []string) (err error) {
	if len(args) != 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	branchName := *migrationFixupMainNameFlag
	if branchName == "" {
		branchName = "main"
	}

	databaseName := *migrationFixupDatabaseNameFlag
	if databaseName == "all" {
		for _, databaseName := range db.DatabaseNames() {
			database, _ := db.DatabaseByName(databaseName)
			if err := migration.RunFixup(database, branchName, *migrationFixupRunFlag); err != nil {
				return err
			}
		}

		return nil
	} else {
		database, ok := db.DatabaseByName(databaseName)
		if !ok {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", databaseName))
			return flag.ErrHelp
		}

		return migration.RunFixup(database, branchName, *migrationFixupRunFlag)
	}
}

func printTestUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg %s <test suite>\n", c.Name)

	// Attempt to parse config so we can list test suites, but don't fail on
	// error, because we should never error when the user wants --help output.
	_, _ = parseConf(*configFlag, *overwriteConfigFlag)

	if globalConf != nil {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE TESTSUITES IN %s%s%s\n", output.StyleBold, *configFlag, output.StyleReset)

		for name := range globalConf.Tests {
			fmt.Fprintf(&out, "  %s\n", name)
		}
	}

	return out.String()
}

func printRunSetUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "DEPRECATED! 'sg run-set' has been deprecated. Please use 'sg start' instead.\n")

	return out.String()
}

func printStartUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg %s [commandset]\n", c.Name)

	// Attempt to parse config so we can list available sets, but don't fail on
	// error, because we should never error when the user wants --help output.
	_, _ = parseConf(*configFlag, *overwriteConfigFlag)
	if globalConf != nil {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE COMMANDSETS IN %s%s%s\n", output.StyleBold, *configFlag, output.StyleReset)

		for name := range globalConf.Commandsets {
			fmt.Fprintf(&out, "  %s\n", name)
		}
	}

	return out.String()
}

func constructLiveCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "Prints the Sourcegraph version deployed to the given environment.")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE PRESET ENVIRONMENTS\n")

	for _, name := range environmentNames() {
		fmt.Fprintf(&out, "  %s\n", name)
	}

	return out.String()
}

func constructMigrationSubcmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "AVAILABLE DATABASES\n")
	var names []string
	for _, name := range db.DatabaseNames() {
		names = append(names, fmt.Sprintf("  %s", name))
	}
	fmt.Fprintf(&out, strings.Join(names, "\n"))

	return out.String()
}

func printRFCUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg %s <command>\n", c.Name)
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "COMMANDS:\n")
	fmt.Fprintf(&out, "    list - list all RFCs\n")
	fmt.Fprintf(&out, "    search <query> - search for RFCs matching the query\n")
	fmt.Fprintf(&out, "    open <number> - Open the specified RFC\n")

	return out.String()
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

var styleOrange = output.Fg256Color(202)

func printLogo(out io.Writer) {
	fmt.Fprintf(out, "%s", output.StyleLogo)
	fmt.Fprintln(out, `          _____                    _____`)
	fmt.Fprintln(out, `         /\    \                  /\    \`)
	fmt.Fprintf(out, `        /%s::%s\    \                /%s::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `       /%s::::%s\    \              /%s::::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `      /%s::::::%s\    \            /%s::::::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `     /%s:::%s/\%s:::%s\    \          /%s:::%s/\%s:::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `    /%s:::%s/__\%s:::%s\    \        /%s:::%s/  \%s:::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `    \%s:::%s\   \%s:::%s\    \      /%s:::%s/    \%s:::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `  ___\%s:::%s\   \%s:::%s\    \    /%s:::%s/    / \%s:::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, ` /\   \%s:::%s\   \%s:::%s\    \  /%s:::%s/    /   \%s:::%s\ ___\`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `/%s::%s\   \%s:::%s\   \%s:::%s\____\/%s:::%s/____/  ___\%s:::%s|    |`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `\%s:::%s\   \%s:::%s\   \%s::%s/    /\%s:::%s\    \ /\  /%s:::%s|____|`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, ` \%s:::%s\   \%s:::%s\   \/____/  \%s:::%s\    /%s::%s\ \%s::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `  \%s:::%s\   \%s:::%s\    \       \%s:::%s\   \%s:::%s\ \/____/`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `   \%s:::%s\   \%s:::%s\____\       \%s:::%s\   \%s:::%s\____\`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `    \%s:::%s\  /%s:::%s/    /        \%s:::%s\  /%s:::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `     \%s:::%s\/%s:::%s/    /          \%s:::%s\/%s:::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `      \%s::::::%s/    /            \%s::::::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `       \%s::::%s/    /              \%s::::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `        \%s::%s/    /                \%s::%s/____/`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `         \/____/`)
	fmt.Fprintf(out, "%s", output.StyleReset)
}

func logoExec(ctx context.Context, args []string) error {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	randoColor := func() output.Style { return output.Fg256Color(r1.Intn(256)) }

	var (
		color1a = randoColor()
		color1b = randoColor()
		color1c = randoColor()
		color2  = output.StyleLogo
	)

	times := 20
	for i := 0; i < times; i++ {
		const linesPrinted = 23

		stdout.Out.Writef("%s", color2)
		stdout.Out.Write(`          _____                    _____`)
		stdout.Out.Write(`         /\    \                  /\    \`)
		stdout.Out.Writef(`        /%s::%s\    \                /%s::%s\    \`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`       /%s::::%s\    \              /%s::::%s\    \`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`      /%s::::::%s\    \            /%s::::::%s\    \`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`     /%s:::%s/\%s:::%s\    \          /%s:::%s/\%s:::%s\    \`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`    /%s:::%s/__\%s:::%s\    \        /%s:::%s/  \%s:::%s\    \`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`    \%s:::%s\   \%s:::%s\    \      /%s:::%s/    \%s:::%s\    \`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`  ___\%s:::%s\   \%s:::%s\    \    /%s:::%s/    / \%s:::%s\    \`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(` /\   \%s:::%s\   \%s:::%s\    \  /%s:::%s/    /   \%s:::%s\ ___\`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`/%s::%s\   \%s:::%s\   \%s:::%s\____\/%s:::%s/____/  ___\%s:::%s|    |`, color1a, color2, color1b, color2, color1c, color2, color1a, color2, color1b, color2)
		stdout.Out.Writef(`\%s:::%s\   \%s:::%s\   \%s::%s/    /\%s:::%s\    \ /\  /%s:::%s|____|`, color1a, color2, color1b, color2, color1c, color2, color1a, color2, color1b, color2)
		stdout.Out.Writef(` \%s:::%s\   \%s:::%s\   \/____/  \%s:::%s\    /%s::%s\ \%s::%s/    /`, color1a, color2, color1b, color2, color1c, color2, color1a, color2, color1b, color2)
		stdout.Out.Writef(`  \%s:::%s\   \%s:::%s\    \       \%s:::%s\   \%s:::%s\ \/____/`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`   \%s:::%s\   \%s:::%s\____\       \%s:::%s\   \%s:::%s\____\`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`    \%s:::%s\  /%s:::%s/    /        \%s:::%s\  /%s:::%s/    /`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`     \%s:::%s\/%s:::%s/    /          \%s:::%s\/%s:::%s/    /`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`      \%s::::::%s/    /            \%s::::::%s/    /`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`       \%s::::%s/    /              \%s::::%s/    /`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`        \%s::%s/    /                \%s::%s/____/`, color1a, color2, color1b, color2)
		stdout.Out.Write(`         \/____/`)
		stdout.Out.Writef("%s", output.StyleReset)

		time.Sleep(200 * time.Millisecond)

		color1a, color1b, color1c, color2 = randoColor(), color1a, color1b, color1c

		if i != times-1 {
			stdout.Out.MoveUpLines(linesPrinted)
		}
	}

	return nil
}
