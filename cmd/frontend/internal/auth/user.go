package auth

import (
	"context"
	"fmt"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
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
func CreateOrUpdateUser(ctx context.Context, newOrUpdatedUser db.NewUser, externalAccount db.ExternalAccountSpec, data db.ExternalAccountData) (userID int32, safeErrMsg string, err error) {
	if MockCreateOrUpdateUser != nil {
		userID, err = MockCreateOrUpdateUser(newOrUpdatedUser, externalAccount)
		return userID, "", err
	}

	if actor := actor.FromContext(ctx); actor.IsAuthenticated() {
		// There is already an authenticated actor, so this external account will be added to
		// the existing user account.
		userID = actor.UID

		if err := db.ExternalAccounts.AssociateUserAndSave(ctx, userID, externalAccount, data); err != nil {
			safeErrMsg = "Unexpected error associating the external account with your Sourcegraph user. The most likely cause for this problem is that another Sourcegraph user is already linked with this external account. A site admin or the other user can unlink the account to fix this problem."
			return 0, safeErrMsg, err
		}
	} else {
		userID, err = db.ExternalAccounts.LookupUserAndSave(ctx, externalAccount, data)
		if err != nil && !errcode.IsNotFound(err) {
			return 0, "Unexpected error looking up the Sourcegraph user account associated with the external account. Ask a site admin for help.", err
		}
		if errcode.IsNotFound(err) {
			// Looser requirements: if the external auth provider returns a username or email that
			// already exists, just use that user instead of refusing.
			const allowMatchOnUsernameOrEmailOnly = true
			associateUser := false

			userID, err = db.ExternalAccounts.CreateUserAndSave(ctx, newOrUpdatedUser, externalAccount, data)
			switch {
			case db.IsUsernameExists(err):
				if allowMatchOnUsernameOrEmailOnly {
					user, err2 := db.Users.GetByUsername(ctx, newOrUpdatedUser.Username)
					if err2 == nil {
						userID = user.ID
						err = nil
						associateUser = true
					} else {
						log15.Error("Unable to reuse user account with username for authentication via external provider.", "username", newOrUpdatedUser.Username, "err", err)
					}
				}
				safeErrMsg = fmt.Sprintf("The Sourcegraph username %q already exists and is not linked to this external account. If possible, sign using the external account you used previously. If that's not possible, a site admin can unlink or delete the Sourcegraph account with that username to fix this problem.", newOrUpdatedUser.Username)
			case db.IsEmailExists(err):
				if allowMatchOnUsernameOrEmailOnly {
					user, err2 := db.Users.GetByVerifiedEmail(ctx, newOrUpdatedUser.Email)
					if err2 == nil {
						userID = user.ID
						err = nil
						associateUser = true
					} else {
						log15.Error("Unable to reuse user account with email for authentication via external provider.", "email", newOrUpdatedUser.Email, "err", err)
					}
				}
				safeErrMsg = fmt.Sprintf("The email address %q already exists and is associated with a different Sourcegraph user. A site admin can remove the email address from that Sourcegraph user to fix this problem.", newOrUpdatedUser.Email)
			case errcode.PresentationMessage(err) != "":
				safeErrMsg = errcode.PresentationMessage(err)
			case err != nil:
				safeErrMsg = "Unable to create a new user account due to a conflict or other unexpected error. Ask a site admin for help."
			}

			if associateUser {
				if err := db.ExternalAccounts.AssociateUserAndSave(ctx, userID, externalAccount, data); err != nil {
					safeErrMsg = "Unexpected error associating the external account with the existing Sourcegraph user with the same username or email address."
					return 0, safeErrMsg, err
				}
			}

			return userID, safeErrMsg, err
		}
	}

	// Update user in our DB if their profile info changed on the issuer. (Except username,
	// which the user is somewhat likely to want to control separately on Sourcegraph.)
	user, err := db.Users.GetByID(ctx, userID)
	if err != nil {
		return 0, "Unexpected error getting the Sourcegraph user account. Ask a site admin for help.", err
	}
	var userUpdate db.UserUpdate
	if user.DisplayName != newOrUpdatedUser.DisplayName {
		userUpdate.DisplayName = &newOrUpdatedUser.DisplayName
	}
	if user.AvatarURL != newOrUpdatedUser.AvatarURL {
		userUpdate.AvatarURL = &newOrUpdatedUser.AvatarURL
	}
	if userUpdate != (db.UserUpdate{}) {
		if err := db.Users.Update(ctx, user.ID, userUpdate); err != nil {
			return 0, "Unexpected error updating the Sourcegraph user account with new user profile information from the external account. Ask a site admin for help.", err
		}
	}
	return user.ID, "", nil
}
