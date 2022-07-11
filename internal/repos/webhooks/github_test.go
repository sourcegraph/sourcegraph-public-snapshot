package githubwebhook

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var repoName = "susantoscott/Task-Tracker"

func TestGitHubWebhooks_CreateAndList(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}
	token := os.Getenv("ACCESS_TOKEN")

	id, err := CreateSyncWebhook(repoName, "secret", token)
	if err != nil {
		t.Fatal(err)
	}

	payloads, err := ListSyncWebhooks(repoName, token)
	if err != nil {
		t.Fatal(err)
	}

	idCount := 0
	for _, payload := range payloads {
		if payload.ID == id {
			idCount++
		}
	}

	countErr := ""
	if idCount != 1 {
		countErr = "Created webhook more not equal to 1"
	}

	deleted, err := DeleteSyncWebhook(repoName, id, token)
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
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}
	token := os.Getenv("ACCESS_TOKEN")

	found := FindSyncWebhook(repoName, token)
	if !found {
		t.Fatalf("Could not find webhook")
	}
}
