// Run with bazel test //testing/tools/upgradetest:sh_upgradetest --config darwin-docker
// Run with bazel run //testing/tools/upgradetest:sh_upgradetest_run --config darwin-docker
// Run in CI: sg ci bazel run //testing/tools/upgradetest:sh_upgradetest
//
// TODO
// - Test things in CI
// - Turn the rule for this from a bazel run into a bazel test
// - Use urfave/cli to parse args
// - Make it so it can fail early if needed perhaps?
// - Print stuff (fail/pass/errs) as it goes through.

package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	_ "github.com/lib/pq"
	"k8s.io/utils/strings/slices"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/run"
)

// This is a CI test intended to verify that the upgrade process works as expected. For the three primary Sourcegraph databases (frontend, codeintel-db, and codeinsights-db)
// We conduct multiversion upgrades, and standard upgrades, based on their respective upgrade policies.
// - For Standard upgrades (migrator up) we test each patch version defined in the previous minor version of sourcegraph.
// - For MVU upgrades (migrator upgrade) we test all versions defined at least two minor versions prior to the latest patch release. i.e. all versions for which a standard upgrade will not work.
// - TODO: autoupgrades
// A test consists of upgradeing from an initial version to a candidate version. Defined as the latest builds of frontend, migrator, and the db schemes as found on the local branch.
// This test does not test Sourcegraph features, only the basic operations of schema upgrade paths.
// TODO: test OOB migrations by seeding data.
// TODO: definition file for tests with known bugs and tests to be run
func main() {
	ctx := context.Background()
	// check docker is running
	if err := run.Cmd(ctx, "docker", "ps").Run().Wait(); err != nil {
		fmt.Println("🚨 Error: could not connect to docker: ", err)
		os.Exit(1)
	}

	// Get init versions to use for initializing upgrade environments for tests
	latestMinorVersion, latestVersion, stdVersions, mvuVersions, err := getVersions(ctx)
	if err != nil {
		fmt.Println("🚨 Error: failed to get test version ranges: ", err)
		os.Exit(1)
	}

	fmt.Println("Latest version: ", latestVersion)
	fmt.Println("Latest minor version: ", latestMinorVersion)
	fmt.Println("Standard Versions:", stdVersions)
	fmt.Println("MVU Versions:", mvuVersions)

	// Get the release candidate image tarball
	args := os.Args
	fmt.Println(args[1])
	fmt.Println(args[2])
	fmt.Println(args[3])

	// initialize test results
	var results TestResults

	// Run Standard Upgrade Tests in goroutines. The current limit is set as 10 concurrent goroutines per test type (std, mvu, auto). This is to address
	// dynamic port allocation issues that occur in docker when creating many bridge networks, but tests begin to fail when a sufficient number of
	// goroutines are running on local machine. We may tune this in CI.
	// TODO this should likely be made an env var or something to make it easy to swamp out depending on the test box.
	stdTestPool := pool.New().WithMaxGoroutines(10).WithErrors()
	for _, version := range stdVersions {
		version := version
		if slices.Contains(knownBugVersions, version.String()) {
			continue
		}
		stdTestPool.Go(func() error {
			fmt.Println("std: ", version)
			start := time.Now()
			result := standardUpgradeTest(ctx, version, latestVersion)
			result.Runtime = time.Since(start)
			results.AddStdTest(result)
			return nil
		})
	}
	if err := stdTestPool.Wait(); err != nil {
		log.Fatal(err)
	}

	// Run MVU Upgrade Tests
	mvuTestPool := pool.New().WithMaxGoroutines(10).WithErrors()
	for _, version := range mvuVersions {
		version := version
		if slices.Contains(knownBugVersions, version.String()) {
			continue
		}
		mvuTestPool.Go(func() error {
			fmt.Println("mvu: ", version)
			start := time.Now()
			result := multiversionUpgradeTest(ctx, version, latestVersion)
			result.Runtime = time.Since(start)
			results.AddMVUTest(result)
			return nil
		})
	}
	if err := mvuTestPool.Wait(); err != nil {
		log.Fatal(err)
	}

	// TODO
	if err := autoUpgradeTest(ctx); err != nil {
		fmt.Println("--- 🚨 Auto Upgrade Test Failed: ", err)
		os.Exit(1)
	}

	// This is where we do the majority of our printing to stdout.
	results.OrderByVersion()
	results.PrintSimpleResults()
}

// Tests are the basic unit of this program and represent a version being tested. Its methods are generally used to control its logging behavior.
// Tests are further organized by TestResults, a test result aggregation.
type Test struct {
	Version  semver.Version
	Type     string
	Runtime  time.Duration
	LogLines []string
	Errors   []error
}

// Addlog registers a log entry.
func (t *Test) AddLog(log string) {
	t.LogLines = append(t.LogLines, log)
}

// AddError registers an error.
func (t *Test) AddError(err error) {
	t.LogLines = append(t.LogLines, err.Error())
	t.Errors = append(t.Errors, err)
}

// DisplayErrors prints errors to stdout
func (t *Test) DisplayErrors() {
	for _, err := range t.Errors {
		fmt.Println(err)
	}
}

// TODO godoc should always start with the function/method name.
// Print logs to stdout
func (t *Test) DisplayLog() {
	for _, log := range t.LogLines {
		fmt.Println(log)
	}
}

// TestResults is a collection of tests, organized by type. Its methods are generally used to control its logging behavior.
type TestResults struct {
	StandardUpgradeTests []Test
	MVUUpgradeTests      []Test
	AutoupgradeTests     []Test
	Mutex                sync.Mutex
}

// Add a standard test to the results
func (r *TestResults) AddStdTest(test Test) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.StandardUpgradeTests = append(r.StandardUpgradeTests, test)
}

// Add a multiversion upgrade test to the results
func (r *TestResults) AddMVUTest(test Test) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.MVUUpgradeTests = append(r.MVUUpgradeTests, test)
}

// Add an autoupgrade test to the results
func (r *TestResults) AddAutoTest(test Test) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.AutoupgradeTests = append(r.AutoupgradeTests, test)
}

// Known bug versions
// versions 4.1.0 to v4.4.2 are affected by a known bug in MVU if initialized in these versions: https://github.com/sourcegraph/sourcegraph/pull/46969
var knownBugVersions = []string{
	"4.1.0",
	"4.1.1",
	"4.1.2",
	"4.1.3",
	"4.2.0",
	"4.2.1",
	"4.3.0",
	"4.3.1",
	"4.4.0",
	"4.4.1",
	"4.4.2",
}

// Display quick view of test resutls, errors will log the first error line only.
func (r *TestResults) PrintSimpleResults() {
	stdRes := []string{}
	for _, test := range r.StandardUpgradeTests {
		if 0 < len(test.Errors) {
			stdRes = append(stdRes, fmt.Sprintf("🚨 %s Failed -- %s\n%s", test.Version.String(), test.Runtime, test.Errors[0]))
		} else {
			stdRes = append(stdRes, fmt.Sprintf("✅ %s Passed -- %s ", test.Version.String(), test.Runtime))
		}
	}
	mvuRes := []string{}
	for _, test := range r.MVUUpgradeTests {
		if 0 < len(test.Errors) {
			mvuRes = append(mvuRes, fmt.Sprintf("🚨 %s Failed -- %s\n%s", test.Version.String(), test.Runtime, test.Errors[0]))
		} else {
			mvuRes = append(mvuRes, fmt.Sprintf("✅ %s Passed -- %s", test.Version.String(), test.Runtime))
		}
	}
	autoRes := []string{}
	for _, test := range r.AutoupgradeTests {
		if 0 < len(test.Errors) {
			autoRes = append(autoRes, fmt.Sprintf("🚨 %s Failed -- %s\n%s", test.Version.String(), test.Runtime, test.Errors[0]))
		} else {
			autoRes = append(autoRes, fmt.Sprintf("✅ %s Passed -- %s", test.Version.String(), test.Runtime))
		}
	}
	fmt.Println("--- Standard Upgrade Tests:")
	fmt.Println(strings.Join(stdRes, "\n"))
	fmt.Println("--- Multiversion Upgrade Tests:")
	fmt.Println(strings.Join(mvuRes, "\n"))
	fmt.Println("--- Auto Upgrade Tests:")
	fmt.Println(strings.Join(autoRes, "\n"))
}

// Display test Errors, prints errors for all tests that errored.
func (r *TestResults) DisplayErrors() {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	for _, test := range r.StandardUpgradeTests {
		if 0 < len(test.Errors) {
			fmt.Printf("--- 🚨 Standard Upgrade Test %s Failed:\n", test.Version.String())
			test.DisplayErrors()
		}
	}
	for _, test := range r.MVUUpgradeTests {
		if 0 < len(test.Errors) {
			fmt.Printf("--- 🚨 Multiversion Upgrade Test %s Failed:\n", test.Version.String())
			test.DisplayErrors()
		}
	}
	for _, test := range r.AutoupgradeTests {
		if 0 < len(test.Errors) {
			fmt.Printf("--- 🚨 Auto Upgrade Test %s Failed:\n", test.Version.String())
			test.DisplayErrors()
		}
	}
}

// Sort tests TestResults by their test.Version value
func (r *TestResults) OrderByVersion() {
	sort.Slice(r.StandardUpgradeTests, func(i, j int) bool {
		return r.StandardUpgradeTests[i].Version.LessThan(&r.StandardUpgradeTests[j].Version)
	})
	sort.Slice(r.MVUUpgradeTests, func(i, j int) bool {
		return r.MVUUpgradeTests[i].Version.LessThan(&r.MVUUpgradeTests[j].Version)
	})
	sort.Slice(r.AutoupgradeTests, func(i, j int) bool {
		return r.AutoupgradeTests[i].Version.LessThan(&r.AutoupgradeTests[j].Version)
	})
}

// standardUpgradeTest initializes Sourcegraph's dbs and runs a standard upgrade
// i.e. an upgrade test between some last minor version and the current release candidate
func standardUpgradeTest(ctx context.Context, initVersion, migratorVersion *semver.Version) Test {
	//start test env
	test, networkName, dbs, cleanup, err := setupTestEnv(ctx, "standard", initVersion)
	if err != nil {
		test.AddError(fmt.Errorf("failed to setup env: %w", err))
		cleanup()
		return test
	}
	defer cleanup()

	// ensure env correctly initialized
	if err := validateDBs(ctx, &test, initVersion.String(), fmt.Sprintf("sourcegraph/migrator:%s", migratorVersion.String()), networkName, dbs, false); err != nil {
		test.AddError(fmt.Errorf("🚨 Upgrade failed: %w", err))
		return test
	}

	test.AddLog("-- ⚙️  performing standard upgrade")

	// Run standard upgrade via migrators "up" command
	out, err := run.Cmd(ctx, dockerMigratorBaseString(test, "up", "migrator:candidate", networkName, dbs)...).Run().String()
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to upgrade: %w", err))
		cleanup()
		return test
	}
	test.AddLog(out)

	// Start frontend with candidate
	var cleanFrontend func()
	cleanFrontend, err = startFrontend(ctx, test, "frontend", "candidate", networkName, dbs)
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to start candidate frontend: %w", err))
		cleanFrontend()
		return test
	}
	defer cleanFrontend()

	test.AddLog("-- ⚙️  post upgrade validation")
	// Validate the upgrade
	if err := validateDBs(ctx, &test, "0.0.0+dev", "migrator:candidate", networkName, dbs, true); err != nil {
		test.AddError(fmt.Errorf("🚨 Upgrade failed: %w", err))
		return test
	}

	// test.DisplayLog()
	return test
}

// multiversionUpgradeTest tests the migrator upgrade command,
// initializing the three main dbs and conducting an upgrade to the release candidate version
func multiversionUpgradeTest(ctx context.Context, initVersion, latestVersion *semver.Version) Test {

	test, networkName, dbs, cleanup, err := setupTestEnv(ctx, "multiversion", initVersion)
	if err != nil {
		fmt.Println("failed to setup env: ", err)
		cleanup()
		return test
	}
	defer cleanup()

	// ensure env correctly initialized, always use latest migrator for drift check,
	// this allows us to avoid issues from changes in migrators invocation
	if err := validateDBs(ctx, &test, initVersion.String(), fmt.Sprintf("sourcegraph/migrator:%s", latestVersion.String()), networkName, dbs, false); err != nil {
		test.AddError(fmt.Errorf("🚨 Initializing env in multiversion test failed: %w", err))
		return test
	}

	// Run multiversion upgrade using candidate image
	// TODO: target the schema of the candidate version rather than latest released tag on branch, this has less effect than you might expect.
	// migrator upgrade only applies migrations defined between first minor versions i.e. 5.1 -> 5.2, migrator `up` incorperates migrations defined in a patch version and is run later in the test.
	// We may be able to work around this furhter in CI
	test.AddLog(fmt.Sprintf("-- ⚙️  performing multiversion upgrade (--from %s --to %s)", initVersion.String(), latestVersion.String()))
	out, err := run.Cmd(ctx,
		dockerMigratorBaseString(test, fmt.Sprintf("upgrade --from %s --to %s", initVersion.String(), latestVersion.String()), "migrator:candidate", networkName, dbs)...).
		Run().String()
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to upgrade: %w", err))
		cleanup()
		return test
	}
	test.AddLog(out)

	// Run migrator up with migrator candidate to apply any patch migrations defined on the candidate version
	out, err = run.Cmd(ctx,
		dockerMigratorBaseString(test, "up", "migrator:candidate", networkName, dbs)...).
		Run().String()
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to upgrade: %w", err))
		cleanup()
		return test
	}
	test.AddLog(out)

	// Start frontend with candidate
	var cleanFrontend func()
	cleanFrontend, err = startFrontend(ctx, test, "frontend", "candidate", networkName, dbs)
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to start candidate frontend: %w", err))
		cleanFrontend()
		return test
	}
	defer cleanFrontend()

	test.AddLog("-- ⚙️  post upgrade validation")
	// Validate the upgrade
	if err := validateDBs(ctx, &test, "0.0.0+dev", "migrator:candidate", networkName, dbs, true); err != nil {
		test.AddError(fmt.Errorf("🚨 Upgrade failed: %w", err))
		return test
	}

	// test.DisplayLog()
	return test
}

// TODO
func autoUpgradeTest(ctx context.Context) error {
	fmt.Println("--- 🕵️  auto upgrade test")
	return nil
}

// testDB is an organizational type to make orchestrating the three dbs easier, and also to store a dynamically allocated port for postgres
type testDB struct {
	DbName            string
	ContainerName     string
	Image             string
	ContainerHostPort string
}

// Initialize a test environment ad object. Create a docker network for testing as well as instances of our three databases. Returning a cleanup function.
// An instance of Sourcegraph-Frontend is also started to initialize the versions table of the database.
// TODO: setupTestEnv should seed some initial data at the target initVersion. This will be usefull for testing OOB migrations
func setupTestEnv(ctx context.Context, testType string, initVersion *semver.Version) (test Test, networkName string, dbs []*testDB, cleanup func(), err error) {
	test = Test{
		Version:  *initVersion,
		Type:     testType,
		LogLines: []string{},
		Errors:   []error{},
	}

	if testType == "standard" {
		test.AddLog("--- 🕵️  standard upgrade test")
	}
	if testType == "multiversion" {
		test.AddLog("--- 🕵️  multiversion upgrade test")
	}
	test.AddLog(fmt.Sprintf("Upgrading from version (%s) to release candidate.", initVersion))
	test.AddLog("-- 🏗️  setting up test environment")

	// Create a docker network for testing
	//
	// Docker bridge networks take up a lot of the docker daemons available port allocation. We run only a limited amount of test parallelization to get around this.
	// see https://straz.to/2021-09-08-docker-address-pools/
	networkName = fmt.Sprintf("%s_test_%s", testType, initVersion)
	test.AddLog(fmt.Sprintf("🐋 creating network %s", networkName))

	out, err := run.Cmd(ctx, "docker", "network", "create", networkName).Run().String()
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to create test network: %s", err))
	}
	test.AddLog(out)

	// Note that we changed postgres versions in very early versions of Sourcegraph,
	// In v3.38+ we use image postgres-12-alpine,
	// in v3.37-v3.30 we use postgres-12.6-alpine,
	// in v3.29-v3.27 we use postgres-12.6
	// in v3.26 and earliar we use postgres:11.4
	//
	// This isn't relevant since this test will only ever initialize instances v3.38+
	// worth noting in case this changes in the future.
	dbs = []*testDB{
		{"pgsql", fmt.Sprintf("wg_pgsql_%s", initVersion), "postgres-12-alpine", ""},
		{"codeintel-db", fmt.Sprintf("wg_codeintel-db_%s", initVersion), "codeintel-db", ""},
		{"codeinsights-db", fmt.Sprintf("wg_codeinsights-db_%s", initVersion), "codeinsights-db", ""},
	}

	// Here we create the three databases using docker run.
	for _, db := range dbs {
		test.AddLog(fmt.Sprintf("🐋 creating %s, with db image %s:%s", db.ContainerName, db.Image, initVersion))
		err := run.Cmd(ctx, "docker", "run", "--rm",
			"--detach",
			"--platform", "linux/amd64",
			"--name", db.ContainerName,
			"--network", networkName,
			"-p", "5432",
			fmt.Sprintf("sourcegraph/%s:%s", db.Image, initVersion),
		).Run().Wait()
		if err != nil {
			test.AddError(fmt.Errorf("🚨 failed to create test databases: %s", err))
		}
		// get the dynamically allocated port and register it to the test
		port, err := run.Cmd(ctx, "docker", "port", db.ContainerName, "5432").Run().String()
		if err != nil {
			test.AddError(fmt.Errorf("🚨 failed to get port for %s: %s", db.ContainerName, err))
		}
		db.ContainerHostPort = port
	}

	// Create a 10 minute timeout to validate the databases have initialized, this is to prevent a hung test
	// When many goroutines are running this test this is a point of failure.
	dbPingTimeout, cancel := context.WithTimeout(ctx, time.Minute*10)
	wgDbPing := pool.New().WithErrors().WithContext(dbPingTimeout)
	defer cancel()

	// Here we poll/ping the dbs to ensure postgres has initialized before we make calls to the databases.
	for _, db := range dbs {
		db := db // this closure locks the index for the inner for loop
		wgDbPing.Go(func(ctx context.Context) error {
			// TODO only ping codeinsights after timescaleDB is deprecated in v3.39
			dbClient, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@%s/sg?sslmode=disable", db.ContainerHostPort))
			if err != nil {
				test.AddError(fmt.Errorf("🚨 failed to connect to %s: %s", db.DbName, err))
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
					test.AddLog(fmt.Sprintf(" ... pinging %s", db.DbName))
					if err == sql.ErrConnDone || strings.Contains(err.Error(), "connection refused") {
						test.AddError(fmt.Errorf("🚨 unrecoverable error pinging %s: %w", db.DbName, err))
						return err
					}
					time.Sleep(1 * time.Second)
					continue
				} else {
					test.AddLog(fmt.Sprintf("✅ %s is up", db.DbName))
					return nil
				}
			}
		})
	}
	if err := wgDbPing.Wait(); err != nil {
		test.AddError(fmt.Errorf("🚨 containerized database startup error: %w", err))
	}

	// Initialize the databases by running migrator with the `up` command.
	test.LogLines = append(test.LogLines, "-- 🏗️  initializing database schemas with migrator")
	out, err = run.Cmd(ctx, dockerMigratorBaseString(test, "up", fmt.Sprintf("sourcegraph/migrator:%s", initVersion), networkName, dbs)...).Run().String()
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to initialize database: %w", err))
	}
	test.AddLog(out)

	// Verify that the databases are initialized.
	test.AddLog("🔎 checking db schemas initialized")
	for _, db := range dbs {
		dbClient, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@%s/sg?sslmode=disable", db.ContainerHostPort))
		if err != nil {
			test.AddError(fmt.Errorf("🚨 failed to connect to %s: %s", db.DbName, err))
			continue
		}
		defer dbClient.Close()

		// check if tables have been created
		rows, err := dbClient.Query(`SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public';`)
		if err != nil {
			test.AddError(fmt.Errorf("🚨 failed to check %s for init: %s", db.DbName, err))
			continue
		}
		defer rows.Close()
		if err := rows.Err(); err != nil {
			test.AddError(fmt.Errorf("🚨 failed to check %s for init: %s", db.DbName, err))
			continue
		} else {
			test.AddLog(fmt.Sprintf("✅ %s initialized", db.DbName))
		}
	}

	//start frontend and poll db until initial version is set by frontend
	var cleanFrontend func()
	cleanFrontend, err = startFrontend(ctx, test, "sourcegraph/frontend", initVersion.String(), networkName, dbs)
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to start frontend: %w", err))
	}
	defer cleanFrontend()

	// Return a cleanup function that will remove the containers and network.
	cleanup = func() {
		test.LogLines = append(test.LogLines, "🧹 removing database containers")
		out, err := run.Cmd(ctx, "docker", "container", "stop",
			dbs[0].ContainerName,
			dbs[1].ContainerName,
			dbs[2].ContainerName).
			Run().String()
		if err != nil {
			test.AddError(fmt.Errorf("🚨 failed to stop database containers after testing: %w", err))
		}
		test.AddLog(out)
		out, err = run.Cmd(ctx, "docker", "container", "rm",
			dbs[0].ContainerName,
			dbs[1].ContainerName,
			dbs[2].ContainerName).
			Run().String()
		if err != nil {
			test.AddError(fmt.Errorf("🚨 failed to remove database containers after testing: %w", err))
		}
		test.AddLog(out)
		test.AddLog("🧹 removing testing network")
		out, err = run.Cmd(ctx, "docker", "network", "rm", networkName).Run().String()
		if err != nil {
			test.AddError(fmt.Errorf("🚨 failed to remove test network after testing: %w", err))
		}
		test.AddLog(out)
	}

	test.AddLog("-- 🏗️  setup complete")

	return test, networkName, dbs, cleanup, err
}

// validateDBs runs a few tests to assess the readiness of the database and wether or not drift exists on the schema.
// It is used in initializing a new db as well as "validating" the db after an version change. This behavior is controlled by the upgrade parameter.
func validateDBs(ctx context.Context, test *Test, version, migratorImage, networkName string, dbs []*testDB, upgrade bool) error {
	test.AddLog("-- ⚙️  validating dbs")

	// Get DB clients
	clients := make(map[string]*sql.DB)
	for _, db := range dbs {
		client, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@%s/sg?sslmode=disable", db.ContainerHostPort))
		if err != nil {
			test.AddError(fmt.Errorf("🚨 failed to connect to %s: %w", db.DbName, err))
			return err
		}
		defer client.Close()

		clients[db.DbName] = client
	}

	// Verify the versions.version value in the frontend db
	test.AddLog("🔎 checking pgsql versions.version set")
	var testVersion string
	row := clients["pgsql"].QueryRowContext(ctx, `SELECT version FROM versions;`)
	if err := row.Scan(&testVersion); err != nil {
		test.AddError(fmt.Errorf("🚨 failed to get version from pgsql db: %w", err))
	}
	if version != testVersion {
		test.AddError(fmt.Errorf("🚨 versions.version not set: %s!= %s", version, testVersion))
	}

	test.AddLog(fmt.Sprintf("✅ versions.version set: %s", testVersion))

	// Check for any failed migrations in the migration logs tables
	// migration_logs table is introduced in v3.36.0
	for _, db := range dbs {
		test.AddLog(fmt.Sprintf("🔎 checking %s migration_logs", db.ContainerName))
		var numFailedMigrations int
		row = clients[db.DbName].QueryRowContext(ctx, `SELECT count(*) FROM migration_logs WHERE success=false;`)
		if err := row.Scan(&numFailedMigrations); err != nil {
			test.AddError(fmt.Errorf("🚨 failed to get failed migrations from %s db: %w", db.ContainerName, err))
		}
		if numFailedMigrations > 0 {
			test.AddError(fmt.Errorf("🚨 failed migrations found: %d", numFailedMigrations))
		}

		test.AddLog("✅ no failed migrations found")
	}

	// Check DBs for drift
	test.AddLog("🔎 Checking DBs for drift")
	if upgrade {
		// Get the last commit in the release branch, if validating an upgrade the upgrade boolean is true,
		// in this case the drift target is the latest commit on the release candidate branch.
		// If working on this, the drift check will fail if you have local commits not yet pushed to remote.
		// example schema check target: https://raw.githubusercontent.com/sourcegraph/sourcegraph/7648573357fef049e1a0bf11f400068ef83f2596/internal/database/schema.json
		var candidateGitHead bytes.Buffer
		if err := run.Cmd(ctx, "git", "rev-parse", "HEAD").Run().Stream(&candidateGitHead); err != nil {
			test.AddError(fmt.Errorf("🚨 failed to get latest commit on candidate branch: %w", err))
		}
		test.AddLog(fmt.Sprintf("Latest commit on candidate branch: %s", candidateGitHead.String()))
		for _, db := range dbs {
			out, err := run.Cmd(ctx, dockerMigratorBaseString(*test, fmt.Sprintf("drift --db %s --version %s --ignore-migrator-update --skip-version-check", db.DbName, candidateGitHead.String()),
				migratorImage, networkName, dbs)...).Run().String()
			if err != nil {
				test.AddError(fmt.Errorf("🚨 failed to check drift on %s: %s", db.DbName, err))
			}
			test.AddLog(out)
		}
	} else {
		for _, db := range dbs {
			out, err := run.Cmd(ctx, dockerMigratorBaseString(*test, fmt.Sprintf("drift --db %s --version v%s --ignore-migrator-update", db.DbName, version),
				migratorImage, networkName, dbs)...).Run().String()
			if err != nil {
				test.AddError(fmt.Errorf("🚨 failed to check drift on %s: %w", db.DbName, err))
			}
			test.AddLog(out)
		}
	}

	return nil
}

// startFrontend starts the frontend container in the CI test env.
func startFrontend(ctx context.Context, test Test, image, version, networkName string, dbs []*testDB) (cleanup func(), err error) {
	hash, err := newContainerHash()
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to get container hash: %w", err))
		return nil, err
	}
	test.AddLog(fmt.Sprintf("🐋 creating wg_frontend_%x", hash))
	cleanup = func() {
		test.AddLog("🧹 removing frontend container")
		out, err := run.Cmd(ctx, "docker", "container", "stop",
			fmt.Sprintf("wg_frontend_%x", hash),
		).Run().String()
		if err != nil {
			fmt.Println("🚨 failed to stop frontend after testing: ", err)
		}
		test.AddLog(out)
		out, err = run.Cmd(ctx, "docker", "container", "rm",
			fmt.Sprintf("wg_frontend_%x", hash),
		).Run().String()
		if err != nil {
			fmt.Println("🚨 failed to remove frontend after testing: ", err)
		}
		test.AddLog(out)
	}

	// Start the frontend container
	err = run.Cmd(ctx, "docker", "run",
		"--detach",
		"--platform", "linux/amd64",
		"--name", fmt.Sprintf("wg_frontend_%x", hash),
		"-e", "DEPLOY_TYPE=docker-compose",
		"-e", fmt.Sprintf("PGHOST=%s", dbs[0].ContainerName),
		"-e", fmt.Sprintf("CODEINTEL_PGHOST=%s", dbs[1].ContainerName),
		"-e", fmt.Sprintf("CODEINSIGHTS_PGDATASOURCE=postgres://sg@%s:5432/sg?sslmode=disable", dbs[2].ContainerName),
		"--network", networkName,
		fmt.Sprintf("%s:%s", image, version),
	).Run().Wait()
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to start frontend: %w", err))
		return cleanup, err
	}

	// poll db until initial versions.version is set
	setInitTimeout, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	test.AddLog("🔎 checking db initialization complete")

	dbClient, err := sql.Open("postgres", fmt.Sprintf("postgres://sg@%s/sg?sslmode=disable", dbs[0].ContainerHostPort))
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to connect to %s: %w", dbs[0].DbName, err))
	}
	defer dbClient.Close()

	// Poll till versions.version is set
	for {
		select {
		case <-setInitTimeout.Done():
			return cleanup, setInitTimeout.Err()
		default:
		}
		// check version string set
		var dbVersion string
		row := dbClient.QueryRowContext(setInitTimeout, versionQuery)
		err = row.Scan(&dbVersion)
		if err != nil {
			test.AddLog(fmt.Sprintf("... querying versions.version: %s", err))
			time.Sleep(1 * time.Second)
			continue
		}
		// wait for the frontend to set the database versions.version value before considering the frontend startup complete.
		// "candidate" resolves to "0.0.0+dev" and should always be valid
		if dbVersion == version || dbVersion == "0.0.0+dev" {
			test.AddLog(fmt.Sprintf("✅ versions.version is set: %s", dbVersion))
			break
		}
		if version != dbVersion {
			time.Sleep(1 * time.Second)
			test.AddLog(fmt.Sprintf(" ... waiting for versions.version to be set: %s", version))
			continue
		}
	}

	// poll db until site-config is initialized, migrator will sometimes fail if this initialization of the frontend db hasnt finished
	// returning an error like: "instance is new"
	for {
		select {
		case <-setInitTimeout.Done():
			return cleanup, setInitTimeout.Err()
		default:
		}
		// check version string set
		var siteConfig string
		row := dbClient.QueryRowContext(setInitTimeout, siteConfigQuery)
		err = row.Scan(&siteConfig)
		if err != nil {
			test.AddLog(fmt.Sprintf("... checking site-config initialized: %s", err))
			time.Sleep(1 * time.Second)
			continue
		}
		if siteConfig == "" {
			test.AddLog("... waiting for site-config to be initialized")
			time.Sleep(1 * time.Second)
			continue
		} else {
			test.AddLog("✅ site-config is initialized")
			break
		}
	}

	return cleanup, nil
}

const versionQuery = `SELECT version FROM versions;`

const siteConfigQuery = `
SELECT c.contents
FROM critical_and_site_config c
WHERE c.type = 'site'
ORDER BY c.id DESC
LIMIT 1
`

// dockerMigratorBaseString a slice of strings constituting the necessary arguments to run the migrator via docker container the CI test env.
func dockerMigratorBaseString(test Test, cmd, migratorImage, networkName string, dbs []*testDB) []string {
	hash, err := newContainerHash()
	if err != nil {
		test.AddError(fmt.Errorf("🚨 failed to get container hash: %w", err))
		return nil
	}
	return []string{"docker", "run", "--rm",
		"--platform", "linux/amd64",
		"--name", fmt.Sprintf("wg_migrator_%x", hash),
		// "-e", "SRC_LOG_LEVEL=debug",
		"-e", fmt.Sprintf("PGHOST=%s", dbs[0].ContainerName),
		"-e", "PGPORT=5432",
		"-e", "PGUSER=sg",
		"-e", "PGPASSWORD=sg",
		"-e", "PGDATABASE=sg",
		"-e", "PGSSLMODE=disable",
		"-e", fmt.Sprintf("CODEINTEL_PGHOST=%s", dbs[1].ContainerName),
		"-e", "CODEINTEL_PGPORT=5432",
		"-e", "CODEINTEL_PGUSER=sg",
		"-e", "CODEINTEL_PGPASSWORD=sg",
		"-e", "CODEINTEL_PGDATABASE=sg",
		"-e", "CODEINTEL_PGSSLMODE=disable",
		"-e", fmt.Sprintf("CODEINSIGHTS_PGHOST=%s", dbs[2].ContainerName),
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

// Generate random hash for naming containers in test, used for frontend and migrator
func newContainerHash() ([]byte, error) {
	hash := make([]byte, 4)
	_, err := rand.Read(hash)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// getLatestVersions returns the latest minor semver version, as well as the latest full semver version.
//
// Technically MVU is supported v3.20 and forward, but in older versions codeinsights-db didnt exist and postgres was using version 11.4
// so we reduced the scope of the test.
func getVersions(ctx context.Context) (latestMinor, latestFull *semver.Version, stdVersions, mvuVersions []*semver.Version, err error) {
	tags, err := run.Cmd(ctx, "git",
		"for-each-ref",
		"--format", "'%(refname:short)'",
		"refs/tags").Run().Lines()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var validTags []*semver.Version
	var latestMinorVer *semver.Version
	var latestFullVer *semver.Version

	// Get valid tags
	for _, tag := range tags {
		v, err := semver.NewVersion(tag)
		if err != nil {
			continue // skip non-matching tags
		}
		if v.Prerelease() != "" {
			continue // skip prereleases
		}
		// To simplify this testing range we'll only select a version tags from versions greater than v3.38
		// In v3.39 many things were normalized in the dbs:
		// - codeinsights-db moved from timescaleDB to posgres-12
		// - our image for codeintel-db and pgsql was normalized to postgres-12-alpine
		// - the migration_logs table exists, this was renamed from schema_migrations in v3.36.0
		// - migrator is introduced in v3.38.0
		if v.LessThan(semver.MustParse("v3.39.0")) {
			continue
		}
		validTags = append(validTags, v)
	}

	// Get latest Version and latestMinorVersion
	for _, tag := range validTags {
		// Track latest full version
		if latestFullVer == nil || tag.GreaterThan(latestFullVer) {
			latestFullVer = tag
		}
		latestMinorVer, err = semver.NewVersion(fmt.Sprintf("%d.%d.0", latestFullVer.Major(), latestFullVer.Minor()))
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	// Get range for standardUpgrade test pool and MVU test pool
	for _, tag := range validTags {
		std, err := semver.NewConstraint(fmt.Sprintf(">= %d.%d.x", latestMinorVer.Major(), latestMinorVer.Minor()-1))
		if err != nil {
			fmt.Println("🚨 failed to collect versions for standard upgrade test: ", err)
		}
		// Sort versions into those for the standard test (one minor version behind the latest release candidate), and those for multiversion testing.
		if std.Check(tag) {
			stdVersions = append(stdVersions, tag)
		} else {
			mvuVersions = append(mvuVersions, tag)
		}
	}

	if latestMinorVer == nil {
		return nil, nil, nil, nil, errors.New("No valid minor semver tags found")
	}
	if latestFullVer == nil {
		return nil, nil, nil, nil, errors.New("No valid full semver tags found")
	}

	return latestMinorVer, latestFullVer, stdVersions, mvuVersions, nil

}
