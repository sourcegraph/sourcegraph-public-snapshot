package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/stretchr/testify/require"
)

func TestExecutorSecret_Value(t *testing.T) {
	logger := logtest.NoOp(t)
	ctx := context.Background()

	sqldb := dbtest.NewDB(t)
	db := NewDB(logger, sqldb)

	u, err := db.Users().Create(ctx, NewUser{Username: "testuser"})
	require.NoError(t, err)

	ctx = actor.WithActor(ctx, actor.FromUser(u.ID))

	secret := &ExecutorSecret{Key: "testkey"}
	secretVal := "sosecret"
	err = db.ExecutorSecrets(keyring.Default().ExecutorSecretKey).Create(ctx, ExecutorSecretScopeBatches, secret, secretVal)
	require.NoError(t, err)

	esal := db.ExecutorSecretAccessLogs()
	val, err := secret.Value(ctx, esal)
	if err != nil {
		t.Fatal(err)
	}
	if val != secretVal {
		t.Fatalf("invalid secret value returned: want=%q have=%q", secretVal, val)
	}

	logList, _, err := esal.List(ctx, ExecutorSecretAccessLogsListOpts{})
	require.NoError(t, err)
	if len(logList) != 1 {
		t.Fatal("no access log entry created")
	}
}
