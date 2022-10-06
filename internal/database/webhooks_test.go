package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestWebhookCreate(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	tx, err := db.Transact(ctx)
	assert.Nil(t, err)
	defer func() { _ = tx.Done(errors.New("rollback")) }()

	store := tx.Webhooks(nil)

	hook := &types.Webhook{
		ID:           "",
		CodeHostKind: extsvc.KindGitHub,
		CodeHostURN:  "https://github.com",
		Secret:       types.NewEmptySecret(),
	}
	created, err := store.Create(ctx, hook)
	assert.Nil(t, err)

	// Check that the calculated fields were correctly calculated.
	assert.NotZero(t, created.ID)

	//// Check that the database has bare JSON versions of the request and
	//// response.
	//row := tx.QueryRowContext(ctx, "SELECT request, response FROM webhook_logs")
	//var haveReq, haveResp []byte
	//err = row.Scan(&haveReq, &haveResp)
	//assert.Nil(t, err)
	//
	//logRequest, err := log.Request.Decrypt(ctx)
	//assert.Nil(t, err)
	//logResponse, err := log.Response.Decrypt(ctx)
	//assert.Nil(t, err)
	//
	//wantReq, _ := json.Marshal(logRequest)
	//wantResp, _ := json.Marshal(logResponse)
	//
	//assert.Equal(t, string(wantReq), string(haveReq))
	//assert.Equal(t, string(wantResp), string(haveResp))
}
