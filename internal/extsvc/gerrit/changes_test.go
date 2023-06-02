package gerrit

import (
	"context"
	"testing"

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
