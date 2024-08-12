package migration

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/regexp"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	squasherContainerName         = "squasher"
	squasherContainerExposedPort  = 5433
	squasherContainerPostgresName = "postgres"
)

func SquashAll(database db.Database, inContainer, runInTimescaleDBContainer, skipTeardown, skipData bool, filepath string) error {
	definitions, err := readDefinitions(database)
	if err != nil {
		return err
	}
	var leafIDs []int
	for _, leaf := range definitions.Leaves() {
		leafIDs = append(leafIDs, leaf.ID)
	}

	squashedUpMigration, _, err := generateSquashedMigrations(database, leafIDs, inContainer, runInTimescaleDBContainer, skipTeardown, skipData)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, []byte(squashedUpMigration), os.ModePerm)
}

func Squash(database db.Database, commit string, inContainer, runInTimescaleDBContainer, skipTeardown, skipData bool) error {
	definitions, err := readDefinitions(database)
	if err != nil {
		return err
	}

	newRoot, ok, err := selectNewRootMigration(database, definitions, commit)
	if err != nil {
		return err
	}
	if !ok {
		return errors.Newf("no migrations exist at commit %s", commit)
	}

	// Run migrations up to the new selected root and dump the database into a single migration file pair
	squashedUpMigration, squashedDownMigration, err := generateSquashedMigrations(database, []int{newRoot.ID}, inContainer, runInTimescaleDBContainer, skipTeardown, skipData)
	if err != nil {
		return err
	}
	privilegedUpMigration, unprivilegedUpMigration := splitPrivilegedMigrations(squashedUpMigration)

	// Add newline after progress related to container
	std.Out.Write("")

	unprivilegedFiles, err := makeMigrationFilenames(database, newRoot.ID, "squashed_migrations_unprivileged")
	if err != nil {
		return err
	}
	// Track the files we're generating so we can list what we changed on disk
	files := []MigrationFiles{
		unprivilegedFiles,
	}

	createMetadata := func(name string, parents []int, privileged bool) string {
		content, _ := yaml.Marshal(struct {
			Name          string `yaml:"name"`
			Parents       []int  `yaml:"parents"`
			Privileged    bool   `yaml:"privileged"`
			NonIdempotent bool   `yaml:"nonIdempotent"`
		}{name, parents, privileged, true})

		return string(content)
	}

	contents := map[string]string{
		unprivilegedFiles.UpFile:       unprivilegedUpMigration,
		unprivilegedFiles.DownFile:     squashedDownMigration,
		unprivilegedFiles.MetadataFile: createMetadata("squashed migrations", nil, false),
	}
	if privilegedUpMigration != "" {
		if len(newRoot.Parents) == 0 {
			return errors.New("select (unprivileged) squash root has no parent; create a new privileged root manually for this schema")
		}

		// We need a deterministic place to put our privileged queries _prior_ to the
		// squashed migration. Naturally, we want to re-use a migration identifier that's
		// already been applied. We'll choose any of the parents of this new squash root
		// and replace its contents.
		privilegedRoot := newRoot.Parents[0]

		privilegedFiles, err := makeMigrationFilenames(database, privilegedRoot, "squashed_migrations_privileged")
		if err != nil {
			return err
		}
		files = append(files, privilegedFiles)

		// Add privileged queries into new migration
		contents[privilegedFiles.UpFile] = privilegedUpMigration
		contents[privilegedFiles.DownFile] = squashedDownMigration
		contents[privilegedFiles.MetadataFile] = createMetadata("squashed migrations (privileged)", nil, true)

		// Update new (unprivileged) root to declare the new privileged root as its parent
		contents[unprivilegedFiles.MetadataFile] = createMetadata("squashed migrations (unprivileged)", []int{privilegedRoot}, false)
	}

	// Remove the migration files that were squashed into a new root
	filenames, err := removeAncestorsOf(database, definitions, newRoot.ID)
	if err != nil {
		return err
	}

	// Write new file back onto disk. We do this after deleting since there might
	// be some overlap (and we don't want to delete what we just wrote to disk).
	if err := writeMigrationFiles(contents); err != nil {
		return err
	}

	block := std.Out.Block(output.Styled(output.StyleBold, "Updated filesystem"))
	defer block.Close()

	for _, filename := range filenames {
		block.Writef("Deleted: %s", rootRelative(filename))
	}

	for _, files := range files {
		block.Writef("Up query file: %s", rootRelative(files.UpFile))
		block.Writef("Down query file: %s", rootRelative(files.DownFile))
		block.Writef("Metadata file: %s", rootRelative(files.MetadataFile))
	}

	return nil
}

// selectNewRootMigration selects the most recently defined migration that dominates the leaf
// migrations of the schema at the given commit. This ensures that whenever we squash migrations,
// we do so between a portion of the graph with a single entry and a single exit, which can
// be easily collapsible into one file that can replace an existing migration node in-place.
func selectNewRootMigration(database db.Database, ds *definition.Definitions, commit string) (definition.Definition, bool, error) {
	migrationsDir := filepath.Join("migrations", database.Name)

	gitCmdOutput, err := run.GitCmd("ls-tree", "-r", "--name-only", commit, migrationsDir)
	if err != nil {
		return definition.Definition{}, false, err
	}

	versionsAtCommit := parseVersions(strings.Split(gitCmdOutput, "\n"), migrationsDir)

	filteredDefinitions, err := ds.Filter(versionsAtCommit)
	if err != nil {
		return definition.Definition{}, false, err
	}

	// Determine the set of parents inside the intersection with children outside of
	// the intersection. Unfortunately it's not enough to calculate only the leaf
	// dominator (below) if there were long-standing PRs that caused a migration parent
	// edge to cross over the release boundary. What we actually need is the dominators
	// of the leaves as well as the set of migrations defined more recently than the
	// squash target version.
	parentsMap := make(map[int]struct{}, len(versionsAtCommit))
	for _, migration := range ds.All() {
		if _, ok := filteredDefinitions.GetByID(migration.ID); !ok {
			for _, parent := range migration.Parents {
				if _, ok := filteredDefinitions.GetByID(parent); ok {
					parentsMap[parent] = struct{}{}
				}
			}
		}
	}
	flattenedParents := make([]int, 0, len(parentsMap))
	for id := range parentsMap {
		flattenedParents = append(flattenedParents, id)
	}

	leafDominator, ok := filteredDefinitions.LeafDominator(flattenedParents...)
	if !ok {
		return definition.Definition{}, false, nil
	}

	return leafDominator, true, nil
}

// generateSquashedMigrations generates the content of a migration file pair that contains the contents
// of a database up to a given migration index.
func generateSquashedMigrations(database db.Database, targetVersions []int, inContainer, runInTimescaleDBContainer, skipTeardown, skipData bool) (up, down string, err error) {
	postgresDSN, teardown, err := setupDatabaseForSquash(database, inContainer, runInTimescaleDBContainer)
	if err != nil {
		return "", "", err
	}
	defer func() {
		if !skipTeardown {
			err = teardown(err)
		}
	}()

	if err := runTargetedUpMigrations(database, targetVersions, postgresDSN); err != nil {
		return "", "", err
	}

	upMigration, err := generateSquashedUpMigration(database, postgresDSN, skipData)
	if err != nil {
		return "", "", err
	}

	return upMigration, "-- Nothing\n", nil
}

// setupDatabaseForSquash prepares a database for use in running a schema up to a certain point so it
// can be reliably dumped. If the provided inContainer flag is true, then this function will launch a
// daemon Postgres container. Otherwise, a new database on the host Postgres instance will be created.
//
// If `runIntimescaleDBContainer` is true, then a TimescaleDB-compatible image will be used. This is
// necessary to squash migrations prior to the deprecation of TimescaleDB.
func setupDatabaseForSquash(database db.Database, runInContainer, runInTimescaleDBContainer bool) (string, func(error) error, error) {
	if runInContainer {
		image := "postgres:12.7"
		if runInTimescaleDBContainer {
			image = "timescale/timescaledb-ha:pg14-latest"
		}

		postgresDSN := fmt.Sprintf(
			"postgres://postgres@127.0.0.1:%d/%s?sslmode=disable",
			squasherContainerExposedPort,
			database.Name,
		)
		teardown, err := runPostgresContainer(image, database.Name)
		return postgresDSN, teardown, err
	}

	databaseName := fmt.Sprintf("sg-squasher-%s", database.Name)
	postgresDSN := fmt.Sprintf(
		"postgres://%s@127.0.0.1:%s/%s?sslmode=disable",
		os.Getenv("PGUSER"),
		os.Getenv("PGPORT"),
		databaseName,
	)
	teardown, err := setupLocalDatabase(databaseName)
	return postgresDSN, teardown, err
}

// runTargetedUpMigrations runs up migration targeting the given versions on the given database instance.
func runTargetedUpMigrations(database db.Database, targetVersions []int, postgresDSN string) (err error) {
	logger := log.Scoped("runTargetedUpMigrations")

	pending := std.Out.Pending(output.Line("", output.StylePending, "Migrating PostgreSQL schema..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Migrated PostgreSQL schema"))
		} else {
			pending.Destroy()
		}
	}()

	var dbs []*sql.DB
	defer func() {
		for _, dbHandle := range dbs {
			_ = dbHandle.Close()
		}
	}()

	dsns := map[string]string{
		database.Name: postgresDSN,
	}
	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		// Stash the databases that are passed to us here. On exit of this function
		// we want to make sure that we close the database connections. They're not
		// able to be used on exit and they will block external commands modifying
		// the target database (such as dropdb on cleanup) as it will be seen as
		// in-use.
		dbs = append(dbs, db)

		return connections.NewStoreShim(store.NewWithDB(&observation.TestContext, db, migrationsTable))
	}

	r, err := connections.RunnerFromDSNs(std.Out.Output, logger.IncreaseLevel("runner", "", log.LevelNone), dsns, "sg", storeFactory)
	if err != nil {
		return err
	}

	ctx := context.Background()

	return r.Run(ctx, runner.Options{
		Operations: []runner.MigrationOperation{
			{
				SchemaName:     database.Name,
				Type:           runner.MigrationOperationTypeTargetedUp,
				TargetVersions: targetVersions,
			},
		},
	})
}

func setupLocalDatabase(databaseName string) (_ func(error) error, err error) {
	pending := std.Out.Pending(output.Line("", output.StylePending, fmt.Sprintf("Creating local Postgres database %s...", databaseName)))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, fmt.Sprintf("Created local Postgres database %s", databaseName)))
		} else {
			pending.Destroy()
		}
	}()

	createLocalDatabase := func() error {
		cmd := exec.Command("createdb", databaseName)
		_, err := run.InRoot(cmd, run.InRootArgs{})
		return err
	}

	dropLocalDatabase := func() error {
		cmd := exec.Command("dropdb", databaseName)
		_, err := run.InRoot(cmd, run.InRootArgs{})
		return err
	}

	// Drop in case it already exists; ignore error
	_ = dropLocalDatabase()

	// Try to create new database
	if err := createLocalDatabase(); err != nil {
		return nil, err
	}

	// Drop database on exit
	teardown := func(err error) error {
		if dropErr := dropLocalDatabase(); dropErr != nil {
			err = errors.Append(err, dropErr)
		}
		return err
	}
	return teardown, nil
}

// runPostgresContainer runs the given Postgres-compatible image with an empty db with the
// given name. This method returns a teardown function that filters the error value of the
// calling function, as well as any immediate synchronous error.
func runPostgresContainer(image, databaseName string) (_ func(err error) error, err error) {
	pending := std.Out.Pending(output.Line("", output.StylePending, "Starting PostgreSQL 12 in a container..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Started PostgreSQL in a container"))
		} else {
			pending.Destroy()
		}
	}()

	teardown := func(err error) error {
		killArgs := []string{
			"kill",
			squasherContainerName,
		}
		if _, killErr := run.DockerCmd(killArgs...); killErr != nil {
			err = errors.Append(err, errors.Newf("failed to stop docker container: %s", killErr))
		}

		return err
	}

	runArgs := []string{
		"run",
		"--rm", "-d",
		"--name", squasherContainerName,
		"-p", fmt.Sprintf("%d:5432", squasherContainerExposedPort),
		"-e", "POSTGRES_HOST_AUTH_METHOD=trust",
		image,
	}
	if _, err := run.DockerCmd(runArgs...); err != nil {
		return nil, err
	}

	// TODO - check health instead
	pending.Write("Waiting for container to start up...")
	time.Sleep(5 * time.Second)
	pending.Write("PostgreSQL is accepting connections")

	execArgs := []string{
		"exec",
		"-u", "postgres",
		squasherContainerName,
		"createdb", databaseName,
	}
	if _, err := run.DockerCmd(execArgs...); err != nil {
		return nil, teardown(err)
	}

	return teardown, nil
}

// generateSquashedUpMigration returns the contents of an up migration file containing the
// current contents of the given database.
func generateSquashedUpMigration(database db.Database, postgresDSN string, skipData bool) (_ string, err error) {
	pending := std.Out.Pending(output.Line("", output.StylePending, "Dumping current database..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Dumped current database"))
		} else {
			pending.Destroy()
		}
	}()

	pgDump := func(args ...string) (string, error) {
		cmd := exec.Command("pg_dump", append([]string{postgresDSN}, args...)...)
		var b bytes.Buffer
		err := run.SplitOutputInRoot(cmd, &b, os.Stderr)
		return b.String(), err
	}

	excludeTables := []string{
		"*schema_migrations",
		"migration_logs",
		"migration_logs_id_seq",
	}

	args := []string{
		"--schema-only",
		"--no-owner",
		"--verbose",
	}
	for _, tableName := range excludeTables {
		args = append(args, "--exclude-table", tableName)
	}

	pgDumpOutput, err := pgDump(args...)
	if err != nil {
		return "", err
	}

	if !skipData {
		for _, table := range database.DataTables {
			dataOutput, err := pgDump("--data-only", "--inserts", "--table", table)
			if err != nil {
				return "", err
			}

			pgDumpOutput += dataOutput
		}

		for _, table := range database.CountTables {
			pgDumpOutput += fmt.Sprintf("INSERT INTO %s VALUES (0);\n", table)
		}
	}

	return sanitizePgDumpOutput(pgDumpOutput), nil
}

var (
	migrationDumpRemovePrefixes = []string{
		"--",                                    // remove comments
		"SET ",                                  // remove settings header
		"SELECT pg_catalog.set_config",          // remove settings header
		`could not find a "pg_dump" to execute`, // remove common warning from docker container
		"DROP EXTENSION ",                       // do not drop extensions if they already exist
	}

	migrationDumpRemovePatterns = map[*regexp.Regexp]string{
		regexp.MustCompile(`\bpublic\.`):              "",
		regexp.MustCompile(`\s*WITH SCHEMA public\b`): "",
		regexp.MustCompile(`\n{3,}`):                  "\n\n",
	}
)

// sanitizePgDumpOutput sanitizes the output of pg_dump and wraps the content in a
// transaction block to fit the style of our other migrations.
func sanitizePgDumpOutput(content string) string {
	lines := strings.Split(content, "\n")

	filtered := lines[:0]
outer:
	for _, line := range lines {
		for _, prefix := range migrationDumpRemovePrefixes {
			if strings.HasPrefix(line, prefix) {
				continue outer
			}
		}

		filtered = append(filtered, line)
	}

	filteredContent := strings.Join(filtered, "\n")
	for pattern, replacement := range migrationDumpRemovePatterns {
		filteredContent = pattern.ReplaceAllString(filteredContent, replacement)
	}

	return strings.TrimSpace(filteredContent)
}

var privilegedQueryPattern = lazyregexp.New(`(CREATE|COMMENT ON) EXTENSION .+;\n*`)

// splitPrivilegedMigrations extracts the portion of the squashed migration file that must be run by
// a user with elevated privileges. Both parts of the migration are returned. THe privileged migration
// section is empty when there are no privileged queries.
//
// Currently, we consider the following query patterns as privileged from pg_dump output:
//
//   - CREATE EXTENSION ...
//   - COMMENT ON EXTENSION ...
func splitPrivilegedMigrations(content string) (privilegedMigration string, unprivilegedMigration string) {
	var privilegedQueries []string
	unprivileged := privilegedQueryPattern.ReplaceAllStringFunc(content, func(s string) string {
		privilegedQueries = append(privilegedQueries, s)
		return ""
	})

	return strings.TrimSpace(strings.Join(privilegedQueries, "")) + "\n", unprivileged
}

// removeAncestorsOf removes all migrations that are an ancestor of the given target version.
// This method returns the names of the files that were removed.
func removeAncestorsOf(database db.Database, ds *definition.Definitions, targetVersion int) ([]string, error) {
	allDefinitions := ds.All()

	allIDs := make([]int, 0, len(allDefinitions))
	for _, def := range allDefinitions {
		allIDs = append(allIDs, def.ID)
	}

	properDescendants, err := ds.Down(allIDs, []int{targetVersion})
	if err != nil {
		return nil, err
	}

	keep := make(map[int]struct{}, len(properDescendants))
	for _, def := range properDescendants {
		keep[def.ID] = struct{}{}
	}

	// Gather the set of filtered that are NOT a proper descendant of the given target version.
	// This will leave us with the ancestors of the target version (including itself).
	filteredIDs := make([]int, 0, len(allDefinitions))
	for _, def := range allDefinitions {
		if _, ok := keep[def.ID]; !ok {
			filteredIDs = append(filteredIDs, def.ID)
		}
	}

	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	idsToFilenames := map[int]string{}
	for _, e := range entries {
		if id, err := strconv.Atoi(strings.Split(e.Name(), "_")[0]); err == nil {
			idsToFilenames[id] = e.Name()
		}
	}

	filenames := make([]string, 0, len(filteredIDs))
	for _, id := range filteredIDs {
		filename, ok := idsToFilenames[id]
		if !ok {
			return nil, errors.Newf("could not find file for migration %d", id)
		}

		filenames = append(filenames, filepath.Join(baseDir, filename))
	}

	for _, filename := range filenames {
		if err := os.RemoveAll(filename); err != nil {
			return nil, err
		}
	}

	return filenames, nil
}
