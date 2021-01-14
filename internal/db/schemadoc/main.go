package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"

	_ "github.com/lib/pq"
)

const dbname = "schemadoc-gen-temp"

var versionRe = lazyregexp.New(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta("9.6")))

func main() {
	out, err := generate(log.New(os.Stderr, "", log.LstdFlags), os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		if err := ioutil.WriteFile(os.Args[2], []byte(out), 0644); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Print(out)
	}
}

// This script generates markdown formatted output containing descriptions of
// the current dabase schema, obtained from postgres. The correct PGHOST,
// PGPORT, PGUSER etc. env variables must be set to run this script.
//
// First CLI argument is an optional filename to write the output to.
func generate(logger *log.Logger, databaseName string) (string, error) {
	// If we are using pg9.6 use it locally since it is faster (CI \o/)
	out, _ := exec.Command("psql", "--version").CombinedOutput()
	if versionRe.Match(out) {
		runIgnoreError("dropdb", dbname)
		defer runIgnoreError("dropdb", dbname)

		return generateInternal(logger, databaseName, "dbname="+dbname, func(cmd ...string) (string, error) {
			c := exec.Command(cmd[0], cmd[1:]...)
			c.Stderr = logger.Writer()
			out, err := c.Output()
			return string(out), err
		})
	}

	logger.Printf("Running PostgreSQL 9.6 in docker since local version is %s", strings.TrimSpace(string(out)))
	if err := exec.Command("docker", "image", "inspect", "postgres:9.6").Run(); err != nil {
		logger.Println("docker pull postgres9.6")
		pull := exec.Command("docker", "pull", "postgres:9.6")
		pull.Stdout = logger.Writer()
		pull.Stderr = logger.Writer()
		if err := pull.Run(); err != nil {
			return "", errors.Wrap(err, "docker pull postgres9.6")
		}
		logger.Println("docker pull complete")
	}
	runIgnoreError("docker", "rm", "--force", dbname)
	server := exec.Command("docker", "run", "--rm", "--name", dbname, "-e", "POSTGRES_HOST_AUTH_METHOD=trust", "-p", "5433:5432", "postgres:9.6")
	if err := server.Start(); err != nil {
		return "", err
	}

	defer func() {
		_ = server.Process.Kill()
		runIgnoreError("docker", "kill", dbname)
		_ = server.Wait()
	}()

	attempts := 0
	for {
		attempts++
		if err := exec.Command("pg_isready", "-U", "postgres", "-d", dbname, "-h", "127.0.0.1", "-p", "5433").Run(); err == nil {
			break
		} else if attempts > 30 {
			return "", errors.Wrap(err, "pg_isready timeout")
		}
		time.Sleep(time.Second)
	}
	return generateInternal(logger, databaseName, "postgres://postgres@127.0.0.1:5433/postgres?dbname="+dbname, func(cmd ...string) (string, error) {
		cmd = append([]string{"exec", "-u", "postgres", dbname}, cmd...)
		c := exec.Command("docker", cmd...)
		c.Stderr = logger.Writer()
		out, err := c.Output()
		return string(out), err
	})
}

func generateInternal(logger *log.Logger, databaseName, dataSource string, run func(cmd ...string) (string, error)) (string, error) {
	if out, err := run("createdb", dbname); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("run: %s", out))
	}

	if err := dbconn.SetupGlobalConnection(dataSource); err != nil {
		return "", errors.Wrap(err, "SetupGlobalConnection")
	}

	if err := dbconn.MigrateDB(dbconn.Global, databaseName); err != nil {
		return "", errors.Wrap(err, "MigrateDB")
	}

	db, err := dbconn.Open(dataSource)
	if err != nil {
		return "", errors.Wrap(err, "Open")
	}
	defer db.Close()

	tables, err := getTables(db)
	if err != nil {
		return "", err
	}

	ch := make(chan string, len(tables))
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
				logger.Println("describe", table)

				doc, err := describeTable(db, table, run)
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

	return strings.Join(docs, "\n"), nil
}

func getTables(db *sql.DB) (tables []string, _ error) {
	// Query names of all public tables and views.
	rows, err := db.Query(`
		SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		UNION
		SELECT table_name FROM information_schema.views WHERE table_schema = 'public' AND table_name != 'pg_stat_statements';
	`)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return tables, nil
}

func describeTable(db *sql.DB, table string, run func(cmd ...string) (string, error)) (string, error) {
	comment, err := getTableComment(db, table)
	if err != nil {
		return "", err
	}

	columnComments, err := getColumnComments(db, table)
	if err != nil {
		return "", err
	}

	// Get postgres "describe table" output.
	out, err := run("psql", "-X", "--quiet", "--dbname", dbname, "-c", fmt.Sprintf("\\d %s", table))
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

	return buf.String(), nil
}

func getTableComment(db *sql.DB, table string) (comment string, _ error) {
	rows, err := db.Query("select obj_description($1::regclass)", table)
	if err != nil {
		return "", errors.Wrap(err, "db.Query")
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
		return nil, errors.Wrap(err, "db.Query")
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

func runIgnoreError(cmd string, args ...string) {
	_ = exec.Command(cmd, args...).Run()
}
