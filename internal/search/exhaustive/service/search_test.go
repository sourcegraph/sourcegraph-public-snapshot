package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
)

func TestWrongUser(t *testing.T) {
	assert := require.New(t)

	userID1 := int32(1)
	userID2 := int32(2)

	ctx := actor.WithActor(context.Background(), actor.FromMockUser(userID1))

	newSearcher := FromSearchClient(client.NewStrictMockSearchClient())
	_, err := newSearcher.NewSearch(ctx, userID2, "foo")
	assert.Error(err)
}
