package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/regexp"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	migrationstore "github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type runFunc func(quiet bool, cmd ...string) (string, error)

const databaseNamePrefix = "schemadoc-gen-temp-"

const containerName = "schemadoc"

var logger = log.New(os.Stderr, "", log.LstdFlags)

var versionRe = lazyregexp.New(`\b12\.\d+\b`)

type databaseFactory func(dsn string, appName string, observationContext *observation.Context) (*sql.DB, error)

var schemas = map[string]struct {
	destinationFilename string
	factory             databaseFactory
}{
	"frontend":  {"schema.md", connections.MigrateNewFrontendDB},
	"codeintel": {"schema.codeintel.md", connections.MigrateNewCodeIntelDB},
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
	// // Run pg12 locally if it exists
	// if version, _ := exec.Command("psql", "--version").CombinedOutput(); versionRe.Match(version) {
	return mainLocal()
	// }

	// return mainContainer()
}

func mainLocal() error {
	dataSourcePrefix := "dbname=" + databaseNamePrefix

	g, _ := errgroup.WithContext(context.Background())
	for name, schema := range schemas {
		name, schema := name, schema
		g.Go(func() error {
			return generateAndWrite(name, schema.factory, dataSourcePrefix+name, nil, schema.destinationFilename)
		})
	}

	return g.Wait()
}

func mainContainer() error {
	logger.Printf("Running PostgreSQL 12 in docker")

	prefix, shutdown, err := startDocker()
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown()

	dataSourcePrefix := "postgres://postgres@127.0.0.1:5433/postgres?dbname=" + databaseNamePrefix

	g, _ := errgroup.WithContext(context.Background())
	for name, schema := range schemas {
		name, schema := name, schema
		g.Go(func() error {
			return generateAndWrite(name, schema.factory, dataSourcePrefix+name, prefix, schema.destinationFilename)
		})
	}

	return g.Wait()
}

func generateAndWrite(name string, factory databaseFactory, dataSource string, commandPrefix []string, destinationFile string) error {
	run := runWithPrefix(commandPrefix)

	// Try to drop a database if it already exists
	_, _ = run(true, "dropdb", databaseNamePrefix+name)

	// Let's also try to clean up after ourselves
	defer func() { _, _ = run(true, "dropdb", databaseNamePrefix+name) }()

	if out, err := run(false, "createdb", databaseNamePrefix+name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("run: %s", out))
	}

	db, err := factory(dataSource, "", &observation.TestContext)
	if err != nil {
		return err
	}
	defer db.Close()

	out, err := generateInternalNew(db, name, run)
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

func generateInternalOld(db *sql.DB, name string, run runFunc) (_ string, err error) {
	tables, err := getTables(db)
	if err != nil {
		return "", err
	}

	tableDescriptions, err := fetchTableAndViewDescriptions(name, run)
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

				doc, err := describeTable(db, name, table, tableDescriptions)
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

var nameRe = regexp.MustCompile(`^\s*(Table|View) "public.(?P<name>\w+)"`)

func fetchTableAndViewDescriptions(databaseName string, run runFunc) (map[string]string, error) {
	out, err := run(false, "psql", "-X", "--quiet", "--dbname", databaseNamePrefix+databaseName, "-c", `\d *`)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)

	objectDescriptions := strings.SplitAfter(out, "\n\n")
	for _, objectDescription := range objectDescriptions {
		if match := nameRe.FindStringSubmatch(objectDescription); match != nil {
			res[match[2]] = objectDescription
		}
	}

	return res, nil
}

func describeTable(db *sql.DB, databaseName string, table table, tableDescriptions map[string]string) (string, error) {
	comment, err := getTableComment(db, table.name)
	if err != nil {
		return "", err
	}

	columnComments, err := getColumnComments(db, table.name)
	if err != nil {
		return "", err
	}

	// Get postgres "describe table" output.
	out, ok := tableDescriptions[table.name]
	if !ok {
		return "", errors.Errorf("no description found for table %v", table)
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

func generateInternalNew(db *sql.DB, name string, run runFunc) (_ string, err error) {
	store := migrationstore.NewWithDB(db, "schema_migrations", migrationstore.NewOperations(&observation.TestContext))
	schemas, err := store.Describe(context.Background())
	if err != nil {
		return "", err
	}

	docs := []string{}
	for schemaName, schema := range schemas {
		sortTables(schema.Tables)
		for _, table := range schema.Tables {
			docs = append(docs, formatTable(schemaName, schema, table)...)
		}
	}

	for schemaName, schema := range schemas {
		sortViews(schema.Views)
		for _, view := range schema.Views {
			docs = append(docs, formatView(schemaName, schema, view)...)
		}
	}

	types := map[string][]string{}
	for _, schema := range schemas {
		for _, enum := range schema.Enums {
			types[enum.Name] = enum.Labels
		}
	}
	if len(types) > 0 {
		docs = append(docs, formatTypes(types)...)
	}

	return strings.Join(docs, "\n"), nil
}

func formatTable(schemaName string, description descriptions.SchemaDescription, table descriptions.TableDescription) []string {
	docs := []string{}
	docs = append(docs, fmt.Sprintf("# Table \"%s.%s\"", schemaName, table.Name))
	docs = append(docs, "```")

	headers := []string{"Column", "Type", "Collation", "Nullable", "Default"}
	rows := [][]string{}
	sortColumns(table.Columns)
	for _, column := range table.Columns {
		nullConstraint := "not null"
		if column.IsNullable {
			nullConstraint = ""
		}

		defaultValue := column.Default
		if column.IsGenerated == "ALWAYS" {
			defaultValue = "generated always as (" + column.GenerationExpression + ") stored"
		}

		rows = append(rows, []string{
			column.Name,
			column.TypeName,
			"",
			nullConstraint,
			defaultValue,
		})
	}
	docs = append(docs, formatColumns(headers, rows)...)

	if len(table.Indexes) > 0 {
		docs = append(docs, "Indexes:")
		sortIndexes(table.Indexes)
		for _, index := range table.Indexes {
			if index.IsPrimaryKey {
				def := strings.TrimSpace(strings.Split(index.IndexDefinition, "USING")[1])
				docs = append(docs, fmt.Sprintf("    %q PRIMARY KEY, %s", index.Name, def))
			}
		}
		for _, index := range table.Indexes {
			if !index.IsPrimaryKey {
				uq := ""
				if index.IsUnique {
					c := ""
					if index.ConstraintType == "u" {
						c = " CONSTRAINT"
					}
					uq = " UNIQUE" + c + ","
				}
				deferrable := ""
				if index.IsDeferrable {
					deferrable = " DEFERRABLE"
				}
				def := strings.TrimSpace(strings.Split(index.IndexDefinition, "USING")[1])
				if index.IsExclusion {
					def = "EXCLUDE USING " + def
				}
				docs = append(docs, fmt.Sprintf("    %q%s %s%s", index.Name, uq, def, deferrable))
			}
		}
	}

	numCheckConstraints := 0
	numForeignKeyConstraints := 0
	for _, constraint := range table.Constraints {
		switch constraint.ConstraintType {
		case "c":
			numCheckConstraints++
		case "f":
			numForeignKeyConstraints++
		}
	}

	if numCheckConstraints > 0 {
		docs = append(docs, "Check constraints:")
		for _, constraint := range table.Constraints {
			if constraint.ConstraintType == "c" {
				docs = append(docs, fmt.Sprintf("    %q %s", constraint.Name, constraint.ConstraintDefinition))
			}
		}
	}
	if numForeignKeyConstraints > 0 {
		docs = append(docs, "Foreign-key constraints:")
		for _, constraint := range table.Constraints {
			if constraint.ConstraintType == "f" {
				docs = append(docs, fmt.Sprintf("    %q %s", constraint.Name, constraint.ConstraintDefinition))
			}
		}
	}

	type tableAndConstraint struct {
		descriptions.TableDescription
		descriptions.ConstraintDescription
	}
	tableAndConstraints := []tableAndConstraint{}
	for _, otherTable := range description.Tables {
		for _, constraint := range otherTable.Constraints {
			if constraint.RefTableName == table.Name {
				tableAndConstraints = append(tableAndConstraints, tableAndConstraint{otherTable, constraint})
			}
		}
	}
	sort.Slice(tableAndConstraints, func(i, j int) bool {
		return tableAndConstraints[i].ConstraintDescription.Name < tableAndConstraints[j].ConstraintDescription.Name
	})
	if len(tableAndConstraints) > 0 {
		docs = append(docs, "Referenced by:")

		for _, tableAndConstraint := range tableAndConstraints {
			docs = append(docs, fmt.Sprintf("    TABLE %q CONSTRAINT %q %s", tableAndConstraint.TableDescription.Name, tableAndConstraint.ConstraintDescription.Name, tableAndConstraint.ConstraintDescription.ConstraintDefinition))
		}
	}

	if len(table.Triggers) > 0 {
		docs = append(docs, "Triggers:")
		for _, trigger := range table.Triggers {
			def := strings.TrimSpace(strings.SplitN(trigger.Definition, trigger.Name, 2)[1])
			docs = append(docs, fmt.Sprintf("    %s %s", trigger.Name, def))
		}
	}

	docs = append(docs, "\n```\n")

	if table.Comment != "" {
		docs = append(docs, table.Comment+"\n")
	}

	sortColumnsByName(table.Columns)
	for _, column := range table.Columns {
		if column.Comment != "" {
			docs = append(docs, fmt.Sprintf("**%s**: %s\n", column.Name, column.Comment))
		}
	}

	return docs
}

func formatView(schemaName string, description descriptions.SchemaDescription, view descriptions.ViewDescription) []string {
	docs := []string{}
	docs = append(docs, fmt.Sprintf("# View \"public.%s\"\n", view.Name))
	docs = append(docs, fmt.Sprintf("## View query:\n\n```sql\n%s\n```\n", view.Definition))
	return docs
}

func sortTables(tables []descriptions.TableDescription) {
	sort.Slice(tables, func(i, j int) bool { return tables[i].Name < tables[j].Name })
}

func sortViews(views []descriptions.ViewDescription) {
	sort.Slice(views, func(i, j int) bool { return views[i].Name < views[j].Name })
}

func sortColumns(columns []descriptions.ColumnDescription) {
	sort.Slice(columns, func(i, j int) bool { return columns[i].Index < columns[j].Index })
}

func sortColumnsByName(columns []descriptions.ColumnDescription) {
	sort.Slice(columns, func(i, j int) bool { return columns[i].Name < columns[j].Name })
}

func sortIndexes(indexes []descriptions.IndexDescription) {
	sort.Slice(indexes, func(i, j int) bool {
		if indexes[i].IsUnique && !indexes[j].IsUnique {
			return true
		}
		if !indexes[i].IsUnique && indexes[j].IsUnique {
			return false
		}
		return indexes[i].Name < indexes[j].Name
	})
}

func formatTypes(types map[string][]string) []string {
	typeNames := make([]string, 0, len(types))
	for typeName := range types {
		typeNames = append(typeNames, typeName)
	}
	sort.Strings(typeNames)

	docs := make([]string, 0, len(typeNames)*4)
	for _, name := range typeNames {
		docs = append(docs, fmt.Sprintf("# Type %s", name))
		docs = append(docs, "")
		docs = append(docs, "- "+strings.Join(types[name], "\n- "))
		docs = append(docs, "")
	}

	return docs
}

func formatColumns(headers []string, rows [][]string) []string {
	sizes := make([]int, len(headers))
	headerValues := make([]string, len(headers))
	sepValues := make([]string, len(headers))
	for i, header := range headers {
		sizes[i] = len(header)

		for _, row := range rows {
			if n := len(row[i]); n > sizes[i] {
				sizes[i] = n
			}
		}

		headerValues[i] = centerString(headers[i], sizes[i]+2)
		sepValues[i] = strings.Repeat("-", sizes[i]+2)
	}

	docs := make([]string, 0, len(rows)+2)
	docs = append(docs, strings.Join(headerValues, "|"))
	docs = append(docs, strings.Join(sepValues, "+"))

	for _, row := range rows {
		rowValues := make([]string, 0, len(headers))
		for i := range headers {
			if i == len(headers)-1 {
				rowValues = append(rowValues, row[i])
			} else {
				rowValues = append(rowValues, fmt.Sprintf("%-"+strconv.Itoa(sizes[i])+"s", row[i]))
			}
		}

		docs = append(docs, " "+strings.Join(rowValues, " | "))
	}

	return docs
}

func centerString(s string, n int) string {
	x := float64(n - len(s))
	i := int(math.Floor(x / 2))
	if i <= 0 {
		i = 1
	}
	j := int(math.Ceil(x / 2))
	if j <= 0 {
		j = 1
	}

	return strings.Repeat(" ", i) + s + strings.Repeat(" ", j)
}
