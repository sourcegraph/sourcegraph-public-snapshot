package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type runFunc func(quiet bool, cmd ...string) (string, error)

const databaseNamePrefix = "schemadoc-gen-temp-"

const containerName = "schemadoc"

var logger = log.New(os.Stderr, "", log.LstdFlags)

var versionRe = lazyregexp.New(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(`12\.\d+`)))

var databases = map[*dbconn.Database]string{
	dbconn.Frontend:  "schema.md",
	dbconn.CodeIntel: "schema.codeintel.md",
}

// This script generates markdown formatted output containing descriptions of
// the current dabase schema, obtained from postgres. The correct PGHOST,
// PGPORT, PGUSER etc. env variables must be set to run this script.
func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	// Run pg12 locally if it exists
	if version, _ := exec.Command("psql", "--version").CombinedOutput(); versionRe.Match(version) {
		return mainLocal()
	}

	return mainContainer()
}

func mainLocal() error {
	dataSourcePrefix := "dbname=" + databaseNamePrefix

	for database, destinationFile := range databases {
		if err := generateAndWrite(database, dataSourcePrefix+database.Name, nil, destinationFile); err != nil {
			return err
		}
	}

	return nil
}

func mainContainer() error {
	logger.Printf("Running PostgreSQL 12 in docker")

	prefix, shutdown, err := startDocker()
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown()

	dataSourcePrefix := "postgres://postgres@127.0.0.1:5433/postgres?dbname=" + databaseNamePrefix

	for database, destinationFile := range databases {
		if err := generateAndWrite(database, dataSourcePrefix+database.Name, prefix, destinationFile); err != nil {
			return err
		}
	}

	return nil
}

func generateAndWrite(database *dbconn.Database, dataSource string, commandPrefix []string, destinationFile string) error {
	run := runWithPrefix(commandPrefix)

	// Try to drop a database if it already exists
	_, _ = run(true, "dropdb", databaseNamePrefix+database.Name)

	// Let's also try to clean up after ourselves
	defer func() { _, _ = run(true, "dropdb", databaseNamePrefix+database.Name) }()

	if out, err := run(false, "createdb", databaseNamePrefix+database.Name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("run: %s", out))
	}

	out, err := generateInternal(database, dataSource, run)
	if err != nil {
		return err
	}

	return os.WriteFile(destinationFile, []byte(out), os.ModePerm)
}

func startDocker() (commandPrefix []string, shutdown func(), _ error) {
	if err := exec.Command("docker", "image", "inspect", "postgres:12").Run(); err != nil {
		logger.Println("docker pull postgres:12")
		pull := exec.Command("docker", "pull", "postgres:12")
		pull.Stdout = logger.Writer()
		pull.Stderr = logger.Writer()
		if err := pull.Run(); err != nil {
			return nil, nil, errors.Wrap(err, "docker pull postgres:12")
		}
		logger.Println("docker pull complete")
	}

	run := runWithPrefix(nil)

	_, _ = run(true, "docker", "rm", "--force", containerName)
	server := exec.Command("docker", "run", "--rm", "--name", containerName, "-e", "POSTGRES_HOST_AUTH_METHOD=trust", "-p", "5433:5432", "postgres:12")
	if err := server.Start(); err != nil {
		return nil, nil, errors.Wrap(err, "docker run")
	}

	shutdown = func() {
		_ = server.Process.Kill()
		_, _ = run(true, "docker", "kill", containerName)
		_ = server.Wait()
	}

	attempts := 0
	for {
		attempts++
		// TODO - not sure why this would work...?
		if err := exec.Command("pg_isready", "-U", "postgres", "-h", "127.0.0.1", "-p", "5433").Run(); err == nil {
			break
		} else if attempts > 30 {
			shutdown()
			return nil, nil, errors.Wrap(err, "pg_isready timeout")
		}
		time.Sleep(time.Second)
	}

	return []string{"docker", "exec", "-u", "postgres", containerName}, shutdown, nil
}

func generateInternal(database *dbconn.Database, dataSource string, run runFunc) (string, error) {
	db, err := dbconn.NewRaw(dataSource)
	if err != nil {
		return "", errors.Wrap(err, "NewRaw")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := dbconn.MigrateDB(db, database); err != nil {
		return "", errors.Wrap(err, "MigrateDB")
	}

	tables, err := getTables(db)
	if err != nil {
		return "", err
	}

	types, err := describeTypes(db)
	if err != nil {
		return "", err
	}

	ch := make(chan table, len(tables))
	for _, table := range tables {
		ch <- table
	}
	close(ch)

	var mu sync.Mutex
	var wg sync.WaitGroup
	var docs []string

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for table := range ch {
				logger.Println("describe", table.name)

				doc, err := describeTable(db, database.Name, table, run)
				if err != nil {
					logger.Fatalf("error: %s", err)
					continue
				}

				mu.Lock()
				docs = append(docs, doc)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	sort.Strings(docs)

	combined := strings.Join(docs, "\n")

	if len(types) > 0 {
		buf := bytes.NewBuffer(nil)
		buf.WriteString("\n")

		var typeNames []string
		for k := range types {
			typeNames = append(typeNames, k)
		}
		sort.Strings(typeNames)

		for _, name := range typeNames {
			buf.WriteString("# Type ")
			buf.WriteString(name)
			buf.WriteString("\n\n- ")
			buf.WriteString(strings.Join(types[name], "\n- "))
			buf.WriteString("\n\n")
		}

		combined += buf.String()
	}

	return combined, nil
}

type table struct {
	name   string
	isView bool
}

func getTables(db *sql.DB) (tables []table, _ error) {
	// Query names of all public tables and views.
	rows, err := db.Query(`
		SELECT table_name, FALSE AS is_view FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		UNION
		SELECT table_name, TRUE AS is_view FROM information_schema.views WHERE table_schema = 'public' AND table_name != 'pg_stat_statements';
	`)
	if err != nil {
		return nil, errors.Wrap(err, "database.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var t table
		if err := rows.Scan(&t.name, &t.isView); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return tables, nil
}

func describeTable(db *sql.DB, databaseName string, table table, run runFunc) (string, error) {
	comment, err := getTableComment(db, table.name)
	if err != nil {
		return "", err
	}

	columnComments, err := getColumnComments(db, table.name)
	if err != nil {
		return "", err
	}

	// Get postgres "describe table" output.
	out, err := run(false, "psql", "-X", "--quiet", "--dbname", databaseNamePrefix+databaseName, "-c", fmt.Sprintf("\\d %s", table.name))
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("run: %s", out))
	}

	lines := strings.Split(out, "\n")

	buf := bytes.NewBuffer(nil)
	buf.WriteString("# ")
	buf.WriteString(strings.TrimSpace(lines[0]))
	buf.WriteString("\n")
	buf.WriteString("```\n")
	buf.WriteString(strings.Join(lines[1:], "\n"))
	buf.WriteString("```\n")

	if comment != "" {
		buf.WriteString("\n")
		buf.WriteString(comment)
		buf.WriteString("\n")
	}

	var columns []string
	for k := range columnComments {
		columns = append(columns, k)
	}
	sort.Strings(columns)

	for _, k := range columns {
		buf.WriteString("\n**")
		buf.WriteString(k)
		buf.WriteString("**: ")
		buf.WriteString(columnComments[k])
		buf.WriteString("\n")
	}

	if table.isView {
		buf.WriteString("\n## View query:\n\n```sql\n")
		q, err := getViewQuery(db, table.name)
		if err != nil {
			return "", err
		}
		buf.WriteString(q)
		buf.WriteString("\n```\n")
	}

	return buf.String(), nil
}

func getTableComment(db *sql.DB, table string) (comment string, _ error) {
	rows, err := db.Query("select obj_description($1::regclass)", table)
	if err != nil {
		return "", errors.Wrap(err, "database.Query")
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&dbutil.NullString{S: &comment}); err != nil {
			return "", errors.Wrap(err, "rows.Scan")
		}
	}
	if err = rows.Err(); err != nil {
		return "", errors.Wrap(err, "rows.Err")
	}

	return comment, nil
}

func getViewQuery(db *sql.DB, view string) (query string, _ error) {
	rows, err := db.Query("SELECT definition FROM pg_views WHERE viewname = $1", view)
	if err != nil {
		return "", errors.Wrap(err, "database.Query")
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&query); err != nil {
			return "", errors.Wrap(err, "rows.Scan")
		}
	}
	if err = rows.Err(); err != nil {
		return "", errors.Wrap(err, "rows.Err")
	}

	return query, nil
}

func getColumnComments(db *sql.DB, table string) (map[string]string, error) {
	rows, err := db.Query(`
		SELECT
			cols.column_name,
			(
				SELECT pg_catalog.col_description(c.oid, cols.ordinal_position::int)
				FROM pg_catalog.pg_class c
				WHERE c.oid = (SELECT cols.table_name::regclass::oid) AND c.relname = cols.table_name
			) as column_comment
		FROM information_schema.columns cols
		WHERE cols.table_name = $1;
	`, table)
	if err != nil {
		return nil, errors.Wrap(err, "database.Query")
	}
	defer rows.Close()

	comments := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &dbutil.NullString{S: &v}); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		if v != "" {
			comments[k] = v
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return comments, nil
}

func describeTypes(db *sql.DB) (map[string][]string, error) {
	rows, err := db.Query(`
		SELECT
			t.typname as type_name,
			array_agg(e.enumlabel ORDER BY e.enumsortorder) as values
		FROM pg_catalog.pg_type t
			JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
			JOIN pg_catalog.pg_enum e ON t.oid = e.enumtypid
		GROUP BY t.typname;
	`)
	if err != nil {
		return nil, errors.Wrap(err, "database.Query")
	}
	defer rows.Close()

	values := map[string][]string{}
	for rows.Next() {
		var k string
		var v []string
		if err := rows.Scan(&k, pq.Array(&v)); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		values[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return values, nil
}

func runWithPrefix(prefix []string) runFunc {
	return func(quiet bool, cmd ...string) (string, error) {
		cmd = append(prefix, cmd...)

		c := exec.Command(cmd[0], cmd[1:]...)
		if !quiet {
			c.Stderr = logger.Writer()
		}

		out, err := c.Output()
		return string(out), err
	}
}
