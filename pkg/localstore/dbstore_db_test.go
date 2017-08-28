package localstore

import (
	"context"
	"log"
	"os/exec"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/accesscontrol"
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

// testContext constructs a new context that holds a temporary test DB
// handle and other test configuration.
func testContext() context.Context {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "1", Login: "test"})
	ctx = accesscontrol.WithInsecureSkip(ctx, true)

	Mocks = MockStores{}

	if err := globalMigrate.Down(); err != nil {
		log.Fatal(err)
	}

	if err := globalMigrate.Up(); err != nil {
		log.Fatal(err)
	}

	return ctx
}
