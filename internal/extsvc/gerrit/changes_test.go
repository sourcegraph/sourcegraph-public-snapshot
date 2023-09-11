package gerrit

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestClient_GetChange(t *testing.T) {
	cli, save := NewTestClient(t, "GetChange", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.GetChange(ctx, "I52bede3e6dd80b9048924d0416e5d1a7bf49cf5b")
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/GetChange.json", *update, resp)
}

func TestClient_GetChangeMultipleChanges(t *testing.T) {
	cli, save := NewTestClient(t, "GetChangeMultipleChanges", *update)
	defer save()

	ctx := context.Background()

	// In order to recreate this tests you need to push two changes to two different branches using the same Chande-Id.
	_, err := cli.GetChange(ctx, "I52bede3e6dd80b9048924d0416e5d1a7bf49cf5a")
	assert.NotNil(t, err)
	assert.True(t, errors.As(err, &MultipleChangesError{}))
}

func TestClient_WriteReviewComment(t *testing.T) {
	cli, save := NewTestClient(t, "WriteReviewComment", *update)
	defer save()

	ctx := context.Background()

	err := cli.WriteReviewComment(ctx, "I52bede3e6dd80b9048924d0416e5d1a7bf49cf5b", ChangeReviewComment{
		Message: "test message",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_AbandonChange(t *testing.T) {
	cli, save := NewTestClient(t, "AbandonChange", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.AbandonChange(ctx, "I4ae8b9886059252657eef100c74602251b544e82")
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "testdata/golden/AbandonChange.json", *update, resp)
}

func TestClient_DeleteChange(t *testing.T) {
	cli, save := NewTestClient(t, "DeleteChange", *update)
	defer save()

	ctx := context.Background()

	// A change can only be deleted once. To re-record this test, publish a new change and
	// update the change ID.
	// You will need the "delete own changes" permission in order to delete your change:
	// https://gerrit-review.googlesource.com/Documentation/access-control.html#category_delete_own_changes
	changeID := "I2e55bf947cc1fe96b2663f4d3fedaa992628f8d4"
	err := cli.DeleteChange(ctx, changeID)
	if err != nil {
		t.Fatal(err)
	}

	// Delete again to ensure that the change is not found.
	err = cli.DeleteChange(ctx, changeID)
	if err == nil {
		t.Fatal("expected error, but got nil")
	}
	assert.ErrorContains(t, err, "code=404")
}

func TestClient_SubmitChange(t *testing.T) {
	cli, save := NewTestClient(t, "SubmitChange", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.SubmitChange(ctx, "I4ae8b9886059252657eef100c74602251b544e82")
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "testdata/golden/SubmitChange.json", *update, resp)
}

func TestClient_RestoreChange(t *testing.T) {
	cli, save := NewTestClient(t, "RestoreChange", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.RestoreChange(ctx, "Ida085bb4e62b9adb5991496ab31987e45cfd5d62")
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "testdata/golden/RestoreChange.json", *update, resp)
}

func TestClient_SetReadyForReview(t *testing.T) {
	cli, save := NewTestClient(t, "SetReadyForReview", *update)
	defer save()

	ctx := context.Background()

	err := cli.SetReadyForReview(ctx, "Ibcbfdbb6e19cec62feabfec5da47251820947067")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_SetWIP(t *testing.T) {
	cli, save := NewTestClient(t, "SetWIP", *update)
	defer save()

	ctx := context.Background()

	err := cli.SetWIP(ctx, "Ibcbfdbb6e19cec62feabfec5da47251820947067")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetChangeReviews(t *testing.T) {
	cli, save := NewTestClient(t, "GetChangeReviews", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.GetChangeReviews(ctx, "Ic433e1f2e4edfebe4cf75b23ded032bb790d872a")
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "testdata/golden/GetChangeReviews.json", *update, resp)
}

func TestClient_MoveChange(t *testing.T) {
	cli, save := NewTestClient(t, "MoveChange", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.MoveChange(ctx, "I8a43a17e679cf4ee3ba862e875746be2ed2215ec", MoveChangePayload{
		DestinationBranch: "newest-batch",
	})
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "testdata/golden/MoveChange.json", *update, resp)
}

func TestClient_SetCommitMessage(t *testing.T) {
	cli, save := NewTestClient(t, "SetCommitMessage", *update)
	defer save()

	ctx := context.Background()

	err := cli.SetCommitMessage(ctx, "I8a43a17e679cf4ee3ba862e875746be2ed2215ec", SetCommitMessagePayload{
		Message: "New commit message\n\nChange-Id: I8a43a17e679cf4ee3ba862e875746be2ed2215ec\n",
	})
	if err != nil {
		t.Fatal(err)
	}
}
