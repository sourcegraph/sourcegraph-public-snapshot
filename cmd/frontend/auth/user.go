package auth

import (
	"context"
	"encoding/json"
	"fmt"

	sglog "github.com/sourcegraph/log"

	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var MockGetAndSaveUser func(ctx context.Context, op GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error)

type GetAndSaveUserOp struct {
	UserProps           database.NewUser
	ExternalAccount     extsvc.AccountSpec
	ExternalAccountData extsvc.AccountData
	CreateIfNotExist    bool
	LookUpByUsername    bool
}

// GetAndSaveUser accepts authentication information associated with a given user, validates and applies
// the necessary updates to the DB, and returns the user ID after the updates have been applied.
//
// At a high level, it does the following:
//  1. Determine the identity of the user by applying the following rules in order:
//     a. If ctx contains an authenticated Actor, the Actor's identity is the user identity.
//     b. Look up the user by external account ID.
//     c. If the email specified in op.UserProps is verified, Look up the user by verified email.
//     If op.LookUpByUsername is true, look up by username instead of verified email.
//     (Note: most clients should look up by email, as username is typically insecure.)
//     d. If op.CreateIfNotExist is true, attempt to create a new user with the properties
//     specified in op.UserProps. This may fail if the desired username is already taken.
//     e. If a new user is successfully created, attempt to grant pending permissions.
//  2. Ensure that the user is associated with the external account information. This means
//     creating the external account if it does not already exist or updating it if it
//     already does.
//  3. Update any user props that have changed.
//  4. Return the user ID.
//
// 🚨 SECURITY: It is the caller's responsibility to ensure the veracity of the information that
// op contains (e.g., by receiving it from the appropriate authentication mechanism). It must
// also ensure that the user identity implied by op is consistent. Specifically, the values used
// in step 1 above must be consistent:
// * The authenticated Actor, if it exists
// * op.ExternalAccount
// * op.UserProps, especially op.UserProps.Email
//
// 🚨 SECURITY: The safeErrMsg is an error message that can be shown to unauthenticated users to
// describe the problem. The err may contain sensitive information and should only be written to the
// server error logs, not to the HTTP response to shown to unauthenticated users.
func GetAndSaveUser(ctx context.Context, db database.DB, op GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
	if MockGetAndSaveUser != nil {
		return MockGetAndSaveUser(ctx, op)
	}

	externalAccountsStore := db.UserExternalAccounts()
	logger := sglog.Scoped("authGetAndSaveUser", "get and save user authenticated by external providers")

	userID, userSaved, extAcctSaved, safeErrMsg, err := getAndSaveUser(ctx, db, op)
	if err != nil {
		const eventName = "ExternalAuthSignupFailed"
		serviceTypeArg := json.RawMessage(fmt.Sprintf(`{"serviceType": %q}`, op.ExternalAccount.ServiceType))
		if logErr := usagestats.LogBackendEvent(db, sgactor.FromContext(ctx).UID, deviceid.FromContext(ctx), eventName, serviceTypeArg, serviceTypeArg, featureflag.GetEvaluatedFlagSet(ctx), nil); logErr != nil {
			logger.Error(
				"failed to log event",
				sglog.String("eventName", eventName),
				sglog.Error(err),
			)
		}
		return newUserCreated, 0, safeErrMsg, err
	}

	// Update user properties, if they've changed
	if !userSaved {
		// Update user in our DB if their profile info changed on the issuer. (Except username and
		// email, which the user is somewhat likely to want to control separately on Sourcegraph.)
		user, err := db.Users().GetByID(ctx, userID)
		if err != nil {
			return newUserCreated, 0, "Unexpected error getting the Sourcegraph user account. Ask a site admin for help.", err
		}
		var userUpdate database.UserUpdate
		if user.DisplayName == "" && op.UserProps.DisplayName != "" {
			userUpdate.DisplayName = &op.UserProps.DisplayName
		}
		if user.AvatarURL == "" && op.UserProps.AvatarURL != "" {
			userUpdate.AvatarURL = &op.UserProps.AvatarURL
		}
		if userUpdate != (database.UserUpdate{}) {
			if err := db.Users().Update(ctx, user.ID, userUpdate); err != nil {
				return newUserCreated, 0, "Unexpected error updating the Sourcegraph user account with new user profile information from the external account. Ask a site admin for help.", err
			}
		}
	}

	// Create/update the external account and ensure it's associated with the user ID
	if !extAcctSaved {
		err := externalAccountsStore.AssociateUserAndSave(ctx, userID, op.ExternalAccount, op.ExternalAccountData)
		if err != nil {
			return newUserCreated, 0, "Unexpected error associating the external account with your Sourcegraph user. The most likely cause for this problem is that another Sourcegraph user is already linked with this external account. A site admin or the other user can unlink the account to fix this problem.", err
		}

		// Schedule a permission sync, since this is probably a new external account for the user
		permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
			UserIDs:           []int32{userID},
			Reason:            database.ReasonExternalAccountAdded,
			TriggeredByUserID: userID,
		})

		if err = db.Authz().GrantPendingPermissions(ctx, &database.GrantPendingPermissionsArgs{
			UserID: userID,
			Perm:   authz.Read,
			Type:   authz.PermRepos,
		}); err != nil {
			logger.Error(
				"failed to grant user pending permissions",
				sglog.Int32("userID", userID),
				sglog.Error(err),
			)
			// OK to continue, since this is a best-effort to improve the UX with some initial permissions available.
		}
	}

	return userSaved, userID, "", nil
}

func getUser(ctx context.Context, db database.DB, op GetAndSaveUserOp) (int32, bool, string, error) {
	if actor := sgactor.FromContext(ctx); actor.IsAuthenticated() {
		return actor.UID, false, "", nil
	}

	extAcct, err := db.UserExternalAccounts().LookupUserAndSave(ctx, op.ExternalAccount, op.ExternalAccountData)
	if err != nil {
		// If it's any error other than "not found", return it
		if !errcode.IsNotFound(err) {
			return 0, false, "Unexpected error looking up the Sourcegraph user account associated with the external account. Ask a site admin for help.", err
		}
		// Otherwise, check if we should look up by username
		if op.LookUpByUsername {
			user, err := db.Users().GetByUsername(ctx, op.UserProps.Username)
			if err != nil {
				return 0, false, "Unexpected error looking up the Sourcegraph user by username. Ask a site admin for help.", err
			}

			return user.ID, false, "", nil
		}

		// If not, try to get the user by verified email
		if op.UserProps.EmailIsVerified {
			user, err := db.Users().GetByVerifiedEmail(ctx, op.UserProps.Email)
			if err != nil {
				if !errcode.IsNotFound(err) {
					return 0, false, "Unexpected error looking up the Sourcegraph user by verified email. Ask a site admin for help.", err
				}
				return 0, false, fmt.Sprintf("User account with verified email %q does not exist. Ask a site admin to create your account and then verify your email.", op.UserProps.Email), err
			}

			return user.ID, false, "", nil
		}
	}

	return extAcct.UserID, true, "", nil
}

func createUser(ctx context.Context, db database.DB, op GetAndSaveUserOp) (int32, string, error) {
	logger := sglog.Scoped("authGetAndSaveUser", "get and save user authenticated by external providers")
	act := &sgactor.Actor{
		SourcegraphOperator: op.ExternalAccount.ServiceType == auth.SourcegraphOperatorProviderType,
	}

	// Fourth and finally, create a new user account and return it.
	//
	// If CreateIfNotExist is true, create the new user, regardless of whether the
	// email was verified or not.
	//
	// NOTE: It is important to propagate the correct context that carries the
	// information of the actor, especially whether the actor is a Sourcegraph
	// operator or not.
	ctx = sgactor.WithActor(ctx, act)
	user, err := db.UserExternalAccounts().CreateUserAndSave(ctx, op.UserProps, op.ExternalAccount, op.ExternalAccountData)
	switch {
	case database.IsUsernameExists(err):
		return 0, fmt.Sprintf("Username %q already exists, but no verified email matched %q", op.UserProps.Username, op.UserProps.Email), err
	case errcode.PresentationMessage(err) != "":
		return 0, errcode.PresentationMessage(err), err
	case err != nil:
		return 0, "Unable to create a new user account due to a unexpected error. Ask a site admin for help.", errors.Wrapf(err, "username: %q, email: %q", op.UserProps.Username, op.UserProps.Email)
	}
	act.UID = user.ID

	// Schedule a permission sync, since this is new user
	permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
		UserIDs:           []int32{user.ID},
		Reason:            database.ReasonUserAdded,
		TriggeredByUserID: user.ID,
	})

	if err = db.Authz().GrantPendingPermissions(ctx, &database.GrantPendingPermissionsArgs{
		UserID: user.ID,
		Perm:   authz.Read,
		Type:   authz.PermRepos,
	}); err != nil {
		logger.Error(
			"failed to grant user pending permissions",
			sglog.Int32("userID", user.ID),
			sglog.Error(err),
		)
		// OK to continue, since this is a best-effort to improve the UX with some initial permissions available.
	}

	const eventName = "ExternalAuthSignupSucceeded"
	args, err := json.Marshal(map[string]any{
		// NOTE: The conventional name should be "service_type", but keeping as-is for
		// backwards capability.
		"serviceType": op.ExternalAccount.ServiceType,
	})
	if err != nil {
		logger.Error(
			"failed to marshal JSON for event log argument",
			sglog.String("eventName", eventName),
			sglog.Error(err),
		)
		// OK to continue, we still want the event log to be created
	}

	// NOTE: It is important to propagate the correct context that carries the
	// information of the actor, especially whether the actor is a Sourcegraph
	// operator or not.
	err = usagestats.LogEvent(
		ctx,
		db,
		usagestats.Event{
			EventName: eventName,
			UserID:    act.UID,
			Argument:  args,
			Source:    "BACKEND",
		},
	)
	if err != nil {
		logger.Error(
			"failed to log event",
			sglog.String("eventName", eventName),
			sglog.Error(err),
		)
	}

	return user.ID, "", nil
}

func getAndSaveUser(ctx context.Context, db database.DB, op GetAndSaveUserOp) (int32, bool, bool, string, error) {
	newUser := false

	userID, extAcctSaved, safeErr, err := getUser(ctx, db, op)
	if err != nil {
		if !errcode.IsNotFound(err) {
			return 0, false, false, safeErr, err
		}

		if !op.CreateIfNotExist {
			return 0, false, false, safeErr, err
		}

		userID, safeErr, err = createUser(ctx, db, op)
		if err != nil {
			return 0, false, false, safeErr, err
		}
		extAcctSaved = true
		newUser = true
	}

	return userID, newUser, extAcctSaved, "", nil
}

func GetUser(ctx context.Context, db database.DB, op GetAndSaveUserOp) (int32, error) {
	if actor := sgactor.FromContext(ctx); actor.IsAuthenticated() {
		return actor.UID, nil
	}

	extAcct, err := db.UserExternalAccounts().LookupUserAndSave(ctx, op.ExternalAccount, op.ExternalAccountData)
	if err != nil {
		return 0, err
	}

	return extAcct.UserID, err
}
