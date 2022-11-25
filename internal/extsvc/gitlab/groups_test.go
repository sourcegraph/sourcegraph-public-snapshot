package gitlab

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListGroups(t *testing.T) {
	ctx := context.Background()
	mockedGroups := []*Group{
		{
			ID:       1,
			FullPath: "group1",
		},
	}

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPPaginatedResponse{
			pages: []*mockHTTPResponseBody{
				{responseBody: `[{"id": 1,"full_path": "group1"}]`},
				{responseBody: `[]`},
			},
		}

		groups, err := client.ListGroups(ctx)
		assert.NoError(t, err)
		assert.Equal(t, mockedGroups, groups.mustAll())
	})

	t.Run("malformed response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `this is not valid JSON`,
		}

		groups, err := client.ListGroups(ctx)
		assert.Error(t, err)
		assert.Nil(t, groups)
	})
}
