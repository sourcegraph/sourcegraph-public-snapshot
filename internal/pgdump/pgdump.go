package pgdump

import (
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Targets represents configuration for each of Sourcegraph's databases.
type Targets struct {
	Primary      Target `yaml:"primary"`
	CodeIntel    Target `yaml:"codeintel"`
	CodeInsights Target `yaml:"codeinsights"`
}

// Target represents a database for pg_dump to export.
type Target struct {
	// Target is the DSN of the database deployment:
	//
	// - in docker, the name of the database container, e.g. pgsql, codeintel-db, codeinsights-db
	// - in k8s, the name of the deployment or statefulset, e.g. deploy/pgsql, sts/pgsql
	// - in plain pg_dump, the server host or socket directory
	Target string `yaml:"target"`

	DBName   string `yaml:"dbname"`
	Username string `yaml:"username"`

	// Only include password if non-sensitive
	Password string `yaml:"password"`
}

// Command generates a pg_dump command that can be used for on-prem-to-Cloud migrations.
func Command(t Target) string {
	dump := fmt.Sprintf("pg_dump --no-owner --format=p --no-acl --username=%s --dbname=%s",
		t.Username, t.DBName)
	if t.Password == "" {
		return dump
	}
	return fmt.Sprintf("PGPASSWORD=%s %s", t.Password, dump)
}

type Output struct {
	Output string
	Target Target
}

// Outputs generates a set of mappings between a pgdump.Target and the desired output
// path. It can be provided a zero-value Targets to just generate the output paths.
func Outputs(dir string, targets Targets) []Output {
	return []Output{{
		Output: filepath.Join(dir, "primary.sql"),
		Target: targets.Primary,
	}, {
		Output: filepath.Join(dir, "codeintel.sql"),
		Target: targets.CodeIntel,
	}, {
		Output: filepath.Join(dir, "codeinsights.sql"),
		Target: targets.CodeInsights,
	}}
}

type CommandBuilder func(Target) (string, error)

// BuildCommands generates commands that output Postgres dumps and sends them to predefined
// files for each target database.
func BuildCommands(outDir string, commandBuilder CommandBuilder, targets Targets) ([]string, error) {
	var commands []string
	for _, t := range Outputs(outDir, targets) {
		c, err := commandBuilder(t.Target)
		if err != nil {
			return nil, errors.Wrapf(err, "generating command for %q", t.Output)
		}
		commands = append(commands, fmt.Sprintf("%s > %s", c, t.Output))
	}
	return commands, nil
}
