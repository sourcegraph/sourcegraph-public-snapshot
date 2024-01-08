package actor

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

func TestActor_TraceAttributes(t *testing.T) {
	tests := []struct {
		name     string
		actor    *Actor
		wantAttr autogold.Value
	}{
		{
			name:     "nil actor",
			actor:    nil,
			wantAttr: autogold.Expect(`[{"Key":"actor","Value":{"Type":"STRING","Value":"\u003cnil\u003e"}}]`),
		},
		{
			name: "with ID and access enabled",
			actor: &Actor{
				ID:            "abc123",
				AccessEnabled: true,
			},
			wantAttr: autogold.Expect(`[{"Key":"actor.id","Value":{"Type":"STRING","Value":"abc123"}},{"Key":"actor.accessEnabled","Value":{"Type":"BOOL","Value":true}}]`),
		},
		{
			name: "with rate limits",
			actor: &Actor{
				ID: "abc123",
				RateLimits: map[codygateway.Feature]RateLimit{
					codygateway.FeatureCodeCompletions: {
						Limit: 50,
					},
					codygateway.FeatureEmbeddings: {
						Limit: 50,
					},
				},
			},
			wantAttr: autogold.Expect(`[{"Key":"actor.rateLimits.embeddings","Value":{"Type":"STRING","Value":"{\"allowedModels\":null,\"limit\":50,\"interval\":0,\"concurrentRequests\":0,\"concurrentRequestsInterval\":0}"}},{"Key":"actor.rateLimits.code_completions","Value":{"Type":"STRING","Value":"{\"allowedModels\":null,\"limit\":50,\"interval\":0,\"concurrentRequests\":0,\"concurrentRequestsInterval\":0}"}},{"Key":"actor.id","Value":{"Type":"STRING","Value":"abc123"}},{"Key":"actor.accessEnabled","Value":{"Type":"BOOL","Value":false}}]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAttr := tt.actor.TraceAttributes()
			sort.Slice(gotAttr, func(i, j int) bool {
				return string(gotAttr[i].Key) > string(gotAttr[j].Key)
			})
			// Just a sanity check, keep in one line for test stability
			b, err := json.Marshal(gotAttr)
			require.NoError(t, err)
			tt.wantAttr.Equal(t, string(b))
		})
	}
}

func TestIsDotComActor(t *testing.T) {
	t.Run("true for dotcom subscription", func(t *testing.T) {
		actor := &Actor{
			ID:     "d3d2b638-d0a2-4539-a099-b36860b09819",
			Source: FakeSource{codygateway.ActorSourceProductSubscription},
		}
		require.True(t, actor.IsDotComActor())
	})

	t.Run("true for dotcom user", func(t *testing.T) {
		actor := &Actor{
			Source: FakeSource{codygateway.ActorSourceDotcomUser},
		}
		require.True(t, actor.IsDotComActor())
	})

	t.Run("false for other subscription", func(t *testing.T) {
		actor := &Actor{
			ID:     "other-sub-id",
			Source: FakeSource{codygateway.ActorSourceProductSubscription},
		}
		require.False(t, actor.IsDotComActor())
	})

	t.Run("false for nil", func(t *testing.T) {
		var actor *Actor = nil
		require.False(t, actor.IsDotComActor())
	})
}
