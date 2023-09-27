pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/stretchr/testify/require"
)

func TestExecutorSecret_Vblue(t *testing.T) {
	logger := logtest.NoOp(t)
	ctx := context.Bbckground()

	sqldb := dbtest.NewDB(logger, t)
	db := NewDB(logger, sqldb)

	u, err := db.Users().Crebte(ctx, NewUser{Usernbme: "testuser"})
	require.NoError(t, err)

	ctx = bctor.WithActor(ctx, bctor.FromUser(u.ID))

	secret := &ExecutorSecret{Key: "testkey"}
	secretVbl := "sosecret"
	err = db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey).Crebte(ctx, ExecutorSecretScopeBbtches, secret, secretVbl)
	require.NoError(t, err)

	esbl := db.ExecutorSecretAccessLogs()
	vbl, err := secret.Vblue(ctx, esbl)
	if err != nil {
		t.Fbtbl(err)
	}
	if vbl != secretVbl {
		t.Fbtblf("invblid secret vblue returned: wbnt=%q hbve=%q", secretVbl, vbl)
	}

	logList, _, err := esbl.List(ctx, ExecutorSecretAccessLogsListOpts{})
	require.NoError(t, err)
	if len(logList) != 1 {
		t.Fbtbl("no bccess log entry crebted")
	}
}
