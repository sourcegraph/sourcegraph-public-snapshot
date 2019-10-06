package labels

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func (GraphQLResolver) LabelsForLabelable(ctx context.Context, labelable graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.LabelConnection, error) {
	// 🚨 SECURITY: Any viewer can add/remove labels to/from a thread.
	thread, err := threads.GraphQLResolver{}.ThreadByID(ctx, labelable)
	if err != nil {
		return nil, err
	}
	threadDBID, err := graphqlbackend.UnmarshalThreadID(thread.ID())
	if err != nil {
		return nil, err
	}

	list, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{ThreadID: threadDBID})
	if err != nil {
		return nil, err
	}

	labels := make([]*gqlLabel, len(list))
	for i, l := range list {
		label, err := labelByDBID(ctx, l.Label)
		if err != nil {
			return nil, err
		}
		labels[i] = label
	}
	return &labelConnection{arg: arg, labels: toLabels(labels)}, nil
}

func (GraphQLResolver) LabelsInRepository(ctx context.Context, repositoryID graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.LabelConnection, error) {
	// Check existence.
	repository, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	list, err := dbLabels{}.List(ctx, dbLabelsListOptions{RepositoryID: int64(repository.DBID())})
	if err != nil {
		return nil, err
	}
	labels := make([]*gqlLabel, len(list))
	for i, a := range list {
		labels[i] = &gqlLabel{db: a}
	}
	return &labelConnection{arg: arg, labels: toLabels(labels)}, nil
}

func ConstLabelConnection(labels []graphqlbackend.Label) graphqlbackend.LabelConnection {
	return &labelConnection{arg: &graphqlutil.ConnectionArgs{}, labels: labels}
}

type labelConnection struct {
	arg    *graphqlutil.ConnectionArgs
	labels []graphqlbackend.Label
}

func (r *labelConnection) Nodes(ctx context.Context) ([]graphqlbackend.Label, error) {
	labels := r.labels
	if first := r.arg.First; first != nil && len(labels) > int(*first) {
		labels = labels[:int(*first)]
	}
	return labels, nil
}

func toLabels(gqlLabels []*gqlLabel) []graphqlbackend.Label {
	labels := make([]graphqlbackend.Label, len(gqlLabels))
	for i, l := range gqlLabels {
		labels[i] = l
	}
	return labels
}

func (r *labelConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.labels)), nil
}

func (r *labelConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.labels)), nil
}
