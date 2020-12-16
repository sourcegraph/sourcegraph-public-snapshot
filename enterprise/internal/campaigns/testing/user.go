package testing

import (
	"fmt"
	"sync"
	"testing"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var CreateTestUser = func() func(*testing.T, bool) *types.User {
	var mu sync.Mutex
	count := 0

	// This function replicates the minium amount of work required by
	// db.Users.Create to create a new user, but it doesn't require passing in
	// a full db.NewUser every time.
	return func(t *testing.T, siteAdmin bool) *types.User {
		t.Helper()

		mu.Lock()
		num := count
		count++
		mu.Unlock()

		user := &types.User{
			Username:    fmt.Sprintf("testuser-%d", num),
			DisplayName: "testuser",
		}

		q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id, site_admin", user.Username, siteAdmin)
		err := dbconn.Global.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&user.ID, &user.SiteAdmin)
		if err != nil {
			t.Fatal(err)
		}

		if user.SiteAdmin != siteAdmin {
			t.Fatalf("user.SiteAdmin=%t, but expected is %t", user.SiteAdmin, siteAdmin)
		}

		_, err = dbconn.Global.Exec("INSERT INTO names(name, user_id) VALUES($1, $2)", user.Username, user.ID)
		if err != nil {
			t.Fatalf("failed to create name: %s", err)
		}

		return user
	}
}()
