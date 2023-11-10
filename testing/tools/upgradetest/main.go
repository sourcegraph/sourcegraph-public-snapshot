// Init DBs and network
// Seed DBs with test data

/* Have a unit-like/integration test that can use a test db + in-process code (doesn’t require docker), but will end up modifying the git state with new tags when it runs
Use internal/database/dbtesting to create a cheap new (empty) database
Create a new versioned git tag on the current commit, modify the min/max versions constants that modify the stitched migration graph (link above), and run go-generate in that directory to regenerate the json file. If you call test code or build the migrator after this, it will include the tag. Also note that if you just create a tag like this (but don’t push it to remote) then you won’t have to modify any code where it assumes versions are accessible via git (second link above).
Call the internal/database/migration/cliutil shims directly for the unit test (you might want to expose some simpler core you can call for testing that doesn’t do stuff like flag parsing)
You can inspect the databases directly at this point (post-upgrade maybe check that there’s zero drift as the assertion?)
*/

// Run with bazel run //testing/tools/upgradetest:sh_upgradetest --config=darwin-docker

package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/docker/docker/client"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/run"
)

func main() {

	ctx := context.Background()

	// Initialize docker client for tests
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Get the release candidate image tarball
	args := os.Args
	if len(args) < 2 {
		fmt.Println("--- 🚨 Error: release candidate image not provided")
		os.Exit(1)
	}
	imageTarball := args[1]
	fmt.Println(imageTarball)

	// debugging
	run.Cmd(ctx, "docker", "info").Run().Stream(os.Stdout)
	run.Cmd(ctx, "env").Run().Stream(os.Stdout)

	if err := standardUpgradeTest(ctx, cli); err != nil {
		fmt.Println("--- 🚨 Standard Upgrade Test Failed: ", err)
		os.Exit(1)
	}

	multiversionUpgradeTest(ctx)
	autoUpgradeTest(ctx)
}

func standardUpgradeTest(ctx context.Context, cli *client.Client) error {
	fmt.Println("--- 🕵️  standard upgrade test")
	cleanup, err := setupTestEnv(ctx, "5.2.0", cli)
	if err != nil {
		fmt.Println("failed to setup env: ", err)
		return err
	}
	defer cleanup()
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

// Create a docker network for testing as well as instances of our three databases. Return a cleanup function.
func setupTestEnv(ctx context.Context, initVersion string, cli *client.Client) (cleanup func(), err error) {
	fmt.Println("--- 🏗️  setting up test environment")

	type testDB struct {
		Name             string
		User             string
		Image            string
		CreateMessage    string
		LocalPort        string
		ConnectionString string
	}

	// Generate random hash for naming containers in test
	hash := make([]byte, 4)
	_, err = rand.Read(hash)
	if err != nil {
		return nil, err
	}

	// TODO: current connection strings are hardcoded examples to illustrate the postgres connection protocol.
	dbs := []testDB{
		{"pgsql", "sg", "postgres-12-alpine", "--- 🐋 creating pgsql", "5432", ""},
		{"codeintel-db", "sg", "codeintel-db", "--- 🐋 creating codeintel-db", "5433", ""},
		{"codeinsights-db", "sg", "codeinsights-db", "--- 🐋 creating codeinsights-db", "5434", ""},
	}

	// Create a docker network for testing
	networkName := fmt.Sprintf("wg_test_%x", hash)
	fmt.Println("--- 🐋 creating network", networkName)

	if err := run.Cmd(ctx, "docker", "network", "create", networkName).Run().Wait(); err != nil {
		fmt.Printf("🚨 failed to create test network: %s", err)
		return nil, err
	}

	// TODO start doing things via the docker API
	// _, err = cli.NetworkCreate(ctx, networkName, types.NetworkCreate{
	// 	Driver: "bridge",
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// Here we create the three databases using docker run.
	wgInit := pool.New().WithErrors()
	for _, db := range dbs {
		fmt.Println(db.CreateMessage)
		wgInit.Go(func() error {
			db := db
			cmd := run.Cmd(ctx, "docker", "run", "--rm",
				"--detach",
				"--platform", "linux/amd64",
				"--name", fmt.Sprintf("wg_%s_%x", db.Name, hash),
				"--network", networkName,
				fmt.Sprintf("sourcegraph/%s:%s", db.Image, initVersion),
			)
			return cmd.Run().Wait()
		})
	}
	if err := wgInit.Wait(); err != nil {
		fmt.Printf("🚨 failed to create test databases: %s", err)
	}

	timeout, cancel := context.WithTimeout(ctx, time.Second*5)
	wgPing := pool.New().WithErrors().WithContext(timeout)
	defer cancel()

	// Here we poll/ping the dbs to ensure postgres has initialized before we make calls to the databases.
	// TODO: I think I need to use the docker client go package to poll the containers, but I'll need to
	// set the client up to get the db connection schemes. I'll need to do some research on that.
	for _, db := range dbs {
		db := db // this closure locks the index for the inner for loop
		wgPing.Go(func(ctx context.Context) error {
			fmt.Println(" --- ", db.Name, " --- ")
			dbClient, err := sql.Open("postgres", db.ConnectionString)
			if err != nil {
				fmt.Printf("🚨 failed to connect to %s: %s\n", db.Name, err)
			}
			defer dbClient.Close()
			for {
				select {
				case <-timeout.Done():
					return timeout.Err()
				default:
				}
				err = dbClient.Ping()
				if err != nil {
					fmt.Printf("🚨 failed to ping %s: %s\n", db.Name, err)
					time.Sleep(1 * time.Second)
					continue
				} else {
					fmt.Printf("--- ✅ %s is up\n", db.Name)
					return nil
				}
			}
		})
	}
	if err := wgPing.Wait(); err != nil {
		fmt.Println("🚨 failed to ping test databases: ", err)
	}

	// time.Sleep(time.Second * 5)

	// initialize the database and schema
	fmt.Println("--- 🏗️  initializing database schemas with migrator")
	if err := run.Cmd(ctx, dockerMigratorBaseString("sourcegraph/migrator:5.2.0", networkName, "up", hash)...).Run().Stream(os.Stdout); err != nil {
		fmt.Println("--- 🚨 failed to initialize database: ", err)
	}

	for _, db := range dbs {
		fmt.Printf("--- 🔎 checking %s for init\n", db.Name)
		if err := run.Cmd(ctx, "docker", "exec",
			fmt.Sprintf("wg_%s_%x", db.Name, hash),
			"psql", "-U", db.User, "-c", "'\\dt'").
			Run().Stream(os.Stdout); err != nil {
			fmt.Printf("--- 🚨 %s not initialized: %s\n", db.Name, err)
		}
	}

	cleanup = func() {
		fmt.Println("--- 🧹 removing database containers")
		if err := run.Cmd(ctx, "docker", "kill",
			fmt.Sprintf("wg_pgsql_%x", hash),
			fmt.Sprintf("wg_codeintel-db_%x", hash),
			fmt.Sprintf("wg_codeinsights-db_%x", hash)).
			Run().Stream(os.Stdout); err != nil {
			fmt.Println("--- 🚨 failed to remove database containers after testing: ", err)
		}
		fmt.Println("--- 🧹 removing testing network")
		if err := run.Cmd(ctx, "docker", "network", "rm", networkName).Run().Stream(os.Stdout); err != nil {
			fmt.Println("--- 🚨 failed to remove test network after testing: ", err)
		}
	}

	return cleanup, err
}

// dockerMigratorBaseString a slice of strings constituting the necessary arguments to run the migrator via docker container the CI test env.
func dockerMigratorBaseString(migratorImage, networkName, cmd string, hash []byte) []string {
	return []string{"docker", "run", "--rm",
		"--platform", "linux/amd64",
		"--name", fmt.Sprintf("wg_migrator_%x", hash),
		"-e", fmt.Sprintf("PGHOST=wg_pgsql_%x", hash),
		"-e", "PGPORT=5432",
		"-e", "PGUSER=sg",
		"-e", "PGPASSWORD=sg",
		"-e", "PGDATABASE=sg",
		"-e", "PGSSLMODE=disable",
		"-e", fmt.Sprintf("CODEINTEL_PGHOST=wg_codeintel-db_%x", hash),
		"-e", "CODEINTEL_PGPORT=5432",
		"-e", "CODEINTEL_PGUSER=sg",
		"-e", "CODEINTEL_PGPASSWORD=sg",
		"-e", "CODEINTEL_PGDATABASE=sg",
		"-e", "CODEINTEL_PGSSLMODE=disable",
		"-e", fmt.Sprintf("CODEINSIGHTS_PGHOST=wg_codeinsights-db_%x", hash),
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

// Main

// standardUpgrade: runs a standard upgrade with run from the last version to the release candidate, checks for drift
// multiversionUpgrade: runs a multiversion upgrade from some version defined at each last Major release, to the current release candidate
// - might make sense to do only a few random versions here from previous major releases
// autoUpgrade: runs an auto upgrade from the last version to the current release candidate, runs a multiversion upgrade from some previous major version

// Note: these upgrade tests should be deployment independent and call migrator methods from relevnat packages rather than as an invocation of migrator binary.
// Invocation of migrator binary should be done by the upgradeTests defined in the various deployment repos.

// Helper functions
//
// Setup:
//   initDBs: creates dbs at some specified version, seeds with data for the purpose of testing that OOB migrations are working correctly
//   createNetwork: creates a docker network for the test
//   base string for invocations of migrator binary

// runMigrator
//   up
//   upgrade
//   drift

// DB checks:
//   verify frontend db version
//   verify migration_logs table

// Test tests:
//   introduce a bad db migration i.e. alter metadata for a registered migration
//
