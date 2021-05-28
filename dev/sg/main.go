package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	runFlagSet = flag.NewFlagSet("sg run", flag.ExitOnError)
	runCommand = &ffcli.Command{
		Name:       "run",
		ShortUsage: "sg run <command>",
		ShortHelp:  "Run the given command.",
		FlagSet:    runFlagSet,
		Exec:       runExec,
		UsageFunc:  printRunUsage,
	}

	runSetFlagSet = flag.NewFlagSet("sg run-set", flag.ExitOnError)
	runSetCommand = &ffcli.Command{
		Name:       "run-set",
		ShortUsage: "sg run-set <commandset>",
		ShortHelp:  "Run the given command set.",
		FlagSet:    runSetFlagSet,
		Exec:       runSetExec,
		UsageFunc:  printRunSetUsage,
	}

	startFlagSet = flag.NewFlagSet("sg start", flag.ExitOnError)
	startCommand = &ffcli.Command{
		Name:       "start",
		ShortUsage: "sg start",
		ShortHelp:  "Runs the commandset with the name 'start'.",
		FlagSet:    startFlagSet,
		Exec:       startExec,
		UsageFunc:  printStartUsage,
	}

	testFlagSet = flag.NewFlagSet("sg test", flag.ExitOnError)
	testCommand = &ffcli.Command{
		Name:       "test",
		ShortUsage: "sg test <testsuite>",
		ShortHelp:  "Run the given test suite.",
		FlagSet:    testFlagSet,
		Exec:       testExec,
		UsageFunc:  printTestUsage,
	}

	doctorFlagSet = flag.NewFlagSet("sg doctor", flag.ExitOnError)
	doctorCommand = &ffcli.Command{
		Name:       "doctor",
		ShortUsage: "sg doctor",
		ShortHelp:  "Run the checks defined in the config file to make sure your system is healthy.",
		FlagSet:    doctorFlagSet,
		Exec:       doctorExec,
		UsageFunc:  printDoctorUsage,
	}

	liveFlagSet = flag.NewFlagSet("sg live", flag.ExitOnError)
	liveCommand = &ffcli.Command{
		Name:       "live",
		ShortUsage: "sg live <environment>",
		ShortHelp:  "Reports which version of Sourcegraph is currently live in the given environment",
		FlagSet:    liveFlagSet,
		Exec:       liveExec,
		UsageFunc:  printLiveUsage,
	}

	migrationAddFlagSet          = flag.NewFlagSet("sg migration add", flag.ExitOnError)
	migrationAddDatabaseNameFlag = migrationAddFlagSet.String("db", defaultDatabaseName, "The target database instance.")
	migrationAddCommand          = &ffcli.Command{
		Name:       "add",
		ShortUsage: fmt.Sprintf("sg migration add [-db=%s] <name>", defaultDatabaseName),
		ShortHelp:  "Add a new migration file",
		FlagSet:    migrationAddFlagSet,
		Exec:       migrationAddExec,
		UsageFunc:  printMigrationAddUsage,
	}

	migrationUpFlagSet          = flag.NewFlagSet("sg migration up", flag.ExitOnError)
	migrationUpDatabaseNameFlag = migrationUpFlagSet.String("db", defaultDatabaseName, "The target database instance.")
	migrationUpNFlag            = migrationUpFlagSet.Int("n", 1, "How many migrations to apply.")
	migrationUpCommand          = &ffcli.Command{
		Name:       "up",
		ShortUsage: fmt.Sprintf("sg migration up [-db=%s] [-n=1]", defaultDatabaseName),
		ShortHelp:  "Run up migration files",
		FlagSet:    migrationUpFlagSet,
		Exec:       migrationUpExec,
		UsageFunc:  printMigrationUpUsage,
	}

	migrationDownFlagSet          = flag.NewFlagSet("sg migration down", flag.ExitOnError)
	migrationDownDatabaseNameFlag = migrationDownFlagSet.String("db", defaultDatabaseName, "The target database instance.")
	migrationDownNFlag            = migrationDownFlagSet.Int("n", 1, "How many migrations to apply.")
	migrationDownCommand          = &ffcli.Command{
		Name:       "down",
		ShortUsage: fmt.Sprintf("sg migration down [-db=%s] [-n=1]", defaultDatabaseName),
		ShortHelp:  "Run down migration files",
		FlagSet:    migrationDownFlagSet,
		Exec:       migrationDownExec,
		UsageFunc:  printMigrationDownUsage,
	}

	migrationSquashFlagSet          = flag.NewFlagSet("sg migration squash", flag.ExitOnError)
	migrationSquashDatabaseNameFlag = migrationSquashFlagSet.String("db", defaultDatabaseName, "The target database instance")
	migrationSquashCommand          = &ffcli.Command{
		Name:       "squash",
		ShortUsage: fmt.Sprintf("sg migration squash [-db=%s] <current-release>", defaultDatabaseName),
		ShortHelp:  "Collapse migration files from historic releases together",
		FlagSet:    migrationSquashFlagSet,
		Exec:       migrationSquashExec,
		UsageFunc:  printMigrationSquashUsage,
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
		UsageFunc: printMigrationUsage,
		Subcommands: []*ffcli.Command{
			migrationAddCommand,
			migrationUpCommand,
			migrationDownCommand,
			migrationSquashCommand,
		},
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
		UsageFunc: func(c *ffcli.Command) string {
			var out strings.Builder

			printLogo(&out)

			fmt.Fprintf(&out, "USAGE\n")
			fmt.Fprintf(&out, "  sg <subcommand>\n")

			fmt.Fprintf(&out, "\n")
			fmt.Fprintf(&out, "AVAILABLE COMMANDS\n")
			for _, sub := range c.Subcommands {
				fmt.Fprintf(&out, "  %s\n", sub.Name)
			}

			fmt.Fprintf(&out, "\nRun 'sg <subcommand> -help' to get help output for each subcommand\n")

			return out.String()
		},
		Subcommands: []*ffcli.Command{
			runCommand,
			runSetCommand,
			startCommand,
			testCommand,
			doctorCommand,
			liveCommand,
			migrationCommand,
		},
	}
)

func main() {
	if err := rootCommand.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		out.WriteLine(errLine)
		os.Exit(1)
	}

	if err := rootCommand.Run(context.Background()); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

var conf *Config

// parseConf parses the config file and the optional overwrite file.
// If the conf
// has already been parsed it's a noop.
func parseConf(confFile, overwriteFile string) (bool, output.FancyLine) {
	if conf != nil {
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

	conf, err = ParseConfigFile(confFile)
	if err != nil {
		return false, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as configuration file:%s\n%s\n", output.StyleBold, confFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
	}

	if ok, _ := fileExists(overwriteFile); ok {
		overwriteConf, err := ParseConfigFile(overwriteFile)
		if err != nil {
			return false, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as overwrites configuration file:%s\n%s\n", output.StyleBold, overwriteFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
		}
		conf.Merge(overwriteConf)
	}

	return true, output.FancyLine{}
}

func runSetExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No commandset specified\n"))
		return flag.ErrHelp
	}

	if len(args) != 1 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	names, ok := conf.Commandsets[args[0]]
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: commandset %q not found :(\n", args[0]))
		return flag.ErrHelp
	}

	cmds := make([]Command, 0, len(names))
	for _, name := range names {
		cmd, ok := conf.Commands[name]
		if !ok {
			return fmt.Errorf("command %q not found in commandset %q", name, args[0])
		}

		cmds = append(cmds, cmd)
	}

	return run(ctx, cmds...)
}

func testExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No test suite specified\n"))
		return flag.ErrHelp
	}

	cmd, ok := conf.Tests[args[0]]
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: test suite %q not found :(\n", args[0]))
		return flag.ErrHelp
	}

	return runTest(ctx, cmd, args[1:])
}

func startExec(ctx context.Context, args []string) error {
	if len(args) != 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	return runSetExec(ctx, []string{"default"})
}

func runExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No command specified\n"))
		return flag.ErrHelp
	}

	if len(args) != 1 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	cmd, ok := conf.Commands[args[0]]
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: command %q not found :(\n", args[0]))
		return flag.ErrHelp
	}

	return run(ctx, cmd)
}

func doctorExec(ctx context.Context, args []string) error {
	return runChecks(ctx, conf.Checks)
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
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: environment %q not found :(\n", args[0]))
		return flag.ErrHelp
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
	)

	if !isValidDatabaseName(databaseName) {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", databaseName))
		return flag.ErrHelp
	}

	upFile, downFile, err := createNewMigration(databaseName, migrationName)
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

	if !isValidDatabaseName(*migrationUpDatabaseNameFlag) {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", *migrationUpDatabaseNameFlag))
		return flag.ErrHelp
	}

	// TODO
	return errors.New("up unimplemented")
}

func migrationDownExec(ctx context.Context, args []string) error {
	if len(args) != 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments\n"))
		return flag.ErrHelp
	}

	if !isValidDatabaseName(*migrationDownDatabaseNameFlag) {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", *migrationDownDatabaseNameFlag))
		return flag.ErrHelp
	}

	// TODO
	return errors.New("down unimplemented")
}

func migrationSquashExec(ctx context.Context, args []string) error {
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
	)

	if !isValidDatabaseName(databaseName) {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(\n", databaseName))
		return flag.ErrHelp
	}

	currentVersion, err := semver.NewVersion(migrationName)
	if err != nil {
		return err
	}

	commit := fmt.Sprintf("v%d.%d.0", currentVersion.Major(), currentVersion.Minor()-3) // TODO - define this as a constant

	lastMigrationIndex, ok, err := lastMigrationIndexAtCommit(databaseName, commit)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no migrations exist at commit %s", commit)
	}

	cmd := exec.Command(
		"docker", "run",
		"--rm", "-d",
		"--name", "squasher",
		"-p", "5432:5432",
		"-e", "POSTGRES_HOST_AUTH_METHOD=trust",
		"postgres:12.6",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "'docker %s' failed: %s", strings.Join(args, " "), out)
	}

	// TODO

	// DBNAME='squasher'
	// SERVER_VERSION=$(psql --version)

	// if [ "${SERVER_VERSION}" != 12.6 ]; then
	//   echo "running PostgreSQL 12.6 in docker since local version is ${SERVER_VERSION}"
	//   docker image inspect postgres:12.6 >/dev/null || docker pull postgres:12.6
	//   docker rm --force "${DBNAME}" 2>/dev/null || true
	//   docker run --rm --name "${DBNAME}" -p 5433:5432 -e POSTGRES_HOST_AUTH_METHOD=trust -d postgres:12.6

	//   function kill() {
	//     docker kill "${DBNAME}" >/dev/null
	//   }
	//   trap kill EXIT

	//   sleep 5
	//   docker exec -u postgres "${DBNAME}" createdb "${DBNAME}"
	//   export PGHOST=127.0.0.1
	//   export PGPORT=5433
	//   export PGDATABASE="${DBNAME}"
	//   export PGUSER=postgres
	// fi

	// # First, apply migrations up to the version we want to squash
	// migrate -database "postgres://${PGHOST}:${PGPORT}/${PGDATABASE}?sslmode=disable&x-migrations-table=${migrations_table}" -path . goto "${VERSION}"

	// # Dump the database into a temporary file that we need to post-process
	// pg_dump --schema-only --no-owner --no-comments --exclude-table='*schema_migrations' -f tmp_squashed.sql

	// # Remove settings header from pg_dump output
	// sed -i '' -e 's/^SET .*$//g' tmp_squashed.sql
	// sed -i '' -e 's/^SELECT pg_catalog.set_config.*$//g' tmp_squashed.sql

	// # Do not drop extensions if they already exist. This causes some
	// # weird problems with the back-compat tests as the extensions are
	// # not dropped in the correct order to honor dependencies.
	// sed -i '' -e 's/^DROP EXTENSION .*$//g' tmp_squashed.sql

	// # Remove references to public schema
	// sed -i '' -e 's/public\.//g' tmp_squashed.sql
	// sed -i '' -e 's/ WITH SCHEMA public//g' tmp_squashed.sql

	// # Remove comments, multiple blank lines
	// sed -i '' -e 's/^--.*$//g' tmp_squashed.sql
	// sed -i '' -e '/^$/N;/^\n$/D' tmp_squashed.sql

	// # Now clean up all of the old migration files. `ls` will return files in
	// # alphabetical order, so we can delete all files from the migration directory
	// # until we hit our squashed migration.

	filenames, err := removeMigrationFilesBefore(databaseName, lastMigrationIndex)
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		fmt.Printf("> Squashed migration file %s\n", filename)
	}

	// # Wrap squashed migration in transaction
	// printf "BEGIN;\n" >"./${VERSION}_squashed_migrations.up.sql"
	// cat tmp_squashed.sql >>"./${VERSION}_squashed_migrations.up.sql"
	// printf "\nCOMMIT;\n" >>"./${VERSION}_squashed_migrations.up.sql"
	// rm tmp_squashed.sql

	// cat >"./${VERSION}_squashed_migrations.down.sql" <<EOL
	// DROP SCHEMA IF EXISTS public CASCADE;
	// CREATE SCHEMA public;

	// CREATE TABLE IF NOT EXISTS ${migrations_table} (
	//     version bigint NOT NULL PRIMARY KEY,
	//     dirty boolean NOT NULL
	// );
	// EOL

	// echo ""
	// echo "squashed migrations written to ${VERSION}_squashed_migrations.{up,down}.sql"

	return errors.New("squash unimplemented")
}

func printRunUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg %s <command>\n", c.Name)

	// Attempt to parse config to list available commands, but don't fail on
	// error, because we should never error when the user wants --help output.
	_, _ = parseConf(*configFlag, *overwriteConfigFlag)

	if conf != nil {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE COMMANDS IN %s%s%s\n", output.StyleBold, *configFlag, output.StyleReset)

		for name := range conf.Commands {
			fmt.Fprintf(&out, "  %s\n", name)
		}
	}

	return out.String()
}

func printTestUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg %s <test suite>\n", c.Name)

	// Attempt to parse config so we can list test suites, but don't fail on
	// error, because we should never error when the user wants --help output.
	_, _ = parseConf(*configFlag, *overwriteConfigFlag)

	if conf != nil {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE TESTSUITES IN %s%s%s\n", output.StyleBold, *configFlag, output.StyleReset)

		for name := range conf.Tests {
			fmt.Fprintf(&out, "  %s\n", name)
		}
	}

	return out.String()
}

func printRunSetUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg %s <commandset>\n", c.Name)

	// Attempt to parse config so we can list available sets, but don't fail on
	// error, because we should never error when the user wants --help output.
	_, _ = parseConf(*configFlag, *overwriteConfigFlag)
	if conf != nil {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE COMMANDSETS IN %s%s%s\n", output.StyleBold, *configFlag, output.StyleReset)

		for name := range conf.Commandsets {
			fmt.Fprintf(&out, "  %s\n", name)
		}
	}

	return out.String()
}

func printStartUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintln(&out, "  sg start")

	return out.String()
}

func printDoctorUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg doctor\n")

	return out.String()
}

func printLiveUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg live <environment>\n")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE ENVIRONMENTS\n")

	for _, name := range environmentNames() {
		fmt.Fprintf(&out, "  %s\n", name)
	}

	return out.String()
}

func printMigrationUsage(c *ffcli.Command) string {
	var out strings.Builder

	printLogo(&out)

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg migration <subcommand>\n")

	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE COMMANDS\n")
	for _, sub := range c.Subcommands {
		fmt.Fprintf(&out, "  %s\n", sub.Name)
	}

	fmt.Fprintf(&out, "\nRun 'sg migration <subcommand> -help' to get help output for each subcommand\n")

	return out.String()
}

func printMigrationAddUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg migration add [-db=%s] <name>\n", defaultDatabaseName)
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE DATABASES\n")

	for _, name := range databaseNames {
		fmt.Fprintf(&out, "  %s\n", name)
	}

	return out.String()
}

func printMigrationUpUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg migration up [-db=%s] [-n=1]\n", defaultDatabaseName)
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE DATABASES\n")

	for _, name := range databaseNames {
		fmt.Fprintf(&out, "  %s\n", name)
	}

	return out.String()
}

func printMigrationDownUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg migration down [-db=%s] [-n=1]\n", defaultDatabaseName)
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE DATABASES\n")

	for _, name := range databaseNames {
		fmt.Fprintf(&out, "  %s\n", name)
	}

	return out.String()
}

func printMigrationSquashUsage(c *ffcli.Command) string {
	var out strings.Builder

	fmt.Fprintf(&out, "USAGE\n")
	fmt.Fprintf(&out, "  sg migration squash [-db=%s] <current-release>\n", defaultDatabaseName)
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE DATABASES\n")

	for _, name := range databaseNames {
		fmt.Fprintf(&out, "  %s\n", name)
	}

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
