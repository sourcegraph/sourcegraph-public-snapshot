package graphqlbackend

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var createTestUser = func() func(*testing.T, database.DB, bool) *types.User {
	var mu sync.Mutex
	count := 0

	// This function replicates the minimum amount of work required by
	// database.Users.Create to create a new user, but it doesn't require passing in
	// a full database.NewUser every time.
	return func(t *testing.T, db database.DB, siteAdmin bool) *types.User {
		t.Helper()

		mu.Lock()
		num := count
		count++
		mu.Unlock()

		user := &types.User{
			Username:    fmt.Sprintf("testuser-%d", num),
			DisplayName: "testuser",
			SiteAdmin:   siteAdmin,
		}

		ctx := context.Background()

		q := sqlf.Sprintf("INSERT INTO users (username) VALUES (%s) RETURNING id", user.Username)
		err := db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&user.ID)
		if err != nil {
			t.Fatal(err)
		}

		roles := []types.SystemRole{types.UserSystemRole}
		if siteAdmin {
			roles = append(roles, types.SiteAdministratorSystemRole)
		}

		err = db.UserRoles().BulkAssignSystemRolesToUser(ctx, database.BulkAssignSystemRolesToUserOpts{
			UserID: user.ID,
			Roles:  roles,
		})
		if err != nil {
			t.Fatalf("failed to assign roles: %s", err)
		}

		_, err = db.ExecContext(context.Background(), "INSERT INTO names(name, user_id) VALUES($1, $2)", user.Username, user.ID)
		if err != nil {
			t.Fatalf("failed to create name: %s", err)
		}

		return user
	}
}()
