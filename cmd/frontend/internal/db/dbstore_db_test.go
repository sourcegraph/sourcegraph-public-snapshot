package db

import (
	"context"
	"database/sql"
	"hash/fnv"
	"io"
	"log"
	"strconv"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

func init() {
	InitTest("db")

	// We can't care about security in tests, we care about speed.
	mockHashPassword = func(password string) (sql.NullString, error) {
		h := fnv.New64()
		io.WriteString(h, password)
		return sql.NullString{Valid: true, String: strconv.FormatUint(h.Sum64(), 16)}, nil
	}
	mockValidPassword = func(hash, password string) bool {
		h := fnv.New64()
		io.WriteString(h, password)
		return hash == strconv.FormatUint(h.Sum64(), 16)
	}
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

func TestPassword(t *testing.T) {
	// By default we use fast mocks for our password in tests. This ensures
	// our actual implementation is correct.
	oldHash := mockHashPassword
	oldValid := mockValidPassword
	mockHashPassword = nil
	mockValidPassword = nil
	defer func() {
		mockHashPassword = oldHash
		mockValidPassword = oldValid
	}()

	h, err := hashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	if !validPassword(h.String, "correct-password") {
		t.Fatal("validPassword should of returned true")
	}
	if validPassword(h.String, "wrong-password") {
		t.Fatal("validPassword should of returned false")
	}
}

// testContext constructs a new context that holds a temporary test DB
// handle and other test configuration.
func testContext() context.Context {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})

	Mocks = MockStores{}

	if err := globalMigrate.Down(); err != nil {
		log.Fatal(err)
	}

	if err := globalMigrate.Up(); err != nil {
		log.Fatal(err)
	}

	return ctx
}
