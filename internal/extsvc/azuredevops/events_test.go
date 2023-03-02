package azuredevops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseWebhookEvent(t *testing.T) {
	for key, tc := range map[string]struct {
		payload  string
		wantType any
	}{
		"git.pullrequest.created": {
			payload:  `{"eventType":"git.pullrequest.created","pullrequest":{},"resource":{}}`,
			wantType: &PullRequestCreatedEvent{},
		},
		"git.pullrequest.merged": {
			payload:  `{"eventType":"git.pullrequest.merged","pullrequest":{},"resource":{}}`,
			wantType: &PullRequestMergedEvent{},
		},
		"git.pullrequest.updated": {
			payload:  `{"eventType":"git.pullrequest.updated","pullrequest":{},"resource":{}}`,
			wantType: &PullRequestUpdatedEvent{},
		},
	} {
		t.Run(key, func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				have, err := ParseWebhookEvent(AzureDevOpsEvent(key), []byte(tc.payload))
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
