package localstore

import (
	"context"
	"log"
	"os/exec"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/accesscontrol"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
)

// testContext constructs a new context that holds a temporary test DB
// handle and other test configuration.
func testContext() (ctx context.Context) {
	ctx = context.Background()

	ctx = authpkg.WithActor(ctx, &authpkg.Actor{UID: "1", Login: "test"})
	ctx = accesscontrol.WithInsecureSkip(ctx, true)

	Mocks = MockStores{}

	dbname := "localstore-test"
	_ = exec.Command("createdb", dbname).Run()
	dbh, err := dbutil2.Open("dbname="+dbname, AppSchema)
	if err != nil {
		log.Fatal("testdb: open DB:", err)
	}
	if err := dbh.DropSchema(); err != nil {
		log.Fatal("testdb: drop schemas:", err)
	}
	if err := dbh.CreateSchema(); err != nil {
		log.Fatal("testdb: create schemas:", err)
	}
	ctx = context.WithValue(ctx, dbhKey, dbh)

	return ctx
}
