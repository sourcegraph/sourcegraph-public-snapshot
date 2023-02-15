package author

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func GetChangesetAuthorForUser(ctx context.Context, userStore database.UserStore, userID int32) (author *batches.ChangesetSpecAuthor, err error) {

	userEmailStore := database.UserEmailsWith(userStore)

	email, _, err := userEmailStore.GetPrimaryEmail(ctx, userID)
	if errcode.IsNotFound(err) {
		// No match just means there's no author, so we'll return nil. It's not
		// an error, though.
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "getting user e-mail")
	}

	author = &batches.ChangesetSpecAuthor{Email: email}

	user, err := userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "getting user")
	}
	if user.DisplayName != "" {
		author.Name = user.DisplayName
	} else {
		author.Name = user.Username
	}

	return author, nil
}
