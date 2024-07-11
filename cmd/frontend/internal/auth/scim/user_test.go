package scim

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetUniqueUsername(t *testing.T) {
	t.Parallel()

	tx := dbmocks.NewMockUserStore()
	tx.GetByUsernameFunc.SetDefaultReturn(nil, database.NewUserNotFoundError(1))
	ctx := context.Background()

	t.Run("valid username", func(t *testing.T) {
		username, err := getUniqueUsername(ctx, tx, 0, "validusername")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if username != "validusername" {
			t.Errorf("expected username 'validusername', got '%s'", username)
		}
	})

	t.Run("invalid username", func(t *testing.T) {
		username, err := getUniqueUsername(ctx, tx, 0, "invalid username")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(username) == 0 {
			t.Error("expected non-empty username")
		}
	})

	t.Run("existing username", func(t *testing.T) {
		tx := dbmocks.NewMockUserStore()
		tx.GetByUsernameFunc.SetDefaultReturn(&types.User{
			ID: 1,
		}, nil)

		username, err := getUniqueUsername(ctx, tx, 1, "existinguser")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if username != "existinguser" {
			t.Errorf("expected unaltered username, got '%s'", username)
		}
	})

	t.Run("existing username for different user", func(t *testing.T) {
		tx := dbmocks.NewMockUserStore()
		tx.GetByUsernameFunc.SetDefaultReturn(&types.User{
			ID: 1,
		}, nil)

		username, err := getUniqueUsername(ctx, tx, 2, "existinguser")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if username == "existinguser" {
			t.Errorf("expected unique username, got '%s'", username)
		}
	})
}
