package localstore

import (
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestUserActivity_LogPageView(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	// Right here call:
	user, err := Users.Create(ctx, "auth0idtest", "test@test.com", "test-user", "test-user", "test provider", nil, "", "")
	if err != nil {
		t.Fatal(err)
	}

	err = UserActivity.LogPageView(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Now check the DB has correct results by doing a DB query:
	rows, err := globalDB.QueryContext(ctx, "SELECT user_id, page_views FROM user_activity WHERE user_id=$1", user.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		e := sourcegraph.UserActivity{}
		err := rows.Scan(&e.UserID, &e.PageViews)
		if err != nil {
			t.Fatal(err)
		}
		if e.PageViews != 1 {
			t.Errorf("expected 1 pageview, got %v", e.PageViews)
		}
	}

}

func TestUserActivity_LogSearchQuery(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	user, err := Users.Create(ctx, "auth0idtest", "test@test.com", "test-user", "test-user", "test provider", nil, "", "")
	if err != nil {
		t.Fatal(err)
	}

	err = UserActivity.LogSearchQuery(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := globalDB.QueryContext(ctx, "SELECT user_id, search_queries FROM user_activity WHERE user_id=$1", user.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		e := sourcegraph.UserActivity{}
		err := rows.Scan(&e.UserID, &e.SearchQueries)
		if err != nil {
			t.Fatal(err)
		}
		if e.SearchQueries != 1 {
			t.Errorf("expected 1 search query, got %v", e.SearchQueries)
		}
	}

}
