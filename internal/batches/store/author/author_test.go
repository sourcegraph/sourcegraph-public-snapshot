package author

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGetChangesetAuthorForUser(t *testing.T) {

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))

	userStore := db.Users()

	t.Run("User ID doesnt exist", func(t *testing.T) {
		author, err := GetChangesetAuthorForUser(ctx, userStore, 0)
		if err != nil {
			t.Fatal(err)
		}
		if author != nil {
			t.Fatalf("got non-nil author email when author doesnt exist: %v", author)
		}
	})

	t.Run("User exists but doesn't have an email", func(t *testing.T) {

		user, err := userStore.Create(ctx, database.NewUser{
			Username: "mary",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		author, err := GetChangesetAuthorForUser(ctx, userStore, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if author != nil {
			t.Fatalf("got non-nil author email when user doesnt have an email: %v", author)
		}
	})

	t.Run("User exists and has an e-mail but doesn't have a display name", func(t *testing.T) {

		user, err := userStore.Create(ctx, database.NewUser{
			Username:        "jane",
			Email:           "jane1@doe.com",
			EmailIsVerified: true,
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
		author, err := GetChangesetAuthorForUser(ctx, userStore, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, author.Name, user.Username)
	})

	t.Run("User exists", func(t *testing.T) {

		user, err := userStore.Create(ctx, database.NewUser{
			Username:        "johnny",
			Email:           "john@test.com",
			EmailIsVerified: true,
			DisplayName:     "John Tester",
		})

		userEmail := "john@test.com"

		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
		author, err := GetChangesetAuthorForUser(ctx, userStore, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if author.Email != userEmail {
			t.Fatalf("found incorrect email: %v", author)
		}

		if author.Name != user.DisplayName {
			t.Fatalf("found incorrect name: %v", author)
		}
	})
}
