package auth

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

// MockCreateOrUpdateUser is used in tests to mock CreateOrUpdateUser.
var MockCreateOrUpdateUser func(db.NewUser, db.ExternalAccountSpec) (int32, error)

// CreateOrUpdateUser creates or updates a user using the provided information, looking up a user by
// the external account provided.
//
// 🚨 SECURITY: The safeErrMsg is an error message that can be shown to unauthenticated users to
// describe the problem. The err may contain sensitive information and should only be written to the
// server error logs, not to the HTTP response to shown to unauthenticated users.
func CreateOrUpdateUser(ctx context.Context, newOrUpdatedUser db.NewUser, externalAccount db.ExternalAccountSpec) (userID int32, safeErrMsg string, err error) {
	if MockCreateOrUpdateUser != nil {
		userID, err = MockCreateOrUpdateUser(newOrUpdatedUser, externalAccount)
		return userID, "", err
	}

	// TEMPORARY: Copy external account info to db.NewUser (soon these fields will be removed from
	// db.NewUser).
	newOrUpdatedUser.ExternalProvider = externalAccount.ServiceID
	newOrUpdatedUser.ExternalID = externalAccount.AccountID

	usr, err := db.Users.GetByExternalID(ctx, newOrUpdatedUser.ExternalProvider, newOrUpdatedUser.ExternalID)
	if err != nil {
		if !errcode.IsNotFound(err) {
			return 0, "Unexpected error looking up the Sourcegraph user account associated with the external account. Ask a site admin for help.", err
		}

		// Try creating user.
		usr, err = db.Users.Create(ctx, newOrUpdatedUser)

		// Handle the race condition where the new user performs two requests and both try to create
		// the user.
		if err != nil {
			// If GetByExternalID fails, return original Create error (err); otherwise clear the error.
			var err2 error
			usr, err2 = db.Users.GetByExternalID(ctx, newOrUpdatedUser.ExternalProvider, newOrUpdatedUser.ExternalID)
			if err2 == nil {
				err = nil
			}
		}
		switch {
		case db.IsUsernameExists(err):
			safeErrMsg = fmt.Sprintf("The username %q already exists and is linked to a different external account. A site admin can unlink or delete the Sourcegraph account with that username to fix this problem.", newOrUpdatedUser.Username)
		case db.IsEmailExists(err):
			safeErrMsg = fmt.Sprintf("The email address %q already exists and is associated with a different Sourcegraph user. A site admin can remove the email address from that Sourcegraph user to fix this problem.", newOrUpdatedUser.Email)
		default:
			safeErrMsg = "Unable to create a new user account due to a conflict or other unexpected error. Ask a site admin for help."
		}
		if err != nil {
			return 0, safeErrMsg, err
		}
	}

	// Update user in our DB if their profile info changed on the issuer. (Except username,
	// which the user is somewhat likely to want to control separately on Sourcegraph.)
	var userUpdate db.UserUpdate
	if usr.DisplayName != newOrUpdatedUser.DisplayName {
		userUpdate.DisplayName = &newOrUpdatedUser.DisplayName
	}
	if usr.AvatarURL != newOrUpdatedUser.AvatarURL {
		userUpdate.AvatarURL = &newOrUpdatedUser.AvatarURL
	}
	if userUpdate != (db.UserUpdate{}) {
		if err := db.Users.Update(ctx, usr.ID, userUpdate); err != nil {
			return 0, "Unexpected error updating the Sourcegraph user account with new user profile information from the external account. Ask a site admin for help.", err
		}
	}
	return usr.ID, "", nil
}
