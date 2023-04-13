package scim

import (
	"context"
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"k8s.io/utils/strings/slices"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type updateAction interface {
	Update(ctx context.Context, before, after *scim.Resource) error
}

type userEntityUpdate struct {
	db   database.DB
	user *User
}

func (u *userEntityUpdate) Update(ctx context.Context, before, after *scim.Resource) error {
	if u.user == nil {
		return errors.New("invalid user")
	}
	err := u.db.WithTransact(ctx, func(tx database.DB) error {
		// build the list of actions
		// The order is important as some actions may make the user (not) visible to future actions
		actions := []updateAction{
			&activateUser{userID: u.user.ID, tx: tx}, // This is intentionally first so that user record to ready for other attribute changes
			&updateUserProfileData{userID: u.user.ID, userName: u.user.Username, tx: tx},
			&updateUserEmailData{userID: u.user.ID, tx: tx, db: u.db},
			&updateUserExternalAccountData{userID: u.user.ID, tx: tx},
			&softDeleteUser{user: u.user, tx: tx}, // This is intentionally last so that other attribute changes are captured
		}

		// run each action and quit if one fails
		for _, action := range actions {
			err := action.Update(ctx, before, after)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func NewUserUpdate(db database.DB, user *User) updateAction {
	return &userEntityUpdate{db: db, user: user}
}

type updateUserProfileData struct {
	userID   int32
	userName string
	tx       database.DB
}

func (u *updateUserProfileData) Update(ctx context.Context, before, after *scim.Resource) error {
	// Check if changed occurred
	requestedUsername := extractStringAttribute(after.Attributes, AttrUserName)
	if requestedUsername == u.userName {
		return nil
	}

	usernameUpdate, err := getUniqueUsername(ctx, u.tx.Users(), extractStringAttribute(after.Attributes, AttrUserName))
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusBadRequest, Detail: errors.Wrap(err, "invalid username").Error()}
	}
	var displayNameUpdate *string
	var avatarURLUpdate *string
	userUpdate := database.UserUpdate{
		Username:    usernameUpdate,
		DisplayName: displayNameUpdate,
		AvatarURL:   avatarURLUpdate,
	}
	err = u.tx.Users().Update(ctx, u.userID, userUpdate)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}
	return nil
}

type updateUserExternalAccountData struct {
	userID int32
	tx     database.DB
}

func (u *updateUserExternalAccountData) Update(ctx context.Context, before, after *scim.Resource) error {
	// No check for changes always write the latest SCIM data to db

	accountData, err := toAccountData(after.Attributes)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	err = u.tx.UserExternalAccounts().UpsertSCIMData(ctx, u.userID, getUniqueExternalID(after.Attributes), accountData)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}
	return nil
}

type updateUserEmailData struct {
	tx     database.DB
	db     database.DB
	userID int32
}

func (u *updateUserEmailData) changed(before, after *scim.Resource) bool {
	primaryBefore, otherEmailsBefore := extractPrimaryEmail(before.Attributes)
	primaryAfter, otherEmailsAfter := extractPrimaryEmail(after.Attributes)

	// Check primary emails
	if primaryAfter != primaryBefore {
		return true
	}
	// Check rest of the emails
	return !slices.Equal(otherEmailsBefore, otherEmailsAfter)

}
func (u *updateUserEmailData) Update(ctx context.Context, before, after *scim.Resource) error {
	// Only update if emails changed
	if !u.changed(before, after) {
		return nil
	}
	currentEmails, err := u.tx.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{UserID: u.userID, OnlyVerified: false})
	if err != nil {
		return err
	}
	diffs := diffEmails(before.Attributes, after.Attributes, currentEmails)
	// First add any new email address
	for _, newEmail := range diffs.toAdd {
		err = u.tx.UserEmails().Add(ctx, u.userID, newEmail, nil)
		if err != nil {
			return err
		}
		err = u.tx.UserEmails().SetVerified(ctx, u.userID, newEmail, true)
		if err != nil {
			return err
		}
	}

	// Now verify any addresses that already existed but weren't verified
	for _, email := range diffs.toVerify {
		err = u.tx.UserEmails().SetVerified(ctx, u.userID, email, true)
		if err != nil {
			return err
		}
	}

	// Now that all the new emails are added and verified set the primary email if it changed
	if diffs.setPrimaryEmailTo != nil {
		err = u.tx.UserEmails().SetPrimaryEmail(ctx, u.userID, *diffs.setPrimaryEmailTo)
		if err != nil {
			return err
		}
	}

	// Finally remove any email addresses that no longer are needed
	for _, email := range diffs.toRemove {
		err = u.tx.UserEmails().Remove(ctx, u.userID, email)
		if err != nil {
			return err
		}
	}
	return nil
}

// Action to delete the user when SCIM changes the active flag to "false"
// This is a temporary action that will be replaced when soft delete is supported
type hardDeleteInactiveUser struct {
	user *User
	tx   database.DB
}

func (u *hardDeleteInactiveUser) Update(ctx context.Context, before, after *scim.Resource) error {
	// Check if user has been deactivated
	if after.Attributes[AttrActive] != false {
		return nil
	}
	// Save username, verified emails, and external accounts to be used for revoking user permissions after deletion
	revokeUserPermissionsArgsList, err := getRevokeUserPermissionArgs(ctx, u.user.UserForSCIM, u.tx)
	if err != nil {
		return err
	}

	if err := u.tx.Users().HardDelete(ctx, u.user.ID); err != nil {
		return err
	}

	// NOTE: Practically, we don't reuse the ID for any new users, and the situation of left-over pending permissions
	// is possible but highly unlikely. Therefore, there is no need to roll back user deletion even if this step failed.
	// This call is purely for the purpose of cleanup.
	err = u.tx.Authz().RevokeUserPermissionsList(ctx, []*database.RevokeUserPermissionsArgs{revokeUserPermissionsArgsList})

	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}
	return nil
}

// Action to soft delete the user when SCIM changes the active flag to "false"
type softDeleteUser struct {
	user *User
	tx   database.DB
}

func (u *softDeleteUser) Update(ctx context.Context, before, after *scim.Resource) error {
	// Check if user active went from true -> false
	if !(before.Attributes[AttrActive] == true && after.Attributes[AttrActive] == false) {
		return nil
	}

	if err := u.tx.Users().Delete(ctx, u.user.ID); err != nil {
		return err
	}

	return nil
}

// Action to reactivate the user when SCIM changes the active flag to "true"
type activateUser struct {
	userID int32
	tx     database.DB
}

func (u *activateUser) Update(ctx context.Context, before, after *scim.Resource) error {
	// Check moved from active false -> true
	if !(before.Attributes[AttrActive] == false && after.Attributes[AttrActive] == true) {
		return nil
	}

	recoveredIDs, err := u.tx.Users().RecoverUsersList(ctx, []int32{u.userID})
	if err != nil {
		return err
	}

	if len(recoveredIDs) != 1 {
		return errors.New("unable to activate user")
	}

	return nil
}
