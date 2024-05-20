package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	bgql "github.com/sourcegraph/sourcegraph/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestMarshalBatchChange(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(t))

	userID := bt.CreateTestUser(t, db, true).ID
	marshalledUserID := relay.MarshalID("User", userID)
	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, observation.TestContextTB(t), nil, clock)

	batchSpec := bt.CreateBatchSpec(t, ctx, bstore, "test", userID, 0)

	bc := bt.CreateBatchChange(t, ctx, bstore, "test", userID, batchSpec.ID)
	mbID := bgql.MarshalBatchChangeID(bc.ID)
	bcURL := "/batch-change/test/1"

	client := new(mockDoer)
	client.On("Do", mock.Anything).Return(fmt.Sprintf(
		`{"data": {"node": {"id": "%s", "name": "%s", "description": "%s", "state": "%s", "url": "%s", "createdAt": "2023-02-25T00:53:50Z", "updatedAt": "2023-02-25T00:53:50Z", "lastAppliedAt": null, "closedAt": null, "namespace": { "id": "%s" }, "creator": { "id": "%s" }, "lastApplier": null }}}`,
		mbID,
		bc.Name,
		bc.Description,
		bc.State(),
		bcURL,
		marshalledUserID,
		marshalledUserID,
	))

	response, err := marshalBatchChange(ctx, client, mbID)
	require.NoError(t, err)

	var have = &batchChange{}
	err = json.Unmarshal(response, have)
	require.NoError(t, err)

	want := &batchChange{
		ID:            mbID,
		Namespace:     marshalledUserID,
		Name:          bc.Name,
		Description:   bc.Description,
		State:         string(bc.State()),
		Creator:       marshalledUserID,
		LastApplier:   nil,
		URL:           bcURL,
		LastAppliedAt: nil,
		ClosedAt:      nil,
	}

	cmpIgnored := cmpopts.IgnoreFields(batchChange{}, "CreatedAt", "UpdatedAt")
	if diff := cmp.Diff(have, want, cmpIgnored); diff != "" {
		t.Errorf("mismatched response from batchChange marshal, got != want, diff(-got, +want):\n%s", diff)
	}

	client.AssertExpectations(t)
}
