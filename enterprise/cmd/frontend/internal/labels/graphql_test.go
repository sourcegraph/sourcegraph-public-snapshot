package labels

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func init() {
	graphqlbackend.Labels = GraphQLResolver{}
}

type mockThread struct {
	graphqlbackend.Thread
	id int64
}

func (t mockThread) ID() graphql.ID { return graphqlbackend.MarshalThreadID(t.id) }
func (t mockThread) DBID() int64    { return t.id }
func (t mockThread) Labels(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.LabelConnection, error) {
	return GraphQLResolver{}.LabelsForLabelable(ctx, t.ID(), arg)
}
