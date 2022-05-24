package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// NOTE: These checkFuncs are also used by `sg setup`, so make sure that when
// you change something here `sg setup` still works as expected.
var checks = map[string]check.CheckFunc{
	"sourcegraph-database":  checkSourcegraphDatabase,
	"postgres":              check.Any(checkSourcegraphDatabase, checkPostgresConnection),
	"redis":                 check.Retry(checkRedisConnection, 5, 500*time.Millisecond),
	"psql":                  check.InPath("psql"),
	"sourcegraph-test-host": check.FileContains("/etc/hosts", "sourcegraph.test"),
	"caddy-trusted":         checkCaddyTrusted,
	"asdf":                  check.CommandOutputContains("asdf", "version"),
	"git":                   check.Combine(check.InPath("git"), checkGitVersion(">= 2.34.1")),
	"yarn":                  check.Combine(check.InPath("yarn"), checkYarnVersion("~> 1.22.4")),
	"go":                    check.Combine(check.InPath("go"), checkGoVersion("~> 1.18.1")),
	"node":                  check.Combine(check.InPath("node"), check.CommandOutputContains(`node -e "console.log(\"foobar\")"`, "foobar")),
	"rust":                  check.Combine(check.InPath("cargo"), check.CommandOutputContains(`cargo version`, "1.58.0")),
	"docker-installed":      check.WrapErrMessage(check.InPath("docker"), "if Docker is installed and the check fails, you might need to start Docker.app and restart terminal and 'sg setup'"),
	"docker": check.WrapErrMessage(
		check.Combine(check.InPath("docker"), check.CommandExitCode("docker info", 0)),
		"Docker needs to be running",
	),
}

func getCheck(name string) check.CheckFunc {
	return func(ctx context.Context) error {
		fn, ok := checks[name]
		if !ok {
			return errors.Newf("check %q not found", name)
		}
		return fn(ctx)
	}
}

func runChecksWithName(ctx context.Context, names []string) error {
	funcs := make(map[string]check.CheckFunc, len(names))
	for _, name := range names {
		if check, ok := checks[name]; ok {
			funcs[name] = check
		} else {
			return errors.Newf("check %q not found", name)
		}
	}

	return runChecks(ctx, funcs)
}

func runChecks(ctx context.Context, checks map[string]check.CheckFunc) error {
	if len(checks) == 0 {
		return nil
	}

	std.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleBold, "Running %d checks...", len(checks)))

	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	// Scripts used in various CheckFuncs are typically written with bash-compatible shells in mind.
	// Because of this, we throw a warning in non-compatible shells and ask that
	// users set up environments in both their shell and bash to avoid issues.
	if !usershell.IsSupportedShell(ctx) {
		shell := usershell.ShellType(ctx)
		std.Out.WriteWarningf("You're running on unsupported shell '%s'. "+
			"If you run into error, you may run 'SHELL=(which bash) sg setup' to setup your environment.",
			shell)
	}

	var failed []string

	for name, check := range checks {
		p := std.Out.Pending(output.Linef(output.EmojiLightbulb, output.StylePending, "Running check %q...", name))

		if err := check(ctx); err != nil {
			p.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Check %q failed with the following errors:", name))

			std.Out.WriteLine(output.Styledf(output.StyleWarning, "%s", err))

			failed = append(failed, name)
		} else {
			p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Check %q success!", name))
		}
	}

	if len(failed) == 0 {
		return nil
	}

	std.Out.Write("")
	std.Out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleBold, "The following checks failed:"))
	for _, name := range failed {
		std.Out.Writef("- %s", name)
	}

	std.Out.Write("")
	std.Out.WriteSuggestionf("Run 'sg setup' to make sure your system is setup correctly")
	std.Out.Write("")

	return errors.Newf("%d failed checks", len(failed))
}

func checkInMainRepoOrRepoInDirectory(context.Context) error {
	_, err := root.RepositoryRoot()
	if err != nil {
		ok, err := pathExists("sourcegraph")
		if !ok || err != nil {
			return errors.New("'sg setup' is not run in sourcegraph and repository is also not found in current directory")
		}
		return nil
	}
	return nil
}

func checkDevPrivateInParentOrInCurrentDirectory(context.Context) error {
	ok, err := pathExists("dev-private")
	if ok && err == nil {
		return nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to check for dev-private repository")
	}

	p := filepath.Join(wd, "..", "dev-private")
	ok, err = pathExists(p)
	if ok && err == nil {
		return nil
	}
	return errors.New("could not find dev-private repository either in current directory or one above")
}

// checkPostgresConnection succeeds connecting to the default user database works, regardless
// of if it's running locally or with docker.
func checkPostgresConnection(ctx context.Context) error {
	dsns, err := dsnCandidates()
	if err != nil {
		return err
	}
	var errs []error
	for _, dsn := range dsns {
		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to connect to Postgresql Database at %s", dsn))
			continue
		}
		defer conn.Close(ctx)
		err = conn.Ping(ctx)
		if err == nil {
			// if ping passed
			return nil
		}
		errs = append(errs, errors.Wrapf(err, "failed to connect to Postgresql Database at %s", dsn))
	}

	messages := []string{"failed all attempts to connect to Postgresql database"}
	for _, e := range errs {
		messages = append(messages, "\t"+e.Error())
	}
	return errors.New(strings.Join(messages, "\n"))
}

func dsnCandidates() ([]string, error) {
	env := func(key string) string { val, _ := os.LookupEnv(key); return val }

	// best case scenario
	datasource := env("PGDATASOURCE")
	// most classic dsn
	baseURL := url.URL{Scheme: "postgres", Host: "127.0.0.1:5432"}
	// classic docker dsn
	dockerURL := baseURL
	dockerURL.User = url.UserPassword("postgres", "postgres")
	// other classic docker dsn
	dockerURL2 := baseURL
	dockerURL2.User = url.UserPassword("postgres", "password")
	// env based dsn
	envURL := baseURL
	username, ok := os.LookupEnv("PGUSER")
	if !ok {
		uinfo, err := user.Current()
		if err != nil {
			return nil, err
		}
		username = uinfo.Name
	}
	envURL.User = url.UserPassword(username, env("PGPASSWORD"))
	if host, ok := os.LookupEnv("PGHOST"); ok {
		if port, ok := os.LookupEnv("PGPORT"); ok {
			envURL.Host = fmt.Sprintf("%s:%s", host, port)
		}
		envURL.Host = fmt.Sprintf("%s:%s", host, "5432")
	}
	if sslmode := env("PGSSLMODE"); sslmode != "" {
		qry := envURL.Query()
		qry.Set("sslmode", sslmode)
		envURL.RawQuery = qry.Encode()
	}
	return []string{
		datasource,
		envURL.String(),
		baseURL.String(),
		dockerURL.String(),
		dockerURL2.String(),
	}, nil
}

func checkSourcegraphDatabase(ctx context.Context) error {
	// This check runs only in the `sourcegraph/sourcegraph` repository, so
	// we try to parse the globalConf and use its `Env` to configure the
	// Postgres connection.
	config, _ := sgconf.Get(configFile, configOverwriteFile)
	if config == nil {
		return errors.New("failed to read sg.config.yaml. This step of `sg setup` needs to be run in the `sourcegraph` repository")
	}

	getEnv := func(key string) string {
		// First look into process env, emulating the logic in makeEnv used
		// in internal/run/run.go
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
		// Otherwise check in globalConf.Env
		return config.Env[key]
	}

	dsn := postgresdsn.New("", "", getEnv)
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to Soucegraph Postgres database at %s. Please check the settings in sg.config.yml (see https://docs.sourcegraph.com/dev/background-information/sg#changing-database-configuration)", dsn)
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx)
}

func checkRedisConnection(context.Context) error {
	conn, err := redis.Dial("tcp", ":6379", redis.DialConnectTimeout(5*time.Second))
	if err != nil {
		return errors.Wrap(err, "failed to connect to Redis at 127.0.0.1:6379")
	}

	if _, err := conn.Do("SET", "sg-setup", "was-here"); err != nil {
		return errors.Wrap(err, "failed to write to Redis at 127.0.0.1:6379")
	}

	retval, err := redis.String(conn.Do("GET", "sg-setup"))
	if err != nil {
		return errors.Wrap(err, "failed to read from Redis at 127.0.0.1:6379")
	}

	if retval != "was-here" {
		return errors.New("failed to test write in Redis")
	}
	return nil
}

func checkGitVersion(versionConstraint string) func(context.Context) error {
	return func(ctx context.Context) error {
		out, err := usershell.CombinedExec(ctx, "git version")
		if err != nil {
			return errors.Wrapf(err, "failed to run 'git version'")
		}

		elems := strings.Split(string(out), " ")
		if len(elems) != 3 {
			return errors.Newf("unexpected output from git server: %s", out)
		}

		trimmed := strings.TrimSpace(elems[2])
		return check.Version("git", trimmed, versionConstraint)
	}
}

func checkGoVersion(versionConstraint string) func(context.Context) error {
	return func(ctx context.Context) error {
		cmd := "go version"
		out, err := usershell.CombinedExec(ctx, "go version")
		if err != nil {
			return errors.Wrapf(err, "failed to run %q", cmd)
		}

		elems := strings.Split(string(out), " ")
		if len(elems) != 4 {
			return errors.Newf("unexpected output from %q: %s", out)
		}

		haveVersion := strings.TrimPrefix(elems[2], "go")

		return check.Version("go", haveVersion, versionConstraint)
	}
}

func checkYarnVersion(versionConstraint string) func(context.Context) error {
	return func(ctx context.Context) error {
		cmd := "yarn --version"
		out, err := usershell.CombinedExec(ctx, cmd)
		if err != nil {
			return errors.Wrapf(err, "failed to run %q", cmd)
		}

		elems := strings.Split(string(out), "\n")
		if len(elems) == 0 {
			return errors.Newf("no output from %q", cmd)
		}

		trimmed := strings.TrimSpace(elems[0])
		return check.Version("yarn", trimmed, versionConstraint)
	}
}
