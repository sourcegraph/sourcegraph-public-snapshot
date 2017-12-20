package localstore

import (
	"context"
	"log"
	"os/exec"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

func init() {
	dbname := "localstore-test"
	_ = exec.Command("dropdb", dbname).Run()
	if err := exec.Command("createdb", dbname).Run(); err != nil {
		log.Fatal(err)
	}
	ConnectToDB("dbname=" + dbname)
}

func TestMigrations(t *testing.T) {
	// Run all down migrations then up migrations again to ensure there are no SQL errors.
	if err := globalMigrate.Down(); err != nil {
		t.Errorf("error running down migrations: %s", err)
	}
	if err := globalMigrate.Up(); err != nil {
		t.Errorf("error running up migrations: %s", err)
	}
}

// testContext constructs a new context that holds a temporary test DB
// handle and other test configuration.
func testContext() context.Context {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "1", Login: "test"})

	Mocks = MockStores{}

	if err := globalMigrate.Down(); err != nil {
		log.Fatal(err)
	}

	if err := globalMigrate.Up(); err != nil {
		log.Fatal(err)
	}

	return ctx
}
