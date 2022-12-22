package webhooks

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestBitbucketServerHandler(t *testing.T) {
	repoName := "bitbucket.sgdev.org/private/test-2020-06-01"
	handler := NewBitbucketServerHandler()
	data, err := os.ReadFile("testdata/bitbucket-server-push.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload bitbucketserver.PushEvent
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

func TestBitbucketServerNameFromEvent(t *testing.T) {
	makeEvent := func(name string, href string) *bitbucketserver.PushEvent {
		return &bitbucketserver.PushEvent{
			Repository: bitbucketserver.Repo{
				Links: bitbucketserver.RepoLinks{
					Clone: []bitbucketserver.Link{
						{
							Href: href,
							Name: name,
						},
					},
				},
			},
		}
	}

	tests := []struct {
		name    string
		event   *bitbucketserver.PushEvent
		want    api.RepoName
		wantErr error
	}{
		{
			name:  "valid event",
			event: makeEvent("ssh", "ssh://git@bitbucket.sgdev.org/private/test-2020-06-01"),
			want:  api.RepoName("bitbucket.sgdev.org/private/test-2020-06-01"),
		},
		{
			name:  "valid event with port",
			event: makeEvent("ssh", "ssh://git@bitbucket.sgdev.org:7999/private/test-2020-06-01.git"),
			want:  api.RepoName("bitbucket.sgdev.org/private/test-2020-06-01"),
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
			got, err := bitbucketServerNameFromEvent(tt.event)
			if tt.wantErr != nil {
				assert.EqualError(t, tt.wantErr, err.Error())
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
