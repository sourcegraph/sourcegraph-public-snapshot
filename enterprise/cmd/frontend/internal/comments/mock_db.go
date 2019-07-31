package comments

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type dbMocks struct {
	comments        mockComments
	commentsThreads mockCommentsThreads

	commentByGQLID  func(graphql.ID) (*dbComment, error)
	newGQLToComment func(*dbComment) (graphqlbackend.Comment, error)
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
