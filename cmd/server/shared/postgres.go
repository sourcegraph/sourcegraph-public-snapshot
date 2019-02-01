package shared

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func maybePostgresProcFile() (string, error) {
	// PG is already configured
	if os.Getenv("PGHOST") != "" || os.Getenv("PGDATASOURCE") != "" {
		return "", nil
	}

	// Postgres needs to be able to write to run
	var output bytes.Buffer
	e := execer{Out: &output}
	e.Command("mkdir", "-p", "/run/postgresql")
	e.Command("chown", "-R", "postgres", "/run/postgresql")
	if err := e.Error(); err != nil {
		log.Printf("Setting up postgres failed:\n%s", output.String())
		return "", err
	}

	// Version of Postgres we are running.
	const pgversion = "11"

	// postgres wants its config in the data dir
	path := filepath.Join(os.Getenv("DATA_DIR"), "postgresql")
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		if verbose {
			log.Printf("Setting up PostgreSQL at %s", path)
		}
		log.Println("✱ Sourcegraph is initializing the internal database... (may take 15-20 seconds)")

		var output bytes.Buffer
		e := execer{Out: &output}
		e.Command("mkdir", "-p", path)
		e.Command("chown", "postgres", path)
		// initdb --nosync saves ~3-15s on macOS during initial startup. By the time actual data lives in the
		// DB, the OS should have had time to fsync.
		e.Command("su-exec", "postgres", "initdb", "-D", path, "--nosync")
		e.Command("su-exec", "postgres", "pg_ctl", "-D", path, "-o -c listen_addresses=127.0.0.1", "-l", "/tmp/pgsql.log", "-w", "start")
		e.Command("su-exec", "postgres", "createdb", "sourcegraph")
		e.Command("su-exec", "postgres", "pg_ctl", "-D", path, "-m", "fast", "-l", "/tmp/pgsql.log", "-w", "stop")
		if err := e.Error(); err != nil {
			log.Printf("Setting up postgres failed:\n%s", output.String())
			os.RemoveAll(path)
			return "", err
		}
	} else if bs, err := ioutil.ReadFile(filepath.Join(path, "PG_VERSION")); err != nil {
		return "", errors.Wrap(err, "Failed to detect version of existing Postgres data directory")
	} else if version := string(bs); version != pgversion {
		log.Printf("✱ Sourcegraph needs to upgrade your Postgres data directory. Please run this command and try again:\n")
		cmd := []string{"docker", "run", "-v", os.Getenv("DATA_DIR") + ":/data"}
		for k, v := range map[string]string{
			"PGUSEROLD":         getenv("PGUSER", "postgres"),
			"PGUSERNEW":         getenv("PGUSER", "postgres"),
			"PGDATABASEOLD":     getenv("PGDATABASE", "sourcegraph"),
			"PGDATABASENEW":     getenv("PGDATABASE", "sourcegraph"),
			"POSTGRES_PASSWORD": getenv("POSTGRES_PASSWORD", ""),
			"PGDATAOLD":         path,
			"PGDATANEW":         path + "-" + pgversion,
		} {
			if v != "" {
				cmd = append(cmd, "-e", fmt.Sprintf("'%s=%s'", k, v))
			}
		}
		log.Printf("\t%s", strings.Join(cmd, " "))
		log.Printf("\tmv %s %s-%s", path, path, version)
		log.Printf("\tmv %s-%s %s", path, pgversion, path)
		return "", errors.New("Postgres data directory needs upgrade")
	} else {
		// Between restarts the owner of the volume may have changed. Ensure
		// postgres can still read it.
		var output bytes.Buffer
		e := execer{Out: &output}
		e.Command("chown", "-R", "postgres", path)
		if err := e.Error(); err != nil {
			log.Printf("Adjusting fs owners for postgres failed:\n%s", output.String())
			return "", err
		}
	}

	// Set PGHOST to default to 127.0.0.1, NOT localhost, as localhost does not correctly resolve in some environments
	// (see https://github.com/sourcegraph/issues/issues/34 and https://github.com/sourcegraph/sourcegraph/issues/9129).
	SetDefaultEnv("PGHOST", "127.0.0.1")
	SetDefaultEnv("PGUSER", "postgres")
	SetDefaultEnv("PGDATABASE", "sourcegraph")
	SetDefaultEnv("PGSSLMODE", "disable")

	return "postgres: su-exec postgres sh -c 'postgres -c listen_addresses=127.0.0.1 -D " + path + "' 2>&1 | grep -v 'database system was shut down' | grep -v 'MultiXact member wraparound' | grep -v 'database system is ready' | grep -v 'autovacuum launcher started' | grep -v 'the database system is starting up'", nil
}

func getenv(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v == "" {
		return def
	}
	return name
}
