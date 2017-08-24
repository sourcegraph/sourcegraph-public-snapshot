package localstore

import (
	"context"
	"log"
	"os/exec"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
)

// testContext constructs a new context that holds a temporary test DB
// handle and other test configuration.
func testContext() (ctx context.Context) {
	ctx = context.Background()

	ctx = actor.WithActor(ctx, &actor.Actor{UID: "1", Login: "test"})
	ctx = accesscontrol.WithInsecureSkip(ctx, true)

	Mocks = MockStores{}

	dbname := "localstore-test"
	_ = exec.Command("createdb", dbname).Run()
	db, err := dbutil2.Open("dbname=" + dbname)
	if err != nil {
		log.Fatal("testdb: open DB:", err)
	}
	// TODO reset DB
	ctx = context.WithValue(ctx, dbKey, db)

	return ctx
}
