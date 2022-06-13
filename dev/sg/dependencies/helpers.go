package dependencies

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

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// cmdFix executes the given command as an action in a new user shell.
func cmdFix(cmd string) check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		c := usershell.Command(ctx, cmd)
		if cio.Input != nil {
			c = c.Input(cio.Input)
		}
		return c.Run().StreamLines(cio.Verbose)
	}
}

func cmdFixes(cmds ...string) check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		for _, cmd := range cmds {
			if err := cmdFix(cmd)(ctx, cio, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func enableForTeammatesOnly() check.EnableFunc[CheckArgs] {
	return func(ctx context.Context, args CheckArgs) error {
		if !args.Teammate {
			return errors.New("disabled if not a Sourcegraph teammate")
		}
		return nil
	}
}

func disableInCI() check.EnableFunc[CheckArgs] {
	return func(ctx context.Context, args CheckArgs) error {
		// Docker is quite funky in CI
		if os.Getenv("CI") == "true" {
			return errors.New("disabled in CI")
		}
		return nil
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

func checkSourcegraphDatabase(ctx context.Context, out *std.Output, args CheckArgs) error {
	// This check runs only in the `sourcegraph/sourcegraph` repository, so
	// we try to parse the globalConf and use its `Env` to configure the
	// Postgres connection.
	config, _ := sgconf.Get(args.ConfigFile, args.ConfigOverwriteFile)
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

func getToolVersionConstraint(ctx context.Context, tool string) (string, error) {
	tools, err := root.Run(run.Cmd(ctx, "cat .tool-versions")).Lines()
	if err != nil {
		return "", err
	}
	var version string
	for _, t := range tools {
		parts := strings.Split(t, " ")
		if parts[0] == tool {
			version = parts[1]
			break
		}
	}
	if version == "" {
		return "", errors.Newf("tool %q not found in .tool-versions", tool)
	}
	return fmt.Sprintf("~> %s", version), nil
}

func checkGoVersion(ctx context.Context, out *std.Output, args CheckArgs) error {
	if err := check.InPath("go")(ctx); err != nil {
		return err
	}

	constraint, err := getToolVersionConstraint(ctx, "golang")
	if err != nil {
		return err
	}

	cmd := "go version"
	data, err := usershell.Run(ctx, cmd).String()
	if err != nil {
		return errors.Wrapf(err, "failed to run %q", cmd)
	}
	parts := strings.Split(strings.TrimSpace(data), " ")
	if len(parts) == 0 {
		return errors.Newf("no output from %q", cmd)
	}

	return check.Version("go", strings.TrimPrefix(parts[2], "go"), constraint)
}

func checkYarnVersion(ctx context.Context, out *std.Output, args CheckArgs) error {
	if err := check.InPath("yarn")(ctx); err != nil {
		return err
	}

	constraint, err := getToolVersionConstraint(ctx, "yarn")
	if err != nil {
		return err
	}

	cmd := "yarn --version"
	data, err := usershell.Run(ctx, cmd).String()
	if err != nil {
		return errors.Wrapf(err, "failed to run %q", cmd)
	}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return errors.Newf("no output from %q", cmd)
	}

	return check.Version("yarn", trimmed, constraint)
}

func checkNodeVersion(ctx context.Context, out *std.Output, args CheckArgs) error {
	if err := check.InPath("node")(ctx); err != nil {
		return err
	}

	constraint, err := getToolVersionConstraint(ctx, "nodejs")
	if err != nil {
		return err
	}

	cmd := "node --version"
	data, err := usershell.Run(ctx, cmd).String()
	if err != nil {
		return errors.Wrapf(err, "failed to run %q", cmd)
	}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return errors.Newf("no output from %q", cmd)
	}

	return check.Version("yarn", trimmed, constraint)
}

func checkRustVersion(ctx context.Context, out *std.Output, args CheckArgs) error {
	if err := check.InPath("cargo")(ctx); err != nil {
		return err
	}

	constraint, err := getToolVersionConstraint(ctx, "rust")
	if err != nil {
		return err
	}

	cmd := "cargo --version"
	data, err := usershell.Run(ctx, cmd).String()
	if err != nil {
		return errors.Wrapf(err, "failed to run %q", cmd)
	}
	parts := strings.Split(strings.TrimSpace(data), " ")
	if len(parts) == 0 {
		return errors.Newf("no output from %q", cmd)
	}

	return check.Version("cargo", parts[1], constraint)
}

// check1password defines the 1password dependency check which is uniform across platforms.
func check1password() check.CheckFunc {
	return check.Combine(
		check.WrapErrMessage(check.InPath("op"), "The 1password CLI, 'op', is required"),
		// We must 'list' before trying to 'get', otherwise 'get' will start trying to
		// prompt for things and mess up our output
		check.CommandOutputContains("op account list", "team-sourcegraph.1password.com"),
		check.CommandOutputContains("op account get --account team-sourcegraph.1password.com", "ACTIVE"))
}

func forceASDFPluginAdd(ctx context.Context, plugin string, source string) error {
	err := usershell.Run(ctx, "asdf plugin-add", plugin, source).Wait()
	if err != nil && strings.Contains(err.Error(), "already added") {
		return nil
	}
	return err
}
