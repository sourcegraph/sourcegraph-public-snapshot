package resolvers

import (
	"encoding/json"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type ExportedEventResolver struct{}

var _ graphqlbackend.ExportedEventResolver = &ExportedEventResolver{}

func (r *ExportedEventResolver) ID() graphql.ID {
	return ""
}

func (r *ExportedEventResolver) ExportedAt() *graphql.Time {
	return nil
}

func (r *ExportedEventResolver) Payload() json.RawMessage {
	return nil
}

type ExportedEventsConnectionResolver struct{}

var _ graphqlbackend.ExportedEventsConnectionResolver = &ExportedEventsConnectionResolver{}

func (r *ExportedEventsConnectionResolver) Nodes() []graphqlbackend.ExportedEventResolver {
	return nil
}

func (r *ExportedEventsConnectionResolver) TotalCount() int32 {
	return 0
}

func (r *ExportedEventsConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return nil
}
