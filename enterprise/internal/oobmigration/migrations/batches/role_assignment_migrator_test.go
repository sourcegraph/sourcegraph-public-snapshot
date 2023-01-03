package batches

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestRoleAssignmentMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := basestore.NewWithHandle(db.Handle())

	migrator := NewRoleAssignmentMigrator(store, 5)
	progress, err := migrator.Progress(ctx, false)
	if err != nil {
		t.Fatal(err)
	}

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress with no DB entries, want=%f have=%f", want, have)
	}

	if err = store.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO users (username, display_name, created_at, site_admin)
		VALUES 
			(%s, %s, NOW(), %s),
			(%s, %s, NOW(), %s)
	`,
		"testuser-0",
		"testuser",
		true,
		"testuser-1",
		"testuser1",
		false,
	)); err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 0.0; have != want {
		t.Fatalf("got invalid progress with one unmigrated entry, want=%f have=%f", want, have)
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress after up migration, want=%f have=%f", want, have)
	}

	// Three records should be inserted into the user_roles table:
	// 1. For testuser-0 with DEFAULT role
	// 2. For testuser-0 WITH SITE_ADMINISTRATOR role
	// 3. For testuser-1 WITH DEFAULT role
	count, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf("SELECT COUNT(1) FROM user_roles")))
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}
