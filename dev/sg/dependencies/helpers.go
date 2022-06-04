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

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

func checkSourcegraphDatabase(ctx context.Context, cio check.IO, args CheckArgs) error {
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
