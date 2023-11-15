// Run with bazel run //testing/tools/upgradetest:sh_upgradetest --config=darwin-docker

package main

import (
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
	if len(args) < 2 {
		fmt.Println("🚨 Error: release candidate image not provided")
		os.Exit(1)
	}

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
	if err := validateUpgrade(ctx, latestMinorVersion, dbs); err != nil {
		fmt.Println("🚨 Upgrade failed: ", err)
		cleanup()
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
		dockerMigratorBaseString("migrator:candidate", networkName, "up", initHash, hash)...).
		Run().Stream(os.Stdout); err != nil {
		fmt.Println("🚨 failed to upgrade: ", err)
		cleanup()
		return err
	}
	// Validate the upgrade
	if err := validateUpgrade(ctx, latestMinorVersion, dbs); err != nil {
		fmt.Println("🚨 Upgrade failed: ", err)
		cleanup()
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

// Create a docker network for testing as well as instances of our three databases. Return a cleanup function.
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
		dockerMigratorBaseString(fmt.Sprintf("sourcegraph/migrator:%s", initVersion), networkName, "up", hash, hash)...).
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
	cleanFrontend, err = startFrontend(ctx, initVersion, networkName, hash, dbs)
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

func validateUpgrade(ctx context.Context, initVersion *semver.Version, dbs []testDB) error {
	fmt.Println("-- 🔎  validating dbs")
	fmt.Println("🔎 checking dbs for drift")

	// Check that pgsql versions version has been updated
	fmt.Println("🔎 checking pgsql version")
	dbClient, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@localhost:%s/sg?sslmode=disable", dbs[0].ContainerHostPort))
	if err != nil {
		fmt.Printf("🚨 failed to connect to %s: %s\n", dbs[0].Name, err)
		return err
	}
	defer dbClient.Close()
	rows, err := dbClient.Query(`SELECT version FROM versions;`)
	if err != nil {
		fmt.Println("🚨 failed to get version from pgsql db: ", err)
		return err
	}
	var version string
	if rows.Next() {
		err = rows.Scan(&version)
		if err != nil {
			fmt.Println("🚨 failed to get version from pgsql db: ", err)
			return err
		}
	} else {
		fmt.Println("🚨 failed to get version from pgsql db: no version found")
		return err
	}
	fmt.Println("version from pgsql: ", version)
	defer rows.Close()

	return nil
}

// startFrontend starts the frontend container in the CI test env.
func startFrontend(ctx context.Context, version *semver.Version, networkName string, hash []byte, dbs []testDB) (cleanup func(), err error) {
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
		fmt.Sprintf("sourcegraph/frontend:%s", version),
	).Run().Wait()
	if err != nil {
		fmt.Println("🚨 failed to start frontend: ", err)
		return cleanup, err
	}
	// time.Sleep(time.Second * 10)

	// poll db until initial versions.version is set
	setVersionTimeout, cancel := context.WithTimeout(ctx, time.Second*20)
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
		if version.String() == dbVersion {
			fmt.Printf("✅ versions.version is set: %s\n", version)
			break
		}
		if version.String() == "" {
			time.Sleep(1 * time.Second)
			fmt.Println(" ... waiting for versions.version to be set")
			continue
		}
	}

	return cleanup, nil
}

// dockerMigratorBaseString a slice of strings constituting the necessary arguments to run the migrator via docker container the CI test env.
func dockerMigratorBaseString(migratorImage, networkName, cmd string, initHash, migratorHash []byte) []string {
	return []string{"docker", "run", "--rm",
		"--platform", "linux/amd64",
		"--name", fmt.Sprintf("wg_migrator_%x", migratorHash),
		"-e", fmt.Sprintf("PGHOST=wg_pgsql_%x", initHash),
		"-e", "PGPORT=5432",
		"-e", "PGUSER=sg",
		"-e", "PGPASSWORD=sg",
		"-e", "PGDATABASE=sg",
		"-e", "PGSSLMODE=disable",
		"-e", fmt.Sprintf("CODEINTEL_PGHOST=wg_codeintel-db_%x", initHash),
		"-e", "CODEINTEL_PGPORT=5432",
		"-e", "CODEINTEL_PGUSER=sg",
		"-e", "CODEINTEL_PGPASSWORD=sg",
		"-e", "CODEINTEL_PGDATABASE=sg",
		"-e", "CODEINTEL_PGSSLMODE=disable",
		"-e", fmt.Sprintf("CODEINSIGHTS_PGHOST=wg_codeinsights-db_%x", initHash),
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
