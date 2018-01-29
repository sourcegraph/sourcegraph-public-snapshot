package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

func (*schemaResolver) IsUsernameAvailable(ctx context.Context, args *struct {
	Username string
}) (bool, error) {
	_, err := db.Users.GetByUsername(ctx, args.Username)
	if err != nil {
		if errcode.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
