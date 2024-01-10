package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	bgql "github.com/sourcegraph/sourcegraph/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestMarshalChangeset(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(t))

	userID := bt.CreateTestUser(t, db, true).ID
	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observation.TestContext, nil, clock)

	batchSpec := bt.CreateBatchSpec(t, ctx, bstore, "test", userID, 0)

	batchChange := bt.CreateBatchChange(t, ctx, bstore, "test", userID, batchSpec.ID)
	mbID := bgql.MarshalBatchChangeID(batchChange.ID)

	repos, _ := bt.CreateTestRepos(t, ctx, db, 3)

	repoOne := repos[0]
	repoOneID := gql.MarshalRepositoryID(repoOne.ID)

	repoTwo := repos[1]
	repoTwoID := gql.MarshalRepositoryID(repoTwo.ID)

	uc := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:               repoOne.ID,
		BatchChange:        batchChange.ID,
		PublicationState:   btypes.ChangesetPublicationStatePublished,
		OwnedByBatchChange: batchChange.ID,
	})
	// associate changeset with batch change
	addChangeset(t, ctx, bstore, uc, batchChange.ID)
	mucID := bgql.MarshalChangesetID(uc.ID)
	ucTitle, err := uc.Title()
	require.NoError(t, err)
	ucBody, err := uc.Body()
	require.NoError(t, err)
	ucExternalURL := "https://github.com/test/test/pull/62"
	ucReviewState := string(btypes.ChangesetReviewStateApproved)

	ic := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:             repos[1].ID,
		BatchChange:      batchChange.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	// associate changeset with batch change
	addChangeset(t, ctx, bstore, uc, batchChange.ID)
	micID := bgql.MarshalChangesetID(ic.ID)
	icTitle, err := ic.Title()
	require.NoError(t, err)
	icBody, err := ic.Body()
	require.NoError(t, err)
	icExternalURL := "https://github.com/test/test-2/pull/62"
	icReviewState := string(btypes.ChangesetReviewStateChangesRequested)

	authorName := "TestUser"
	authorEmail := "test@sourcegraph.com"

	testcases := []struct {
		changeset    *btypes.Changeset
		name         string
		httpResponse string
		want         *changeset
	}{
		{
			changeset: uc,
			name:      "unimported changeset",
			httpResponse: fmt.Sprintf(
				`{"data": {"node": {"id": "%s","externalID": "%s","batchChanges": {"nodes": [{"id": "%s"}]},"repository": {"id": "%s","name": "github.com/test/test"},"createdAt": "2023-02-25T00:53:50Z","updatedAt": "2023-02-25T00:53:50Z","title": "%s","body": "%s","author": {"name": "%s", "email": "%s"},"state": "%s","labels": [],"externalURL": {"url": "%s"},"forkNamespace": null,"reviewState": "%s","checkState": null,"error": null,"syncerError": null,"forkName": null,"ownedByBatchChange": "%s"}}}`,
				mucID,
				uc.ExternalID,
				mbID,
				repoOneID,
				ucTitle,
				ucBody,
				authorName,
				authorEmail,
				uc.State,
				ucExternalURL,
				ucReviewState,
				mbID,
			),
			want: &changeset{
				ID:                 mucID,
				ExternalID:         uc.ExternalID,
				RepositoryID:       gql.MarshalRepositoryID(uc.RepoID),
				CreatedAt:          now,
				UpdatedAt:          now,
				BatchChangeIDs:     []graphql.ID{mbID},
				State:              string(uc.State),
				OwnedByBatchChange: &mbID,
				Title:              &ucTitle,
				Body:               &ucBody,
				AuthorName:         &authorName,
				AuthorEmail:        &authorEmail,
				ExternalURL:        &ucExternalURL,
				ReviewState:        &ucReviewState,
			},
		},
		{
			changeset: ic,
			name:      "imported changeset",
			httpResponse: fmt.Sprintf(
				`{"data": {"node": {"id": "%s","externalID": "%s","batchChanges": {"nodes": [{"id": "%s"}]},"repository": {"id": "%s","name": "github.com/test/test"},"createdAt": "2023-02-25T00:53:50Z","updatedAt": "2023-02-25T00:53:50Z","title": "%s","body": "%s","author": {"name": "%s", "email": "%s"},"state": "%s","labels": [],"externalURL": {"url": "%s"},"forkNamespace": null,"reviewState": "%s","checkState": null,"error": null,"syncerError": null,"forkName": null,"ownedByBatchChange": null}}}`,
				micID,
				ic.ExternalID,
				mbID,
				repoTwoID,
				icTitle,
				icBody,
				authorName,
				authorEmail,
				ic.State,
				icExternalURL,
				icReviewState,
			),
			want: &changeset{
				ID:                 micID,
				ExternalID:         uc.ExternalID,
				RepositoryID:       gql.MarshalRepositoryID(ic.RepoID),
				CreatedAt:          now,
				UpdatedAt:          now,
				BatchChangeIDs:     []graphql.ID{mbID},
				State:              string(ic.State),
				OwnedByBatchChange: nil,
				Title:              &icTitle,
				Body:               &icBody,
				AuthorName:         &authorName,
				AuthorEmail:        &authorEmail,
				ExternalURL:        &icExternalURL,
				ReviewState:        &icReviewState,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := new(mockDoer)
			client.On("Do", mock.Anything).Return(tc.httpResponse)

			response, err := marshalChangeset(ctx, client, bgql.MarshalChangesetID(tc.changeset.ID))
			require.NoError(t, err)

			var have = &changeset{}
			err = json.Unmarshal(response, have)
			require.NoError(t, err)

			cmpIgnored := cmpopts.IgnoreFields(changeset{}, "CreatedAt", "UpdatedAt")
			if diff := cmp.Diff(have, tc.want, cmpIgnored); diff != "" {
				t.Errorf("mismatched response from changeset marshal, got != want, diff(-got, +want):\n%s", diff)
			}

			client.AssertExpectations(t)
		})
	}
}

type mockDoer struct {
	mock.Mock
}

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(args.Get(0).(string))),
	}, nil
}

func addChangeset(t *testing.T, ctx context.Context, s *store.Store, c *btypes.Changeset, batchChange int64) {
	t.Helper()

	c.BatchChanges = append(c.BatchChanges, btypes.BatchChangeAssoc{BatchChangeID: batchChange})
	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatal(err)
	}
}
