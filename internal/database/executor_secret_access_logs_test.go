package database

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestExecutorSecretAccessLogs_Create(t *testing.T) {
	ctx := context.Background()
	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(t))
	user, err := db.Users().Create(ctx, NewUser{Username: "johndoe"})
	if err != nil {
		t.Fatal(err)
	}
	secret := &ExecutorSecret{
		Key:       "GH_TOKEN",
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(actor.WithInternalActor(ctx), ExecutorSecretScopeBatches, secret, "sosecret"); err != nil {
		t.Fatal(err)
	}
	store := db.ExecutorSecretAccessLogs()

	t.Run("Create", func(t *testing.T) {
		log := &ExecutorSecretAccessLog{
			ExecutorSecretID: secret.ID,
			UserID:           &user.ID,
		}
		if err := store.Create(ctx, log); err != nil {
			t.Fatal(err)
		}
		if log.CreatedAt.IsZero() {
			t.Fatal("created_at time not set")
		}
	})
}

func TestExecutorSecretAccessLogs_GetListCount(t *testing.T) {
	ctx := context.Background()
	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(t))
	user, err := db.Users().Create(ctx, NewUser{Username: "johndoe"})
	if err != nil {
		t.Fatal(err)
	}
	secret1 := &ExecutorSecret{
		Key:       "GH_TOKEN",
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(actor.WithInternalActor(ctx), ExecutorSecretScopeBatches, secret1, "sosecret"); err != nil {
		t.Fatal(err)
	}
	secret2 := &ExecutorSecret{
		Key:       "NPM_TOKEN",
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(actor.WithInternalActor(ctx), ExecutorSecretScopeBatches, secret2, "sosecret"); err != nil {
		t.Fatal(err)
	}
	store := db.ExecutorSecretAccessLogs()

	// Create a bunch of logs. 2 for secret1 and 1 for secret2.
	createLog := func(secret *ExecutorSecret) *ExecutorSecretAccessLog {
		log := &ExecutorSecretAccessLog{
			ExecutorSecretID: secret.ID,
			UserID:           &user.ID,
		}
		if err := store.Create(ctx, log); err != nil {
			t.Fatal(err)
		}
		return log
	}

	secret1Log1 := createLog(secret1)
	secret1Log2 := createLog(secret1)
	secret2Log1 := createLog(secret2)

	t.Run("GetByID", func(t *testing.T) {
		log, err := store.GetByID(ctx, secret1Log1.ID)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(log, secret1Log1); diff != "" {
			t.Fatal(diff)
		}

		t.Run("not found", func(t *testing.T) {
			_, err := store.GetByID(ctx, secret2Log1.ID+1000)
			if err == nil {
				t.Fatal("unexpected nil error")
			}
			if !errors.As(err, &ExecutorSecretAccessLogNotFoundErr{}) {
				t.Fatal("wrong error returned")
			}
		})
	})

	t.Run("ListCount", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			opts := ExecutorSecretAccessLogsListOpts{}
			logs, _, err := store.List(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}
			count, err := store.Count(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}
			if have, want := count, len(logs); have != want {
				t.Fatalf("invalid count returned: %d", have)
			}
			if diff := cmp.Diff([]*ExecutorSecretAccessLog{secret2Log1, secret1Log2, secret1Log1}, logs); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("For given secret", func(t *testing.T) {
			opts := ExecutorSecretAccessLogsListOpts{ExecutorSecretID: secret1.ID}
			logs, _, err := store.List(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}
			count, err := store.Count(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}
			if have, want := count, len(logs); have != want {
				t.Fatalf("invalid count returned: %d", have)
			}
			if diff := cmp.Diff([]*ExecutorSecretAccessLog{secret1Log2, secret1Log1}, logs); diff != "" {
				t.Fatal(diff)
			}
		})
	})
}

func TestExecutorSecretAccessLogNotFoundError(t *testing.T) {
	err := ExecutorSecretAccessLogNotFoundErr{}
	if have := errcode.IsNotFound(err); !have {
		t.Error("TestExecutorSecretAccessLogNotFoundErr does not say it represents a not found error")
	}
}
