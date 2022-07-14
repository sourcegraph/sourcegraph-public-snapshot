package githubwebhook

import (
	"context"
	"flag"
	"os"
	"testing"
)

var repoName = "ghe.sgdev.org/milton/test"
var update = flag.Bool("update", false, "update github webhooks API testdata")

func TestGitHubWebhooks_List(t *testing.T) {
	ctx := context.Background()
	token := os.Getenv("ACCESS_TOKEN")

	gh, err := NewGithubWebhookAPI()
	if err != nil {
		t.Fatal(err)
	}
	gh.client = NewTestClient(t, "List", update)

	_, err = gh.ListSyncWebhooks(ctx, repoName, token)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGitHubWebhooks_CreateListDelete(t *testing.T) {
	ctx := context.Background()
	token := os.Getenv("ACCESS_TOKEN")

	gh, err := NewGithubWebhookAPI()
	if err != nil {
		t.Fatal(err)
	}
	gh.client = NewTestClient(t, "CreateListDelete", update)

	id, err := gh.CreateSyncWebhook(ctx, repoName, "https://repoupdater.com", "secret", token)
	if err != nil {
		t.Fatal(err)
	}

	payloads, err := gh.ListSyncWebhooks(ctx, repoName, token)
	if err != nil {
		t.Fatal(err)
	}

	idCount := 0
	for _, payload := range payloads {
		if payload.ID == id {
			idCount++
		}
	}

	// even if there is an error,
	// we want to delete the webhook we just created
	countErr := ""
	if idCount != 1 {
		countErr = "Created webhook more not equal to 1"
	}

	deleted, err := gh.DeleteSyncWebhook(ctx, repoName, id, token)
	if err != nil {
		t.Fatal(err)
	}

	if countErr != "" {
		t.Fatal(countErr)
	}

	if !deleted {
		t.Fatalf("Failed to delete recently created webhook")
	}
}

func TestGitHubWebhooks_Find(t *testing.T) {
	ctx := context.Background()
	token := os.Getenv("ACCESS_TOKEN")

	gh, err := NewGithubWebhookAPI()
	if err != nil {
		t.Fatal(err)
	}
	gh.client = NewTestClient(t, "Find", update)

	found := gh.FindSyncWebhook(ctx, repoName, token)
	if !found {
		t.Fatalf("Could not find webhook")
	}
}

// assert object ID == some number
// subsequent queries are not going to make an actual call
