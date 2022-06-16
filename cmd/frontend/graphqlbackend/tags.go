package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) SetTag(ctx context.Context, args *struct {
	Node    graphql.ID
	Tag     string
	Present bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may set tags.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	node, err := r.nodeByID(ctx, args.Node)
	if err != nil {
		return nil, err
	}
	user, ok := node.(*UserResolver)
	if !ok {
		return nil, errors.New("setting tags is only supported for users")
	}

	if err := r.db.Users().SetTag(ctx, user.user.ID, args.Tag, args.Present); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
