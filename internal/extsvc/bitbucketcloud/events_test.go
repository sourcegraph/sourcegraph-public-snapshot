package bitbucketcloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseWebhookEvent(t *testing.T) {
	for key, tc := range map[string]struct {
		payload  string
		wantType any
	}{
		"pullrequest:approved": {
			payload:  `{"approval":{},"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestApprovedEvent{},
		},
		"pullrequest:changes_request_created": {
			payload:  `{"changes_request":{},"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestChangesRequestCreatedEvent{},
		},
		"pullrequest:changes_request_removed": {
			payload:  `{"changes_request":{},"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestChangesRequestRemovedEvent{},
		},
		"pullrequest:comment_created": {
			payload:  `{"comment":{},"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestCommentCreatedEvent{},
		},
		"pullrequest:comment_deleted": {
			payload:  `{"comment":{},"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestCommentDeletedEvent{},
		},
		"pullrequest:comment_updated": {
			payload:  `{"comment":{},"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestCommentUpdatedEvent{},
		},
		"pullrequest:fulfilled": {
			payload:  `{"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestFulfilledEvent{},
		},
		"pullrequest:rejected": {
			payload:  `{"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestRejectedEvent{},
		},
		"pullrequest:unapproved": {
			payload:  `{"approval":{},"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestUnapprovedEvent{},
		},
		"pullrequest:updated": {
			payload:  `{"pullrequest":{},"repository":{}}`,
			wantType: &PullRequestUpdatedEvent{},
		},
		"repo:commit_status_created": {
			payload:  `{"commit_status":{},"pullrequest":{},"repository":{}}`,
			wantType: &RepoCommitStatusCreatedEvent{},
		},
		"repo:commit_status_updated": {
			payload:  `{"commit_status":{},"pullrequest":{},"repository":{}}`,
			wantType: &RepoCommitStatusUpdatedEvent{},
		},
	} {
		t.Run(key, func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				have, err := ParseWebhookEvent(key, []byte(tc.payload))
				assert.Nil(t, err)
				assert.IsType(t, tc.wantType, have)
			})

			t.Run("invalid JSON", func(t *testing.T) {
				_, err := ParseWebhookEvent(key, []byte("invalid JSON"))
				assert.NotNil(t, err)
			})
		})
	}

	t.Run("unknown key", func(t *testing.T) {
		_, err := ParseWebhookEvent("invalid key", []byte("{}"))
		assert.NotNil(t, err)
	})
}
