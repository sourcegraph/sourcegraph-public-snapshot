package webhooks

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	gitlabwebhooks "github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

func TestGitLabHandler(t *testing.T) {
	repoName := "gitlab.com/ryanslade/ryan-test-private"

	db := database.NewMockDB()
	repos := database.NewMockRepoStore()
	repos.GetFirstRepoNameByCloneURLFunc.SetDefaultHook(func(ctx context.Context, s string) (api.RepoName, error) {
		return api.RepoName(repoName), nil
	})
	db.ReposFunc.SetDefaultReturn(repos)

	handler := NewGitLabHandler(db)
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

	if err := handler.handlePushEvent(context.Background(), &payload); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, repoName, updateQueued)
}
