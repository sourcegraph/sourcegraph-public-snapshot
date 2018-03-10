package auth

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

func createOrUpdateUser(ctx context.Context, newOrUpdatedUser db.NewUser) (userID int32, err error) {
	usr, err := db.Users.GetByExternalID(ctx, newOrUpdatedUser.ExternalProvider, newOrUpdatedUser.ExternalID)
	if errcode.IsNotFound(err) {
		usr, err = db.Users.Create(ctx, newOrUpdatedUser)
		if err != nil {
			return 0, err
		}
	}
	if err != nil {
		return 0, err
	}

	// Update user in our DB if their profile info changed on the issuer. (Except username,
	// which the user is somewhat likely to want to control separately on Sourcegraph.)
	var userUpdate db.UserUpdate
	if usr.DisplayName != newOrUpdatedUser.DisplayName {
		userUpdate.DisplayName = &newOrUpdatedUser.DisplayName
	}
	if userUpdate != (db.UserUpdate{}) {
		if err := db.Users.Update(ctx, usr.ID, userUpdate); err != nil {
			return 0, err
		}
	}
	return usr.ID, nil
}
