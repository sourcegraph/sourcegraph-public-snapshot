package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestBatchChangeResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CreateTestUser(t, db, true).ID
	orgName := "test-batch-change-resolver-org"
	orgID := bt.InsertTestOrg(t, db, orgName)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observation.TestContext, nil, clock)

	batchSpec := &btypes.BatchSpec{
		RawSpec:        bt.TestRawBatchSpec,
		UserID:         userID,
		NamespaceOrgID: orgID,
	}
	if err := bstore.CreateBatchSpec(ctx, batchSpec); err != nil {
		t.Fatal(err)
	}

	batchChange := &btypes.BatchChange{
		Name:           "my-unique-name",
		Description:    "The batch change description",
		NamespaceOrgID: orgID,
		CreatorID:      userID,
		LastApplierID:  userID,
		LastAppliedAt:  now,
		BatchSpecID:    batchSpec.ID,
	}
	if err := bstore.CreateBatchChange(ctx, batchChange); err != nil {
		t.Fatal(err)
	}

	s, err := newSchema(db, &Resolver{store: bstore})
	if err != nil {
		t.Fatal(err)
	}

	batchChangeAPIID := string(marshalBatchChangeID(batchChange.ID))
	namespaceAPIID := string(graphqlbackend.MarshalOrgID(orgID))
	apiUser := &apitest.User{DatabaseID: userID, SiteAdmin: true}
	wantBatchChange := apitest.BatchChange{
		ID:            batchChangeAPIID,
		Name:          batchChange.Name,
		Description:   batchChange.Description,
		State:         btypes.BatchChangeStateOpen,
		Namespace:     apitest.UserOrg{ID: namespaceAPIID, Name: orgName},
		Creator:       apiUser,
		LastApplier:   apiUser,
		LastAppliedAt: marshalDateTime(t, now),
		URL:           fmt.Sprintf("/organizations/%s/batch-changes/%s", orgName, batchChange.Name),
		CreatedAt:     marshalDateTime(t, now),
		UpdatedAt:     marshalDateTime(t, now),
		// Not closed.
		ClosedAt: "",
	}

	input := map[string]any{"batchChange": batchChangeAPIID}
	{
		var response struct{ Node apitest.BatchChange }
		apitest.MustExec(ctx, t, s, input, &response, queryBatchChange)

		if diff := cmp.Diff(wantBatchChange, response.Node); diff != "" {
			t.Fatalf("wrong batch change response (-want +got):\n%s", diff)
		}
	}
	// Test resolver by namespace and name
	byNameInput := map[string]any{"name": batchChange.Name, "namespace": namespaceAPIID}
	{
		var response struct{ BatchChange apitest.BatchChange }
		apitest.MustExec(ctx, t, s, byNameInput, &response, queryBatchChangeByName)

		if diff := cmp.Diff(wantBatchChange, response.BatchChange); diff != "" {
			t.Fatalf("wrong batch change response (-want +got):\n%s", diff)
		}
	}

	// Now soft-delete the user and check we can still access the batch change in the org namespace.
	err = database.UsersWith(logger, bstore).Delete(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}

	wantBatchChange.Creator = nil
	wantBatchChange.LastApplier = nil

	{
		var response struct{ Node apitest.BatchChange }
		apitest.MustExec(ctx, t, s, input, &response, queryBatchChange)

		if diff := cmp.Diff(wantBatchChange, response.Node); diff != "" {
			t.Fatalf("wrong batch change response (-want +got):\n%s", diff)
		}
	}

	// Now hard-delete the user and check we can still access the batch change in the org namespace.
	err = database.UsersWith(logger, bstore).HardDelete(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	{
		var response struct{ Node apitest.BatchChange }
		apitest.MustExec(ctx, t, s, input, &response, queryBatchChange)

		if diff := cmp.Diff(wantBatchChange, response.Node); diff != "" {
			t.Fatalf("wrong batch change response (-want +got):\n%s", diff)
		}
	}
}

func TestBatchChangeResolver_BatchSpecs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CreateTestUser(t, db, false).ID
	userCtx := actor.WithActor(ctx, actor.FromUser(userID))

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observation.TestContext, nil, clock)

	s, err := newSchema(db, &Resolver{store: bstore})
	if err != nil {
		t.Fatal(err)
	}

	// Non-created-from-raw, attached to batch change
	batchSpec1, err := btypes.NewBatchSpecFromRaw(bt.TestRawBatchSpec)
	if err != nil {
		t.Fatal(err)
	}
	batchSpec1.UserID = userID
	batchSpec1.NamespaceUserID = userID

	// Non-created-from-raw, not attached to batch change
	batchSpec2, err := btypes.NewBatchSpecFromRaw(bt.TestRawBatchSpec)
	if err != nil {
		t.Fatal(err)
	}
	batchSpec2.UserID = userID
	batchSpec2.NamespaceUserID = userID

	// created-from-raw, not attached to batch change
	batchSpec3, err := btypes.NewBatchSpecFromRaw(bt.TestRawBatchSpec)
	if err != nil {
		t.Fatal(err)
	}
	batchSpec3.UserID = userID
	batchSpec3.NamespaceUserID = userID
	batchSpec3.CreatedFromRaw = true

	for _, bs := range []*btypes.BatchSpec{batchSpec1, batchSpec2, batchSpec3} {
		if err := bstore.CreateBatchSpec(ctx, bs); err != nil {
			t.Fatal(err)
		}
	}

	batchChange := &btypes.BatchChange{
		// They all have the same name/description
		Name:        batchSpec1.Spec.Name,
		Description: batchSpec1.Spec.Description,

		NamespaceUserID: userID,
		CreatorID:       userID,
		LastApplierID:   userID,
		LastAppliedAt:   now,
		BatchSpecID:     batchSpec1.ID,
	}

	if err := bstore.CreateBatchChange(ctx, batchChange); err != nil {
		t.Fatal(err)
	}

	assertBatchSpecsInResponse(t, userCtx, s, batchChange.ID, batchSpec1, batchSpec2, batchSpec3)

	// When viewed as another user we don't want the created-from-raw batch spec to be returned
	otherUserID := bt.CreateTestUser(t, db, false).ID
	otherUserCtx := actor.WithActor(ctx, actor.FromUser(otherUserID))
	assertBatchSpecsInResponse(t, otherUserCtx, s, batchChange.ID, batchSpec1, batchSpec2)
}

func assertBatchSpecsInResponse(t *testing.T, ctx context.Context, s *graphql.Schema, batchChangeID int64, wantBatchSpecs ...*btypes.BatchSpec) {
	t.Helper()

	batchChangeAPIID := string(marshalBatchChangeID(batchChangeID))

	input := map[string]any{
		"batchChange":                 batchChangeAPIID,
		"includeLocallyExecutedSpecs": true,
	}

	var res struct{ Node apitest.BatchChange }
	apitest.MustExec(ctx, t, s, input, &res, queryBatchChangeBatchSpecs)

	expectedIDs := make(map[string]struct{}, len(wantBatchSpecs))
	for _, bs := range wantBatchSpecs {
		expectedIDs[string(marshalBatchSpecRandID(bs.RandID))] = struct{}{}
	}

	if have, want := res.Node.BatchSpecs.TotalCount, len(wantBatchSpecs); have != want {
		t.Fatalf("wrong count of batch changes returned, want=%d have=%d", want, have)
	}
	if have, want := res.Node.BatchSpecs.TotalCount, len(res.Node.BatchSpecs.Nodes); have != want {
		t.Fatalf("totalCount and nodes length don't match, want=%d have=%d", want, have)
	}
	for _, node := range res.Node.BatchSpecs.Nodes {
		if _, ok := expectedIDs[node.ID]; !ok {
			t.Fatalf("received wrong batch change with id %q", node.ID)
		}
	}
}

const queryBatchChange = `
fragment u on User { databaseID, siteAdmin }
fragment o on Org  { id, name }

query($batchChange: ID!){
  node(id: $batchChange) {
    ... on BatchChange {
      id, name, description, state
      creator { ...u }
      lastApplier    { ...u }
      lastAppliedAt
      createdAt
      updatedAt
      closedAt
      namespace {
        ... on User { ...u }
        ... on Org  { ...o }
      }
      url
    }
  }
}
`

const queryBatchChangeByName = `
fragment u on User { databaseID, siteAdmin }
fragment o on Org  { id, name }

query($namespace: ID!, $name: String!){
  batchChange(namespace: $namespace, name: $name) {
    id, name, description, state
    creator { ...u }
    lastApplier    { ...u }
    lastAppliedAt
    createdAt
    updatedAt
    closedAt
    namespace {
      ... on User { ...u }
      ... on Org  { ...o }
    }
    url
  }
}
`

const queryBatchChangeBatchSpecs = `
query($batchChange: ID!, $includeLocallyExecutedSpecs: Boolean){
  node(id: $batchChange) {
    ... on BatchChange {
      id
      batchSpecs(includeLocallyExecutedSpecs: $includeLocallyExecutedSpecs) { totalCount nodes { id } }
    }
  }
}
`
