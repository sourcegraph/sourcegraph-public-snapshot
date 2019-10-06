package comments

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func CommentActorFromContext(ctx context.Context) (actor.DBColumns, error) {
	user, err := graphqlbackend.CurrentUser(ctx)
	if err != nil {
		return actor.DBColumns{}, err
	}
	if user == nil {
		return actor.DBColumns{}, errors.New("authenticated required to create comment")
	}
	return actor.DBColumns{UserID: user.DatabaseID()}, nil
}
