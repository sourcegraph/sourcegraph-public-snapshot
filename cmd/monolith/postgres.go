package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
)

// make the "en_US.UTF-8" locale so postgres will be utf-8 enabled by default
// alpine doesn't require explicit locale-file generation

//docker:env LANG=en_US.utf8

// We run 9.4 in production, but if we are embedding might as well get
// something modern 9.6. We add the version specifier to prevent accidently
// upgrading to an even newer version.

//docker:install 'postgresql<9.7' su-exec
//docker:install 'postgresql-contrib<9.7' su-exec

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

	// postgres wants its config in the data dir
	path := filepath.Join(os.Getenv("DATA_DIR"), "postgresql")
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		log.Printf("Setting up postgres at %s", path)
		log.Println("This may take a few moments...")

		var output bytes.Buffer
		e := execer{Out: &output}
		e.Command("mkdir", "-p", path)
		e.Command("chown", "postgres", path)
		e.Command("su-exec", "postgres", "initdb", "-D", path)
		e.Command("su-exec", "postgres", "pg_ctl", "-D", path, "-o -c listen_addresses=localhost", "-l", "/tmp/pgsql.log", "-w", "start")
		e.Command("su-exec", "postgres", "createdb", "sourcegraph")
		e.Command("su-exec", "postgres", "pg_ctl", "-D", path, "-m", "fast", "-l", "/tmp/pgsql.log", "-w", "stop")
		if err := e.Error(); err != nil {
			log.Printf("Setting up postgres failed:\n%s", output.String())
			os.RemoveAll(path)
			return "", err
		}
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

	setDefaultEnv("PGUSER", "postgres")
	setDefaultEnv("PGDATABASE", "sourcegraph")
	setDefaultEnv("PGSSLMODE", "disable")

	return "postgres: su-exec postgres postgres -D " + path, nil
}
