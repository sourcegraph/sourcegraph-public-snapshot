// Run with bazel run //testing/tools/upgradetest:sh_upgradetest --config=darwin-docker

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/run"

	"github.com/Masterminds/semver"
)

func main() {

	ctx := context.Background()

	latestMinorVersion, latestVersion, err := getLatestVersions(ctx)
	if err != nil {
		fmt.Println("🚨 Error: could not get latest major or minor releases: ", err)
		os.Exit(1)
	}

	// Get the release candidate image tarball
	args := os.Args
	fmt.Println(args)

	if err := standardUpgradeTest(ctx, latestMinorVersion, latestVersion); err != nil {
		fmt.Println("--- 🚨 Standard Upgrade Test Failed: ", err)
		os.Exit(1)
	}

	if err := multiversionUpgradeTest(ctx); err != nil {
		fmt.Println("--- 🚨 Multiversion Upgrade Test Failed: ", err)
		os.Exit(1)
	}

	if err := autoUpgradeTest(ctx); err != nil {
		fmt.Println("--- 🚨 Auto Upgrade Test Failed: ", err)
		os.Exit(1)
	}
}

// standardUpgradeTest initializes Sourcegraph's dbs and runs a standard upgrade
// i.e. an upgrade test between the last minor version and the current release candidate
func standardUpgradeTest(ctx context.Context, latestMinorVersion, latestVersion *semver.Version) error {
	fmt.Println("--- 🕵️  standard upgrade test")
	//start test env
	initHash, networkName, dbs, cleanup, err := setupTestEnv(ctx, latestMinorVersion)
	if err != nil {
		fmt.Println("failed to setup env: ", err)
		cleanup()
		return err
	}
	defer cleanup()

	// ensure env correctly initialized
	if err := validateDBs(ctx, latestMinorVersion.String(), fmt.Sprintf("sourcegraph/migrator:%s", latestMinorVersion.String()), networkName, dbs, false); err != nil {
		fmt.Println("🚨 Upgrade failed: ", err)
		return err
	}

	fmt.Println("-- ⚙️  performing standard upgrade")

	// create hash for new migrator job, migrator runs in env setup, we want a new hash for each run
	hash, err := newContainerHash()
	if err != nil {
		fmt.Println("🚨 failed to generate hash during standard upgrade test: ", err)
		cleanup()
		return err
	}

	// Run standard upgrade via migrators "up" command
	if err := run.Cmd(ctx,
		dockerMigratorBaseString("up", "migrator:candidate", networkName, initHash, dbs)...).
		Run().Stream(os.Stdout); err != nil {
		fmt.Println("🚨 failed to upgrade: ", err)
		cleanup()
		return err
	}

	// Start frontend with candidate
	var cleanFrontend func()
	cleanFrontend, err = startFrontend(ctx, "frontend", "candidate", networkName, hash, dbs)
	if err != nil {
		fmt.Println("🚨 failed to start candidate frontend: ", err)
		cleanFrontend()
		return err
	}
	defer cleanFrontend()

	fmt.Println("-- ⚙️  post upgrade validation")
	// Validate the upgrade
	if err := validateDBs(ctx, "0.0.0+dev", "migrator:candidate", networkName, dbs, true); err != nil {
		fmt.Println("🚨 Upgrade failed: ", err)
		return err
	}

	return nil
}

func multiversionUpgradeTest(ctx context.Context) error {
	fmt.Println("--- 🕵️  multiversion upgrade test")
	return nil
}

func autoUpgradeTest(ctx context.Context) error {
	fmt.Println("--- 🕵️  auto upgrade test")
	return nil
}

type testDB struct {
	Name              string
	HashName          string
	Image             string
	ContainerHostPort string
}

// Create a docker network for testing as well as instances of our three databases. Returning a cleanup function.
// An instance of Sourcegraph-Frontend is also started to initialize the versions table of the database.
// TODO: setupTestEnv should seed some initial data at the target initVersion. This will be usefull for testing OOB migrations
func setupTestEnv(ctx context.Context, initVersion *semver.Version) (hash []byte, networkName string, dbs []testDB, cleanup func(), err error) {
	fmt.Println("-- 🏗️  setting up test environment")

	// Generate random hash for naming containers in test
	hash, err = newContainerHash()
	if err != nil {
		fmt.Println("🚨 failed to generate random hash for naming containers in test: ")
		return nil, "", nil, nil, err
	}

	dbs = []testDB{
		{"pgsql", fmt.Sprintf("wg_pgsql_%x", hash), "postgres-12-alpine", "5433"},
		{"codeintel-db", fmt.Sprintf("wg_codeintel-db_%x", hash), "codeintel-db", "5434"},
		{"codeinsights-db", fmt.Sprintf("wg_codeinsights-db_%x", hash), "codeinsights-db", "5435"},
	}

	// Create a docker network for testing
	networkName = fmt.Sprintf("wg_test_%x", hash)
	fmt.Println("🐋 creating network", networkName)

	if err := run.Cmd(ctx, "docker", "network", "create", networkName).Run().Wait(); err != nil {
		fmt.Printf("🚨 failed to create test network: %s", err)
		return nil, "", nil, nil, err
	}

	// Here we create the three databases using docker run.
	wgInit := pool.New().WithErrors()
	for _, db := range dbs {
		fmt.Printf("🐋 creating %s\n", db.HashName)
		wgInit.Go(func() error {
			db := db
			cmd := run.Cmd(ctx, "docker", "run", "--rm",
				"--detach",
				"--platform", "linux/amd64",
				"--name", db.HashName,
				"--network", networkName,
				"-p", fmt.Sprintf("%s:5432", db.ContainerHostPort),
				fmt.Sprintf("sourcegraph/%s:%s", db.Image, initVersion),
			)
			return cmd.Run().Wait()
		})
	}
	if err := wgInit.Wait(); err != nil {
		fmt.Printf("🚨 failed to create test databases: %s", err)
	}

	dbPingTimeout, cancel := context.WithTimeout(ctx, time.Second*20)
	wgDbPing := pool.New().WithErrors().WithContext(dbPingTimeout)
	defer cancel()

	// Here we poll/ping the dbs to ensure postgres has initialized before we make calls to the databases.
	for _, db := range dbs {
		db := db // this closure locks the index for the inner for loop
		wgDbPing.Go(func(ctx context.Context) error {
			dbClient, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@localhost:%s/sg?sslmode=disable", db.ContainerHostPort))
			if err != nil {
				fmt.Printf("🚨 failed to connect to %s: %s\n", db.Name, err)
			}
			defer dbClient.Close()
			for {
				select {
				case <-dbPingTimeout.Done():
					return dbPingTimeout.Err()
				default:
				}
				err = dbClient.Ping()
				if err != nil {
					fmt.Printf(" ... pinging %s\n", db.Name)
					if err == sql.ErrConnDone || strings.Contains(err.Error(), "connection refused") {
						return fmt.Errorf("🚨 unrecoverable error pinging %s: %w", db.Name, err)
					}
					time.Sleep(1 * time.Second)
					continue
				} else {
					fmt.Printf("✅ %s is up\n", db.Name)
					return nil
				}
			}
		})
	}
	if err := wgDbPing.Wait(); err != nil {
		fmt.Println("🚨 containerized database startup error: ", err)
	}

	// Initialize the databases by running migrator with the `up` command.
	fmt.Println("-- 🏗️  initializing database schemas with migrator")
	if err := run.Cmd(ctx,
		dockerMigratorBaseString("up", fmt.Sprintf("sourcegraph/migrator:%s", initVersion), networkName, hash, dbs)...).
		Run().Stream(os.Stdout); err != nil {
		fmt.Println("🚨 failed to initialize database: ", err)
	}

	// Verify that the databases are initialized.
	fmt.Println("🔎 checking db schemas initialized")
	for _, db := range dbs {
		dbClient, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@localhost:%s/sg?sslmode=disable", db.ContainerHostPort))
		if err != nil {
			fmt.Printf("🚨 failed to connect to %s: %s\n", db.Name, err)
			continue
		}
		defer dbClient.Close()

		// check if tables have been created
		rows, err := dbClient.Query(`SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public';`)
		if err != nil {
			fmt.Printf("🚨 failed to check %s for init: %s\n", db.Name, err)
			continue
		}
		defer rows.Close()
		if err := rows.Err(); err != nil {
			fmt.Printf("🚨 failed to check %s for init: %s\n", db.Name, err)
			continue
		} else {
			fmt.Printf("✅ %s initialized\n", db.Name)
		}
	}

	//start frontend and poll db until initial version is set
	var cleanFrontend func()
	cleanFrontend, err = startFrontend(ctx, "sourcegraph/frontend", initVersion.String(), networkName, hash, dbs)
	if err != nil {
		fmt.Println("🚨 failed to start frontend: ", err)
	}
	defer cleanFrontend()

	// Return a cleanup function that will remove the containers and network.
	cleanup = func() {
		fmt.Println("🧹 removing database containers")
		if err := run.Cmd(ctx, "docker", "kill",
			fmt.Sprintf("wg_pgsql_%x", hash),
			fmt.Sprintf("wg_codeintel-db_%x", hash),
			fmt.Sprintf("wg_codeinsights-db_%x", hash)).
			Run().Stream(os.Stdout); err != nil {
			fmt.Println("🚨 failed to remove database containers after testing: ", err)
		}
		fmt.Println("🧹 removing testing network")
		if err := run.Cmd(ctx, "docker", "network", "rm", networkName).Run().Stream(os.Stdout); err != nil {
			fmt.Println("🚨 failed to remove test network after testing: ", err)
		}
	}

	fmt.Println("-- 🏗️  setup complete")

	return hash, networkName, dbs, cleanup, err
}

func validateDBs(ctx context.Context, version, migratorImage, networkName string, dbs []testDB, upgrade bool) error {
	fmt.Println("-- ⚙️  validating dbs")

	// Get DB clients
	clients := make(map[string]*sql.DB)
	for _, db := range dbs {
		client, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@localhost:%s/sg?sslmode=disable", db.ContainerHostPort))
		if err != nil {
			fmt.Printf("🚨 failed to connect to %s: %s\n", db.Name, err)
			return err
		}
		defer client.Close()

		clients[db.Name] = client
	}

	// Verify the versions.version value in the frontend db
	fmt.Println("🔎 checking pgsql versions.version set")
	var testVersion string
	row := clients["pgsql"].QueryRowContext(ctx, `SELECT version FROM versions;`)
	if err := row.Scan(&testVersion); err != nil {
		fmt.Println("🚨 failed to get version from pgsql db: ", err)
		return err
	}
	if version != testVersion {
		fmt.Printf("🚨 versions.version not set: %s!= %s\n", version, testVersion)
		return fmt.Errorf("versions.versions not set: %s!= %s", version, testVersion)
	}

	fmt.Println("✅ versions.version set: ", testVersion)

	// Check for any failed migrations in the migration logs tables
	for _, db := range dbs {
		fmt.Printf("🔎 checking %s migration_logs\n", db.HashName)
		var numFailedMigrations int
		row = clients[db.Name].QueryRowContext(ctx, `SELECT count(*) FROM migration_logs WHERE success=false;`)
		if err := row.Scan(&numFailedMigrations); err != nil {
			fmt.Printf("🚨 failed to get failed migrations from %s db: %s\n", db.HashName, err)
			return err
		}
		if numFailedMigrations > 0 {
			fmt.Printf("🚨 failed migrations found: %d\n", numFailedMigrations)
			return fmt.Errorf("failed migrations found: %d", numFailedMigrations)
		}

		fmt.Println("✅ no failed migrations found")
	}

	// Get the last commit in the release branch
	var candidateGitHead bytes.Buffer
	if err := run.Cmd(ctx, "git", "rev-parse", "HEAD").Run().Stream(&candidateGitHead); err != nil {
		fmt.Printf("🚨 failed to get latest commit on candidate branch: %s\n", err)
		return err
	}
	fmt.Println("Latest commit on candidate branch: ", candidateGitHead.String())

	// Check DBs for drift
	fmt.Println("🔎 Checking DBs for drift")
	for _, db := range dbs {
		hash, err := newContainerHash()
		if err != nil {
			fmt.Printf("🚨 failed to get container hash: %s\n", err)
			return err
		}
		if upgrade {
			if err = run.Cmd(ctx, dockerMigratorBaseString(fmt.Sprintf("drift --db %s --version %s --ignore-migrator-update --skip-version-check", db.Name, candidateGitHead.String()),
				migratorImage, networkName, hash, dbs)...).Run().Stream(os.Stdout); err != nil {
				fmt.Printf("🚨 failed to check drift on %s: %s\n", db.Name, err)
				return err
			}
		} else {
			if err = run.Cmd(ctx, dockerMigratorBaseString(fmt.Sprintf("drift --db %s --version v%s --ignore-migrator-update", db.Name, version),
				migratorImage, networkName, hash, dbs)...).Run().Stream(os.Stdout); err != nil {
				fmt.Printf("🚨 failed to check drift on %s: %s\n", db.Name, err)
				return err
			}
		}
	}

	return nil
}

// startFrontend starts the frontend container in the CI test env.
func startFrontend(ctx context.Context, image, version, networkName string, hash []byte, dbs []testDB) (cleanup func(), err error) {
	fmt.Printf("🐋 creating wg_frontend_%x\n", hash)
	cleanup = func() {
		fmt.Println("🧹 removing frontend container")
		if err := run.Cmd(ctx, "docker", "kill",
			fmt.Sprintf("wg_frontend_%x", hash),
		).Run().Stream(os.Stdout); err != nil {
			fmt.Println("🚨 failed to remove frontend after testing: ", err)
		}
	}

	// Start the frontend container
	err = run.Cmd(ctx, "docker", "run",
		"--detach",
		"--platform", "linux/amd64",
		"--name", fmt.Sprintf("wg_frontend_%x", hash),
		"-e", "DEPLOY_TYPE=docker-compose",
		"-e", fmt.Sprintf("PGHOST=%s", dbs[0].HashName),
		"-e", fmt.Sprintf("CODEINTEL_PGHOST=%s", dbs[1].HashName),
		"-e", fmt.Sprintf("CODEINSIGHTS_PGDATASOURCE=postgres://sg@%s:5432/sg?sslmode=disable", dbs[2].HashName),
		"--network", networkName,
		fmt.Sprintf("%s:%s", image, version),
	).Run().Wait()
	if err != nil {
		fmt.Println("🚨 failed to start frontend: ", err)
		return cleanup, err
	}

	// poll db until initial versions.version is set
	setVersionTimeout, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	fmt.Println("🔎 checking pgsql versions.version set")

	dbClient, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@localhost:%s/sg?sslmode=disable", dbs[0].ContainerHostPort))
	if err != nil {
		fmt.Printf("🚨 failed to connect to %s: %s\n", dbs[0].HashName, err)
	}
	defer dbClient.Close()

	for {
		select {
		case <-setVersionTimeout.Done():
			return cleanup, setVersionTimeout.Err()
		default:
		}
		// check version string set
		var dbVersion string
		row := dbClient.QueryRowContext(setVersionTimeout, `SELECT version FROM versions;`)
		err = row.Scan(&dbVersion)
		if err != nil {
			fmt.Printf("... querying versions.version: %s\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
		// wait for the frontend to set the database versions.version value before considering the frontend startup complete.
		// "candidate" resolves to "0.0.0+dev" and should always be valid
		if dbVersion == version || dbVersion == "0.0.0+dev" {
			fmt.Printf("✅ versions.version is set: %s\n", dbVersion)
			break
		}
		if version != dbVersion {
			time.Sleep(1 * time.Second)
			fmt.Println(" ... waiting for versions.version to be set: ", version)
			continue
		}
	}

	return cleanup, nil
}

// dockerMigratorBaseString a slice of strings constituting the necessary arguments to run the migrator via docker container the CI test env.
func dockerMigratorBaseString(cmd, migratorImage, networkName string, migratorHash []byte, dbs []testDB) []string {
	return []string{"docker", "run", "--rm",
		"--platform", "linux/amd64",
		"--name", fmt.Sprintf("wg_migrator_%x", migratorHash),
		// "-e", "SRC_LOG_LEVEL=debug",
		"-e", fmt.Sprintf("PGHOST=%s", dbs[0].HashName),
		"-e", "PGPORT=5432",
		"-e", "PGUSER=sg",
		"-e", "PGPASSWORD=sg",
		"-e", "PGDATABASE=sg",
		"-e", "PGSSLMODE=disable",
		"-e", fmt.Sprintf("CODEINTEL_PGHOST=%s", dbs[1].HashName),
		"-e", "CODEINTEL_PGPORT=5432",
		"-e", "CODEINTEL_PGUSER=sg",
		"-e", "CODEINTEL_PGPASSWORD=sg",
		"-e", "CODEINTEL_PGDATABASE=sg",
		"-e", "CODEINTEL_PGSSLMODE=disable",
		"-e", fmt.Sprintf("CODEINSIGHTS_PGHOST=%s", dbs[2].HashName),
		"-e", "CODEINSIGHTS_PGPORT=5432",
		"-e", "CODEINSIGHTS_PGUSER=sg", // starting codeinsights without frontend initializes with user sg rather than postgres
		"-e", "CODEINSIGHTS_PGPASSWORD=password",
		"-e", "CODEINSIGHTS_PGDATABASE=sg", // starting codeinsights without frontend initializes with database name as sg rather than postgres
		"-e", "CODEINSIGHTS_PGSSLMODE=disable",
		"--network", networkName,
		migratorImage,
		cmd,
	}
}

// Generate random hash for naming containers in test
func newContainerHash() ([]byte, error) {
	hash := make([]byte, 4)
	_, err := rand.Read(hash)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// getLatestVersions returns the latest minor semver version, as well as the latest full semver version.
func getLatestVersions(ctx context.Context) (latestMinor *semver.Version, latestFull *semver.Version, err error) {
	tags, err := run.Cmd(ctx, "git",
		"for-each-ref",
		"--format", "'%(refname:short)'",
		"refs/tags").Run().Lines()
	if err != nil {
		return nil, nil, err
	}

	var latestMinorVer *semver.Version
	var latestFullVer *semver.Version

	for _, tag := range tags {
		v, err := semver.NewVersion(tag)
		if err != nil {
			continue // skip non-matching tags
		}
		// Track latest full version
		if latestFullVer == nil || v.GreaterThan(latestFullVer) {
			latestFullVer = v
		}
		latestMinorVer, err = semver.NewVersion(fmt.Sprintf("%d.%d.0", latestFullVer.Major(), latestFullVer.Minor()))
		if err != nil {
			return nil, nil, err
		}
	}

	if latestMinorVer == nil {
		return nil, nil, errors.New("No valid minor semver tags found")
	}
	if latestFullVer == nil {
		return nil, nil, errors.New("No valid full semver tags found")
	}

	return latestMinorVer, latestFullVer, nil

}
