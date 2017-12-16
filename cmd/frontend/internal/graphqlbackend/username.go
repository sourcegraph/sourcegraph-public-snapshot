package graphqlbackend

import (
	"context"

	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func (*schemaResolver) IsUsernameAvailable(ctx context.Context, args *struct {
	Username string
}) (bool, error) {
	_, err := store.Users.GetByUsername(ctx, args.Username)
	if err != nil {
		if _, ok := err.(store.ErrUserNotFound); ok {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
