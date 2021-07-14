package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// createNewMigration creates a new up/down migration file pair for the given database and
// returns the names of the new files. If there was an error, the filesystem should remain
// unmodified.
func createNewMigration(database Database, migrationName string) (up, down string, _ error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return "", "", err
	}

	names, err := readFilenamesNamesInDirectory(baseDir)
	if err != nil {
		return "", "", err
	}

	lastMigrationIndex, ok := parseLastMigrationIndex(names)
	if !ok {
		return "", "", errors.New("no previous migrations exist")
	}

	upPath, downPath, err := makeMigrationFilenames(database, lastMigrationIndex+1, migrationName)
	if err != nil {
		return "", "", err
	}

	if err := writeMigrationFiles(upPath, downPath); err != nil {
		return "", "", err
	}

	return upPath, downPath, nil
}

// generateSquashedMigrations generates the content of a migration file pair that contains the contents
// of a database up to a given migration index. This function will launch a daemon Postgres container,
// migrate a fresh database up to the given migration index, then dump and sanitize the contents.
func generateSquashedMigrations(database Database, migrationIndex int) (up, down string, err error) {
	postgresDSN := fmt.Sprintf(
		"postgres://postgres@127.0.0.1:%d/%s?sslmode=disable",
		squasherContainerExposedPort,
		database.Name,
	)

	teardown, err := runPostgresContainer(database.Name)
	if err != nil {
		return "", "", err
	}
	defer func() {
		err = teardown(err)
	}()

	if err := runMigrationsGoto(database, migrationIndex, postgresDSN); err != nil {
		return "", "", err
	}

	upMigration, err := generateSquashedUpMigration(database, postgresDSN)
	if err != nil {
		return "", "", err
	}

	return upMigration, "-- Nothing\n", nil
}

// removeMigrationFilesUpToIndex removes migration files for the given database falling on
// or before the given migration index. This method returns the names of the files that were
// removed.
func removeMigrationFilesUpToIndex(database Database, targetIndex int) ([]string, error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return nil, err
	}

	names, err := readFilenamesNamesInDirectory(baseDir)
	if err != nil {
		return nil, err
	}

	filtered := names[:0]
	for _, name := range names {
		index, ok := parseMigrationIndex(name)
		if !ok {
			continue
		}

		if index <= targetIndex {
			filtered = append(filtered, name)
		}
	}

	for _, name := range filtered {
		if err := os.Remove(filepath.Join(baseDir, name)); err != nil {
			return nil, err
		}
	}

	return filtered, nil
}

// lastMigrationIndexAtCommit returns the index of the last migration for the given database
// available at the given commit. This function returns a false-valued flag if no migrations
// exist at the given commit.
func lastMigrationIndexAtCommit(database Database, commit string) (int, bool, error) {
	migrationsDir := filepath.Join("migrations", database.Name)

	output, err := runGitCmd("ls-tree", "-r", "--name-only", commit, migrationsDir)
	if err != nil {
		return 0, false, err
	}

	lastMigrationIndex, ok := parseLastMigrationIndex(strings.Split(string(output), "\n"))
	return lastMigrationIndex, ok, nil
}

// migrationDirectoryForDatabase returns the directory where migration files are stored for the
// given database.
func migrationDirectoryForDatabase(database Database) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(repoRoot, "migrations", database.Name), nil
}

// makeMigrationFilenames makes a pair of (absolute) paths to migration files with the
// given migration index and name.
func makeMigrationFilenames(database Database, migrationIndex int, migrationName string) (up string, down string, _ error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return "", "", err
	}

	upPath := filepath.Join(baseDir, fmt.Sprintf("%d_%s.up.sql", migrationIndex, migrationName))
	downPath := filepath.Join(baseDir, fmt.Sprintf("%d_%s.down.sql", migrationIndex, migrationName))
	return upPath, downPath, nil
}

// parseMigrationIndex parse a filename and returns the migration index if the filename
// looks like a migration. Each migration filename has the form {unique_id}_{name}.{dir}.sql.
// This function returns a false-valued flag on failure. Leading directories are stripped
// from the input, so a basename or a full path can be supplied.
func parseMigrationIndex(name string) (int, bool) {
	index, err := strconv.Atoi(strings.Split(filepath.Base(name), "_")[0])
	if err != nil {
		return 0, false
	}

	return index, true
}

// Returns the migration name from a filepath (e.g., 12341234_hello.up.sql -> hello).
func parseMigrationName(name string) (string, bool) {
	names := strings.SplitN(filepath.Base(name), "_", 2)
	if len(names) < 2 {
		return "", false
	}

	baseName := names[1]
	migrationName := strings.ReplaceAll(strings.ReplaceAll(baseName, ".down.sql", ""), ".up.sql", "")
	return migrationName, true
}

// parseLastMigrationIndex parses a list of filenames and returns the highest migration
// index available.
func parseLastMigrationIndex(names []string) (int, bool) {
	indices := make([]int, 0, len(names))
	for _, name := range names {
		if index, ok := parseMigrationIndex(name); ok {
			indices = append(indices, index)
		}
	}
	sort.Ints(indices)

	if len(indices) == 0 {
		return 0, false
	}

	return indices[len(indices)-1], true
}

const migrationFileTemplate = `
BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

COMMIT;
`

// writeMigrationFiles writes the contents of migrationFileTemplate to the given filepaths.
func writeMigrationFiles(paths ...string) (err error) {
	defer func() {
		if err != nil {
			for _, path := range paths {
				// undo any changes to the fs on error
				_ = os.Remove(path)
			}
		}
	}()

	for _, path := range paths {
		if err := os.WriteFile(path, []byte(migrationFileTemplate), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

const (
	squasherContainerName        = "squasher"
	squasherContainerExposedPort = 5433
)

// runPostgresContainer runs a postgres:12.6 daemon with an empty db with the given name.
// This method returns a teardown function that filters the error value of the calling
// function, as well as any immediate synchronous error.
func runPostgresContainer(databaseName string) (_ func(err error) error, err error) {
	pending := out.Pending(output.Line("", output.StylePending, "Starting PostgreSQL 12 in a container..."))
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
		if _, killErr := runDockerCmd(killArgs...); killErr != nil {
			err = multierror.Append(err, errors.Errorf("failed to stop docker container: %s", killErr))
		}

		return err
	}

	runArgs := []string{
		"run",
		"--rm", "-d",
		"--name", squasherContainerName,
		"-p", fmt.Sprintf("%d:5432", squasherContainerExposedPort),
		"-e", "POSTGRES_HOST_AUTH_METHOD=trust",
		"postgres:12.6",
	}
	if _, err := runDockerCmd(runArgs...); err != nil {
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
	if _, err := runDockerCmd(execArgs...); err != nil {
		return nil, teardown(err)
	}

	return teardown, nil
}

// runMigrationsGoto runs the `migrate` utility to migrate up or down to the given
// migration index.
func runMigrationsGoto(database Database, migrationIndex int, postgresDSN string) (err error) {
	pending := out.Pending(output.Line("", output.StylePending, "Migrating PostgreSQL schema..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Migrated PostgreSQL schema"))
		} else {
			pending.Destroy()
		}
	}()

	_, err = runMigrate(
		database,
		postgresDSN+fmt.Sprintf("&x-migrations-table=%s", database.MigrationTable),
		"goto", strconv.FormatInt(int64(migrationIndex), 10),
	)
	return err
}

// runMigrationsUp runs the `migrate` utility to migrate up the given number of steps.
// If n is nil then all migrations are ran.
func runMigrationsUp(database Database, n *int) (string, error) {
	args := []string{"up"}
	if n != nil {
		args = append(args, strconv.Itoa(*n))
	}

	return runMigrate(database, makePostgresDSN(database), args...)
}

// runMigrationsDown runs the `migrate` utility to migrate up the given number of steps.
func runMigrationsDown(database Database, n int) (string, error) {
	return runMigrate(database, makePostgresDSN(database), "down", strconv.Itoa(n))
}

// runMigrate runs the migrate utility with the given arguments.
// TODO - replace with our db utilities
func runMigrate(database Database, postgresDSN string, args ...string) (string, error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return "", err
	}

	return runCommandInRoot(exec.Command("migrate", append([]string{"-database", postgresDSN, "-path", baseDir}, args...)...))
}

// makePostgresDSN returns a PostresDSN for the given database. For all databases except
// the default, the PG* environment variables are prefixed with the database name. The
// resulting address depends on the environment.
func makePostgresDSN(database Database) string {
	var prefix string
	if database.Name != defaultDatabase.Name {
		prefix = strings.ToUpper(database.Name) + "_"
	}

	var port string
	if value := os.Getenv(fmt.Sprintf("%sPGPORT", prefix)); value != "" {
		port = ":" + value
	}

	return fmt.Sprintf(
		"postgres://%s%s/%s?x-migrations-table=%s",
		os.Getenv(fmt.Sprintf("%sPGHOST", prefix)),
		port,
		os.Getenv(fmt.Sprintf("%sPGDATABASE", prefix)),
		database.MigrationTable,
	)
}

// generateSquashedUpMigration returns the contents of an up migration file containing the
// current contents of the given database.
func generateSquashedUpMigration(database Database, postgresDSN string) (_ string, err error) {
	pending := out.Pending(output.Line("", output.StylePending, "Dumping current database..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Dumped current database"))
		} else {
			pending.Destroy()
		}
	}()

	pgDump := func(args ...string) (string, error) {
		cmd := exec.Command("pg_dump", append([]string{postgresDSN}, args...)...)
		cmd.Env = []string{}
		return runCommandInRoot(cmd)
	}

	pgDumpOutput, err := pgDump("--schema-only", "--no-owner", "--exclude-table", "*schema_migrations")
	if err != nil {
		return "", err
	}

	for _, table := range database.DataTables {
		dataOutput, err := pgDump("--data-only", "--inserts", "--table", table)
		if err != nil {
			return "", err
		}

		pgDumpOutput += dataOutput
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

	return fmt.Sprintf("BEGIN;\n\n%s\n\nCOMMIT;\n", strings.TrimSpace(filteredContent))
}

// readFilenamesNamesInDirectory returns a list of names in the given directory.
func readFilenamesNamesInDirectory(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}

type migration struct {
	ID   int
	Name string

	UpName   string
	DownName string
}

type migrationConflict struct {
	ID    int
	Main  migration
	Local migration
}

func getMigrationFilesFromGit(database Database, revision string) ([]string, error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return nil, err
	}

	output, err := runGitCmd("ls-tree", "--name-only", "-r", revision, baseDir)
	if err != nil {
		return nil, err
	}
	files := strings.Split(output, "\n")

	return files, nil
}

func getMigrationFilesFromDisk(database Database) ([]string, error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, fileEntry := range entries {
		files = append(files, fileEntry.Name())
	}

	return files, nil
}

func fixupMigrations(database Database, main string) error {
	out.Write("")

	mainFiles, err := getMigrationFilesFromGit(database, main)
	if err != nil {
		return err
	}

	mainMigrations, err := getMigrationsForRevision(mainFiles)
	if err != nil {
		return err
	}

	localFiles, err := getMigrationFilesFromDisk(database)
	if err != nil {
		return err
	}

	localMigrations, err := getMigrationsForRevision(localFiles)
	if err != nil {
		return err
	}

	block := out.Block(output.Linef(output.EmojiLightbulb, output.StyleItalic, "Checking for conflicting migrations..."))
	defer block.Close()

	conflicts, missing, err := findConflictingMigrations(mainMigrations, localMigrations)
	if err != nil {
		return err
	}

	for _, missed := range missing {
		block.WriteLine(output.Linef(
			output.EmojiWarning,
			output.StyleReset,
			"Missing migration '%d_%s' from %s branch. Consider rebasing.", missed.ID, missed.Name, main,
		))
	}

	if len(conflicts) == 0 {
		block.WriteLine(output.Linef(output.EmojiSuccess, output.StyleReset, "... No conflicting migrations"))
		return nil
	}

	if err := resolveConflictingMigrations(database, conflicts, mainMigrations, block); err != nil {
		return err
	}

	return err
}

func findConflictingMigrations(mainMigrations, localMigrations map[int]migration) ([]migrationConflict, []migration, error) {
	conflicts := []migrationConflict{}
	missing := []migration{}

	for migrationID, mainMigration := range mainMigrations {
		localMigration, ok := localMigrations[migrationID]
		if !ok {
			missing = append(missing, mainMigration)
			continue
		}

		if mainMigration.Name != localMigration.Name {
			conflicts = append(conflicts, migrationConflict{
				ID:    migrationID,
				Main:  mainMigration,
				Local: localMigration,
			})
		}
	}

	return conflicts, missing, nil
}

func getMigrationsForRevision(files []string) (map[int]migration, error) {
	upMigrations := make(map[int]string)
	downMigrations := make(map[int]string)

	for _, file := range files {
		if file == "" {
			continue
		}

		migrationID, ok := parseMigrationIndex(file)
		if !ok {
			return nil, errors.Newf("bad migration file format: %s", file)
		}

		if strings.HasSuffix(file, ".down.sql") {
			downMigrations[migrationID] = file
		} else if strings.HasSuffix(file, ".up.sql") {
			upMigrations[migrationID] = file
		} else if strings.HasSuffix(file, ".sql") {
			return nil, errors.Newf("sql file that doesn't fit migration file format: %s", file)
		}
	}

	if len(upMigrations) != len(downMigrations) {
		return nil, errors.Newf(
			"not the same number of up (%d) and down (%d) migrations. Check for corresponding up & down scripts for each migration.",
			len(upMigrations),
			len(downMigrations),
		)
	}

	migrations := make(map[int]migration)
	for migrationID := range upMigrations {
		upMigration, ok := upMigrations[migrationID]
		if !ok {
			return nil, errors.Newf("missing up migration for ID: %d", migrationID)
		}

		downMigration, ok := downMigrations[migrationID]
		if !ok {
			return nil, errors.Newf("missing down migration for ID: %d", migrationID)
		}

		migrationName, ok := parseMigrationName(upMigration)
		if !ok {
			return nil, errors.Newf("bad migration file name: %s", upMigration)
		}

		migrations[migrationID] = migration{
			ID:       migrationID,
			Name:     migrationName,
			UpName:   upMigration,
			DownName: downMigration,
		}

	}

	return migrations, nil
}

func resolveConflictingMigrations(
	database Database,
	conflicts []migrationConflict,
	mainMigrations map[int]migration,
	block *output.Block,
) error {
	block.Writef("Database: %s\n", database.Name)

	maxID := 0
	for migrationID := range mainMigrations {
		maxID = int(math.Max(float64(maxID), float64(migrationID)))
	}

	for _, conflict := range conflicts {
		maxID = maxID + 1

		newUpPath, newDownPath, err := makeMigrationFilenames(database, maxID, conflict.Local.Name)
		if err != nil {
			return err
		}

		oldUpPath, oldDownPath, err := makeMigrationFilenames(database, conflict.ID, conflict.Local.Name)
		if err != nil {
			return err
		}

		block.Writef("Changing migration: %d %s", conflict.ID, conflict.Main.Name)

		// This is a bit annoying, but git ls-tree only checks commited files
		upErr := checkFile(oldUpPath, block)
		downErr := checkFile(oldDownPath, block)

		if upErr != nil || downErr != nil {
			// This should not be possible :) We should have died earlier but it's good to confirm
			return errors.Newf(
				"could not find both migration files for migration (%d): %s || %s",
				conflict.ID,
				oldUpPath,
				oldDownPath,
			)
		}

		if err := moveMigration(oldUpPath, newUpPath, block); err != nil {
			return nil
		}
		if err := moveMigration(oldDownPath, newDownPath, block); err != nil {
			return nil
		}
	}

	return nil
}

func checkFile(path string, block *output.Block) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		relativePath, _ := filepath.Rel(repoRoot, path)

		block.WriteLine(output.Linef(output.EmojiFailure, output.StyleItalic, "File no longer exists: %s", relativePath))
	}

	return err

}

func moveMigration(oldpath, newpath string, block *output.Block) error {
	if err := os.Rename(oldpath, newpath); err != nil {
		return err
	}

	block.WriteLine(output.Linef(output.EmojiSuccess, output.StyleReset, "%s -> %s", filepath.Base(oldpath), filepath.Base(newpath)))
	return nil
}
