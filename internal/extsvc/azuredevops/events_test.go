package azuredevops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseWebhookEvent(t *testing.T) {
	for key, tc := range map[string]struct {
		payload   string
		eventType string
		wantType  any
	}{
		"git.pullrequest.merged": {
			payload:   `{"eventType":"git.pullrequest.merged","pullrequest":{},"resource":{}}`,
			eventType: "git.pullrequest.merged",
			wantType:  &PullRequestMergedEvent{},
		},
		"git.pullrequest.approved": {
			payload:   `{"eventType":"git.pullrequest.updated","message":{"text":"bob approved pull request 10"},"pullrequest":{},"resource":{}}`,
			eventType: "git.pullrequest.updated",
			wantType:  &PullRequestApprovedEvent{},
		},
		"git.pullrequest.approved_with_suggestions": {
			payload:   `{"eventType":"git.pullrequest.updated","message":{"text":"bob has approved and left suggestions in pull request 10"},"pullrequest":{},"resource":{}}`,
			eventType: "git.pullrequest.updated",
			wantType:  &PullRequestApprovedWithSuggestionsEvent{},
		},
		"git.pullrequest.waiting_for_author": {
			payload:   `{"eventType":"git.pullrequest.updated","message":{"text":"bob is waiting for the author in pull request 10"},"pullrequest":{},"resource":{}}`,
			eventType: "git.pullrequest.updated",
			wantType:  &PullRequestWaitingForAuthorEvent{},
		},
		"git.pullrequest.rejected": {
			payload:   `{"eventType":"ggit.pullrequest.updated","message":{"text":"bob rejected pull request 10"},"pullrequest":{},"resource":{}}`,
			eventType: "git.pullrequest.updated",
			wantType:  &PullRequestRejectedEvent{},
		},
		"git.pullrequest.updated": {
			payload:   `{"eventType":"git.pullrequest.updated","pullrequest":{},"resource":{}}`,
			eventType: "git.pullrequest.updated",
			wantType:  &PullRequestUpdatedEvent{},
		},
	} {
		t.Run(key, func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				have, err := ParseWebhookEvent(AzureDevOpsEvent(tc.eventType), []byte(tc.payload))
				assert.Nil(t, err)
				assert.IsType(t, tc.wantType, have)
			})

			t.Run("invalid JSON", func(t *testing.T) {
				_, err := ParseWebhookEvent(AzureDevOpsEvent(key), []byte("invalid JSON"))
				assert.NotNil(t, err)
			})
		})
	}

	t.Run("unknown key", func(t *testing.T) {
		_, err := ParseWebhookEvent("invalid key", []byte("{}"))
		assert.NotNil(t, err)
	})
}
