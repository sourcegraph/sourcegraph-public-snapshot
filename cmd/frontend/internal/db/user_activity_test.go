package db

import (
	"testing"
)

func TestUserActivity_LogPageView(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	// Right here call:
	user, err := Users.Create(ctx, NewUser{
		ExternalID:       "externalidtest",
		Email:            "test@test.com",
		Username:         "test-user",
		ExternalProvider: "test provider",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = UserActivity.LogPageView(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	a, err := UserActivity.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := int32(1); a.PageViews != want {
		t.Errorf("got %d, want %d", a.PageViews, want)
	}
}

func TestUserActivity_LogSearchQuery(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	user, err := Users.Create(ctx, NewUser{
		ExternalID:       "externalidtest",
		Email:            "test@test.com",
		Username:         "test-user",
		ExternalProvider: "test provider",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = UserActivity.LogSearchQuery(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	a, err := UserActivity.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := int32(1); a.SearchQueries != want {
		t.Errorf("got %d, want %d", a.SearchQueries, want)
	}
}
