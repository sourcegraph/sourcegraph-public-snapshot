package labels

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateLabel(ctx context.Context, arg *graphqlbackend.CreateLabelArgs) (graphqlbackend.Label, error) {
	project, err := graphqlbackend.ProjectByID(ctx, arg.Input.Project)
	if err != nil {
		return nil, err
	}

	label, err := dbLabels{}.Create(ctx, &dbLabel{
		ProjectID:   project.DBID(),
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
		Color:       arg.Input.Color,
	})
	if err != nil {
		return nil, err
	}
	return &gqlLabel{db: label}, nil
}

func (GraphQLResolver) UpdateLabel(ctx context.Context, arg *graphqlbackend.UpdateLabelArgs) (graphqlbackend.Label, error) {
	l, err := labelByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	label, err := dbLabels{}.Update(ctx, l.db.ID, dbLabelUpdate{
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
		Color:       arg.Input.Color,
	})
	if err != nil {
		return nil, err
	}
	return &gqlLabel{db: label}, nil
}

func (GraphQLResolver) DeleteLabel(ctx context.Context, arg *graphqlbackend.DeleteLabelArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlLabel, err := labelByID(ctx, arg.Label)
	if err != nil {
		return nil, err
	}
	return nil, dbLabels{}.DeleteByID(ctx, gqlLabel.db.ID)
}

func (GraphQLResolver) AddLabelsToLabelable(ctx context.Context, arg *graphqlbackend.AddRemoveLabelsToFromLabelableArgs) (graphqlbackend.Labelable, error) {
	return addRemoveLabelsToFromLabelable(ctx, arg.Labelable, arg.Labels, nil)
}

func (GraphQLResolver) RemoveLabelsFromLabelable(ctx context.Context, arg *graphqlbackend.AddRemoveLabelsToFromLabelableArgs) (graphqlbackend.Labelable, error) {
	return addRemoveLabelsToFromLabelable(ctx, arg.Labelable, nil, arg.Labels)
}

func addRemoveLabelsToFromLabelable(ctx context.Context, labelable graphql.ID, addLabels []graphql.ID, removeLabels []graphql.ID) (graphqlbackend.Labelable, error) {
	// 🚨 SECURITY: Any viewer can add/remove labels to/from a thread.
	thread, err := graphqlbackend.DiscussionThreadByID(ctx, labelable)
	if err != nil {
		return nil, err
	}

	if len(addLabels) > 0 {
		addLabelIDs, err := getLabelDBIDs(ctx, addLabels)
		if err != nil {
			return nil, err
		}
		if err := (dbLabelsObjects{}).AddLabelsToThread(ctx, thread.DBID(), addLabelIDs); err != nil {
			return nil, err
		}
	}

	if len(removeLabels) > 0 {
		removeLabelIDs, err := getLabelDBIDs(ctx, removeLabels)
		if err != nil {
			return nil, err
		}
		if err := (dbLabelsObjects{}).RemoveLabelsFromThread(ctx, thread.DBID(), removeLabelIDs); err != nil {
			return nil, err
		}
	}

	return thread, nil
}

func getLabelDBIDs(ctx context.Context, labels []graphql.ID) ([]int64, error) {
	dbIDs := make([]int64, len(labels))
	for i, labelID := range labels {
		// 🚨 SECURITY: Only organization members and site admins may create labels in an
		// organization. The labelByID function performs this check.
		label, err := labelByID(ctx, labelID)
		if err != nil {
			return nil, err
		}
		dbIDs[i] = label.db.ID
	}
	return dbIDs, nil
}
