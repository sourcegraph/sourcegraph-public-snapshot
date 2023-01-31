package archive

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
)

const (
	ROOT     = "databases" // root folder within archive to hold all DBArchive outputs
	PGSQL    = "pgsql.out"
	INTEL    = "code-intel.out"
	INSIGHTS = "code-insights.out"
)

type DBArchive struct {
	name string
	dsn  string
}

func (d *DBArchive) archive(root string) error {
	path := filepath.Join(root, d.name)
	dump, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dump.Close()

	// Fc - custom output format
	// c  - include "drop if exists" sql for restore
	// f  - output file
	return exec.Command("pg_dump", "-Fc", "-c", "-f", path, "--dbname", d.dsn).Run()
}

func (d *DBArchive) restore(root string) error {
	dump := filepath.Join(root, d.name)
	fmt.Println(*d)
	return exec.Command("pg_restore", "-c", "--dbname", d.dsn, dump).Run()
}

func (d *DBArchive) root() string {
	return ROOT
}

func dbSources() []Archiver {
	dsns := dsn()
	dbs := make([]Archiver, 0, len(dsns))
	for name, dsn := range dsns {
		dbs = append(dbs, &DBArchive{
			name: name,
			dsn:  dsn,
		})
	}

	return dbs
}

func dbIdentify(path string) []Archiver {
	path = filepath.Join(path, ROOT)
	if _, err := os.Stat(path); err != nil {
		return nil
	}

	dsns := dsn()

	dbs := make([]Archiver, 0, len(dsns))
	for name, dsn := range dsns {
		p := filepath.Join(path, name)
		if _, err := os.Stat(p); err == nil {
			dbs = append(dbs, &DBArchive{
				name: name,
				dsn:  dsn,
			})
		}
	}

	return dbs
}

func dsn() map[string]string {
	// TODO(jac): does this need better validation?
	m := make(map[string]string, 3)
	m[PGSQL] = postgresdsn.New("frontend", "", os.Getenv)
	m[INTEL] = postgresdsn.New("CODEINTEL", "", os.Getenv)
	m[INSIGHTS] = postgresdsn.New("CODEINSIGHTS", "", os.Getenv)

	// pg_dump errors if TimeZone is supplied
	for name, dsn := range m {
		dsn, _ := url.Parse(dsn)
		q := dsn.Query()
		q.Del("timezone")
		dsn.RawQuery = q.Encode()
		m[name] = dsn.String()
	}

	return m
}
