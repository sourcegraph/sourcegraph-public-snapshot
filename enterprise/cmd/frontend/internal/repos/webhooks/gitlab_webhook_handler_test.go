package webhooks

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	gitlabwebhooks "github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitLabWebhookHandle(t *testing.T) {
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

	if err := handler.handlePushEvent(context.Background(), &payload); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, repoName, updateQueued)
}

func TestGitlabNameFromEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   *gitlabwebhooks.PushEvent
		want    api.RepoName
		wantErr error
	}{
		{
			name: "valid event",
			event: &gitlabwebhooks.PushEvent{
				EventCommon: gitlabwebhooks.EventCommon{
					Project: gitlab.ProjectCommon{
						WebURL: "https://gitlab.com/sourcegraph/sourcegraph",
					},
				},
			},
			want: api.RepoName("gitlab.com/sourcegraph/sourcegraph"),
		},
		{
			name:    "nil event",
			event:   nil,
			want:    api.RepoName(""),
			wantErr: errors.New("nil PushEvent received"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gitlabNameFromEvent(tt.event)
			if tt.wantErr != nil {
				assert.EqualError(t, tt.wantErr, err.Error())
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
