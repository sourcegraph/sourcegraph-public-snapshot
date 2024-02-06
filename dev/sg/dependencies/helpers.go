package dependencies

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/grafana/regexp"
	"github.com/jackc/pgx/v4"

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

func enableOnlyInSourcegraphRepo() check.EnableFunc[CheckArgs] {
	return func(ctx context.Context, args CheckArgs) error {
		_, err := root.RepositoryRoot()
		return err
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

// checkPostgresConnection succeeds connecting to the default user database works, regardless
// of if it's running locally or with docker.
func checkPostgresConnection(ctx context.Context) error {
	dsns, err := dsnCandidates()
	if err != nil {
		return err
	}
	var errs []error
	for _, dsn := range dsns {
		// If any of the candidates succeed, we're good
		if err := pingPG(ctx, dsn); err == nil {
			return nil
		} else {
			errs = append(errs, err)
		}
	}

	messages := []string{"failed all attempts to connect to Postgresql database"}
	for _, e := range errs {
		messages = append(messages, "\t"+e.Error())
	}
	return errors.New(strings.Join(messages, "\n"))
}

func pingPG(ctx context.Context, dsn string) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to Postgresql Database at %s", dsn)
	}
	defer conn.Close(ctx)

	for {
		err := conn.Ping(ctx)
		if err == nil {
			return nil
		}
		// If database is starting up we keep waiting
		if strings.Contains(err.Error(), "database system is starting up") {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		return errors.Wrapf(err, "failed to ping database at %s", dsn)
	}
}

func dsnCandidates() ([]string, error) {
	var candidates []string

	// most classic dsn
	baseURL := url.URL{Scheme: "postgres", Host: "127.0.0.1:5432"}

	env := func(key string) string { val, _ := os.LookupEnv(key); return val }
	add := func(dsn string) { candidates = append(candidates, dsn) }

	withUserPass := func(user, password string) func(dsn url.URL) {
		return func(dsn url.URL) {
			if password == "" {
				dsn.User = url.User(user)
			} else {
				dsn.User = url.UserPassword(user, password)
			}
		}
	}

	withPath := func(path string) func(dsn url.URL) {
		return func(dsn url.URL) {
			dsn.Path = path
		}
	}

	withSSL := func(sslmode string) func(dsn url.URL) {
		return func(dsn url.URL) {
			if sslmode != "" {
				qry := dsn.Query()
				qry.Set("sslmode", sslmode)
				dsn.RawQuery = qry.Encode()
			}
		}
	}

	withHost := func(host, port string) func(dsn url.URL) {
		return func(dsn url.URL) {
			if host == "" {
				return
			}
			if port == "" {
				port = "5432"
			}
			dsn.Host = fmt.Sprintf("%s:%s", host, "5432")
		}
	}

	addURL := func(modifiers ...func(dsn url.URL)) {
		dsn := baseURL
		for _, modifier := range modifiers {
			modifier(dsn)
		}
		add(dsn.String())
	}

	// best case scenario
	add(env("PGDATASOURCE"))

	// homebrew dsn
	if uinfo, err := user.Current(); err == nil {
		addURL(
			withUserPass(uinfo.Username, ""),
			withPath("postgres"),
		)
	}

	// classic docker dsn
	addURL(withUserPass("postgres", "postgres"))
	// other classic docker dsn
	addURL(withUserPass("postgres", "password"))

	// env based dsn
	username, ok := os.LookupEnv("PGUSER")
	if !ok {
		uinfo, err := user.Current()
		if err != nil {
			return nil, err
		}
		username = uinfo.Username
	}

	addURL(
		withUserPass(username, env("PGPASSWORD")),
		withHost(env("PGHOST"), env("PGPORT")),
		withSSL(env("PGSSLMODE")),
	)

	return candidates, nil
}

func checkSourcegraphDatabase(ctx context.Context, out *std.Output, args CheckArgs) error {
	// This check runs only in the `sourcegraph/sourcegraph` repository, so
	// we try to parse the globalConf and use its `Env` to configure the
	// Postgres connection.
	var config *sgconf.Config
	if args.DisableOverwrite {
		config, _ = sgconf.GetWithoutOverwrites(args.ConfigFile)
	} else {
		config, _ = sgconf.Get(args.ConfigFile, args.ConfigOverwriteFile)
	}
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

	if err := pingPG(ctx, dsn); err != nil {
		return errors.Wrapf(err, "failed to connect to Sourcegraph Postgres database at %s. Please check the settings in sg.config.yml (see https://docs.sourcegraph.com/dev/background-information/sg#changing-database-configuration)", dsn)
	}
	return nil
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
		out, err := usershell.Command(ctx, "git version").StdOut().Run().String()
		if err != nil {
			return errors.Wrapf(err, "failed to run 'git version'")
		}

		elems := strings.Split(out, " ")
		if len(elems) != 3 && len(elems) != 5 {
			return errors.Newf("unexpected output from git: %s", out)
		}

		trimmed := strings.TrimSpace(elems[2])
		return check.Version("git", trimmed, versionConstraint)
	}
}

// func checkPostgresVersion(dsn, versionConstraint string) func(context.Context) error {
// 	return func(ctx context.Context) error {
// 		out, err := usershell.Command(ctx, `psql -t -c "select version()"`).StdOut().Run().String()
// 		if err != nil {
// 			return errors.Wrapf(err, "failed to get postgres version")
// 		}

// 		version := majorMinorVersionRegex.FindString(out)
// 		if version == "" {
// 			return errors.Newf("unexpected output from postgres: %s", out)
// 		}

// 		return check.Version("postgres", version, versionConstraint)
// 	}
// }

func checkSrcCliVersion(versionConstraint string) func(context.Context) error {
	return func(ctx context.Context) error {
		lines, err := usershell.Command(ctx, "src version -client-only").StdOut().Run().Lines()
		if err != nil {
			return errors.Wrapf(err, "failed to run 'src version'")
		}

		if len(lines) < 1 {
			return errors.Newf("unexpected output from src: %s", strings.Join(lines, "\n"))
		}
		out := lines[0]

		elems := strings.Split(out, " ")
		if len(elems) != 3 {
			return errors.Newf("unexpected output from src: %s", out)
		}

		trimmed := strings.TrimSpace(elems[2])

		// If the user is using a local dev build, let them get away.
		if trimmed == "dev" {
			return nil
		}
		return check.Version("src", trimmed, versionConstraint)
	}
}

func forceASDFPluginAdd(ctx context.Context, plugin string, source string) error {
	err := usershell.Run(ctx, "asdf plugin-add", plugin, source).Wait()
	if err != nil && strings.Contains(err.Error(), "already added") {
		return nil
	}
	return errors.Wrap(err, "asdf plugin-add")
}

// pgUtilsPathRe is the regexp used to check what value user.bazelrc defines for
// the PG_UTILS_PATH env var.
var pgUtilsPathRe = regexp.MustCompile(`build --action_env=PG_UTILS_PATH=(.*)$`)

// userBazelRcPath is the path to a git ignored file that contains Bazel flags
// specific to the current machine that are required in certain cases.
var userBazelRcPath = ".aspect/bazelrc/user.bazelrc"

// checkPGUtilsPath ensures that a PG_UTILS_PATH is being defined in .aspect/bazelrc/user.bazelrc
// if it's needed. For example, on Linux hosts, it's usually located in /usr/bin, which is
// perfectly fine. But on Mac machines, it's either in the homebrew PATH or on a different
// location if the user installed Posgres through the Postgresql desktop app.
func checkPGUtilsPath(ctx context.Context, out *std.Output, args CheckArgs) error {
	// Check for standard PATH location, that is available inside Bazel when
	// inheriting the shell environment. That is just /usr/bin, not /usr/local/bin.
	_, err := os.Stat("/usr/bin/createdb")
	if err == nil {
		// If we have createdb in /usr/bin/, nothing to do, it will work outside the box.
		return nil
	}

	// Check for the presence of git ignored user.bazelrc, that is specific to local
	// environment. Because createdb is not under /usr/bin, we have to create that file
	// and define the PG_UTILS_PATH for migration rules.
	_, err = os.Stat(userBazelRcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Wrapf(err, "%s doesn't exist", userBazelRcPath)
		}
		return errors.Wrapf(err, "unexpected error with %s", userBazelRcPath)
	}

	// If it exists, we check if the injected PATH actually contains createdb as intended.
	// If not, we'll raise an error for sg setup to correct.
	f, err := os.Open(userBazelRcPath)
	if err != nil {
		return errors.Wrapf(err, "can't open %s", userBazelRcPath)
	}
	defer f.Close()

	err, pgUtilsPath := parsePgUtilsPathInUserBazelrc(f)
	if err != nil {
		return errors.Wrapf(err, "can't parse %s", userBazelRcPath)
	}

	// If the file exists, but doesn't reference PG_UTILS_PATH, that's an error as well.
	if pgUtilsPath == "" {
		return errors.Newf("none on the content in %s matched %q", userBazelRcPath, pgUtilsPathRe.String())
	}

	// Check that this path contains createdb as expected.
	if err := checkPgUtilsPathIncludesBinaries(pgUtilsPath); err != nil {
		return err
	}

	return nil
}

// parsePgUtilsPathInUserBazelrc extracts the defined path to the createdb postgresql
// utilities that are used in a the Bazel migration rules.
func parsePgUtilsPathInUserBazelrc(r io.Reader) (error, string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		matches := pgUtilsPathRe.FindStringSubmatch(line)
		if len(matches) > 1 {
			return nil, matches[1]
		}
	}
	return scanner.Err(), ""
}

// checkPgUtilsPathIncludesBinaries ensures that the given path contains createdb as expected.
func checkPgUtilsPathIncludesBinaries(pgUtilsPath string) error {
	_, err := os.Stat(path.Join(pgUtilsPath, "createdb"))
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Wrap(err, "currently defined PG_UTILS_PATH doesn't include createdb")
		}
		return errors.Wrap(err, "currently defined PG_UTILS_PATH is incorrect")
	}
	return nil
}

// guessPgUtilsPath infers from the environment where the createdb binary
// is located and returns its parent folder, so it can be used to extend
// PATH for the migrations Bazel rules.
func guessPgUtilsPath(ctx context.Context) (error, string) {
	str, err := usershell.Run(ctx, "which", "createdb").String()
	if err != nil {
		return err, ""
	}
	return nil, filepath.Dir(str)
}

func caskInstall(formula string) check.FixAction[CheckArgs] {
	return createBrewInstallFix(formula, true)

}

func brewInstall(formula string) check.FixAction[CheckArgs] {
	return createBrewInstallFix(formula, false)
}

// brewInstall returns a FixAction that installs a brew formula.
// If the brew output contains an autofix for adding the formula to the path
// (in the case of keg-only formula), it will be automatically applied.
func createBrewInstallFix(formula string, cask bool) check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		cmd := "brew install "
		if cask {
			cmd += "--cask "
		}
		cmd += formula
		c := usershell.Command(ctx, cmd)
		if cio.Input != nil {
			c = c.Input(cio.Input)
		}

		pathAddCommandIsNext := false
		return c.Run().StreamLines(func(line string) {
			if pathAddCommandIsNext {
				_ = usershell.Run(ctx, line).Wait()
				pathAddCommandIsNext = false
			}
			if strings.Contains(line, "If you need to have "+formula+" first in your PATH, run:") {
				pathAddCommandIsNext = true
			}
			cio.Verbose(line)
		})
	}
}
