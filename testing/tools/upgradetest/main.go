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
	"fmt"
	"os"

	"github.com/sourcegraph/run"
)

func main() {

	ctx := context.Background()

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: upgrade_tests <image_tarball>")
		return
	}

	imageTarball := args[1]
	fmt.Printf("Image Tarball: %s\n", imageTarball)

	standardUpgradeTest(ctx)
	multiversionUpgradeTest(ctx)
	autoUpgradeTest(ctx)
	setupTestEnv(ctx, imageTarball)
}

func standardUpgradeTest(ctx context.Context) {
	run.Cmd(ctx, "echo", "Standard upgrade test").Run().Stream(os.Stdout)
}

func multiversionUpgradeTest(ctx context.Context) {
	run.Cmd(ctx, "echo", "Multiversion upgrade test").Run().Stream(os.Stdout)
}

func autoUpgradeTest(ctx context.Context) {
	run.Cmd(ctx, "echo", "Auto upgrade test").Run().Stream(os.Stdout)
}

func setupTestEnv(ctx context.Context, imageTarball string) {
	fmt.Println("Setting up test environment ", imageTarball)
	// run.Cmd(ctx, imageTarball).Run().Stream(os.Stdout)

	fmt.Println("TESTING")
	run.Bash(ctx, "docker", "network", "create", "wg_test").Run().Stream(os.Stdout)
	fmt.Println("TESTING")

	type dbImages struct {
		Name  string
		Image string
	}
	dbs := []dbImages{
		{"pgsql", "postgres-12-alpine"},
		{"codeintel-db", "codeintel-db"},
		{"codeinsights-db", "codeinsights-db"},
	}
	for _, db := range dbs {
		run.Cmd(ctx, "docker", "run", "--rm",
			"--detach",
			"--platform", "linux/amd64",
			"--name", fmt.Sprintf("wg_%s", db.Name),
			"--network=wg_test",
			fmt.Sprintf("sourcegraph/%s:5.2.0", db.Image),
		).Run().Stream(os.Stdout)
	}

	migratorBase := []string{"docker", "run", "--rm",
		"--platform", "linux/amd64",
		"--name", "wg_migrator",
		"-e", "PGHOST=wg_pgsql",
		"-e", "PGPORT=5432",
		"-e", "PGUSER=sg",
		"-e", "PGPASSWORD=sg",
		"-e", "PGDATABASE=sg",
		"-e", "PGSSLMODE=disable",
		"-e", "CODEINTEL_PGHOST=wg_codeintel-db",
		"-e", "CODEINTEL_PGPORT=5432",
		"-e", "CODEINTEL_PGUSER=sg",
		"-e", "CODEINTEL_PGPASSWORD=sg",
		"-e", "CODEINTEL_PGDATABASE=sg",
		"-e", "CODEINTEL_PGSSLMODE=disable",
		"-e", "CODEINSIGHTS_PGHOST=wg_codeinsights-db",
		"-e", "CODEINSIGHTS_PGPORT=5432",
		"-e", "CODEINSIGHTS_PGUSER=sg",
		"-e", "CODEINSIGHTS_PGPASSWORD=password",
		"-e", "CODEINSIGHTS_PGDATABASE=postgres",
		"-e", "CODEINSIGHTS_PGSSLMODE=disable",
		"--network=wg_test",
		"migrator:candidate",
		"up",
		// "file", "/migrator",
	}

	run.Cmd(ctx, migratorBase...).Run().Stream(os.Stdout)
}
