package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

func (*schemaResolver) IsUsernameAvailable(ctx context.Context, args *struct {
	Username string
}) (bool, error) {
	_, err := db.Users.GetByUsername(ctx, args.Username)
	if err != nil {
		if _, ok := err.(db.ErrUserNotFound); ok {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
