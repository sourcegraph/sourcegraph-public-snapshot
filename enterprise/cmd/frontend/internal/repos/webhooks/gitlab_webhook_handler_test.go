package webhooks

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	gitlabwebhooks "github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

func TestGitLabWebhookHandle(t *testing.T) {
	ctx := context.Background()

	repoName := "gitlab.com/ryanslade/ryan-test-private"
	handler := NewGitLabWebhookHandler()
	data, err := os.ReadFile("testdata/gitlab-push.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload gitlabwebhooks.PushEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}

	var updateQueued string
	repoupdater.MockEnqueueRepoUpdate = func(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error) {
		updateQueued = string(repo)
		return &protocol.RepoUpdateResponse{
			ID:   1,
			Name: string(repo),
		}, nil
	}
	t.Cleanup(func() { repoupdater.MockEnqueueRepoUpdate = nil })

	if err := handler.handle(ctx, &payload); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, repoName, updateQueued)
}
