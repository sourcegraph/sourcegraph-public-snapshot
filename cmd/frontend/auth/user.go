package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sourcegraph/sourcegraph/internal/conf"

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
	users := db.Users()
	logger := sglog.Scoped("authGetAndSaveUser")

	acct := &extsvc.Account{
		AccountSpec: op.ExternalAccount,
		AccountData: op.ExternalAccountData,
	}

	userID, newUserSaved, extAcctSaved, safeErrMsg, err := func() (int32, bool, bool, string, error) {
		// First, check if the user is already logged in. If so, return that user.
		if actor := sgactor.FromContext(ctx); actor.IsAuthenticated() {
			return actor.UID, false, false, "", nil
		}

		// Second, check if the user account already exists. If so, return that user.
		extsvcAcct, lookupByExternalErr := externalAccountsStore.Update(ctx, acct)
		if lookupByExternalErr == nil {
			return extsvcAcct.UserID, false, true, "", nil
		}
		if !errcode.IsNotFound(lookupByExternalErr) {
			return 0, false, false, "Unexpected error looking up the Sourcegraph user account associated with the external account. Ask a site admin for help.", lookupByExternalErr
		}

		if op.LookUpByUsername {
			user, getByUsernameErr := db.Users().GetByUsername(ctx, op.UserProps.Username)
			if getByUsernameErr == nil {
				return user.ID, false, false, "", nil
			}
			if !errcode.IsNotFound(getByUsernameErr) {
				return 0, false, false, "Unexpected error looking up the Sourcegraph user by username. Ask a site admin for help.", getByUsernameErr
			}
			if !op.CreateIfNotExist {
				return 0, false, false, fmt.Sprintf("User account with username %q does not exist. Ask a site admin to create your account.", op.UserProps.Username), getByUsernameErr
			}
		} else if op.UserProps.EmailIsVerified {
			user, getByVerifiedEmailErr := db.Users().GetByVerifiedEmail(ctx, op.UserProps.Email)
			if getByVerifiedEmailErr == nil {
				return user.ID, false, false, "", nil
			}
			if !errcode.IsNotFound(getByVerifiedEmailErr) {
				return 0, false, false, "Unexpected error looking up the Sourcegraph user by verified email. Ask a site admin for help.", getByVerifiedEmailErr
			}
			if !op.CreateIfNotExist {
				return 0, false, false, fmt.Sprintf("User account with verified email %q does not exist. Ask a site admin to create your account and then verify your email.", op.UserProps.Email), getByVerifiedEmailErr
			}
		}

		// Third, return an error here if creating new users is disabled.
		if !op.CreateIfNotExist {
			return 0, false, false, "It looks like this is your first time signing in with this external identity. Sourcegraph couldn't link it to an existing user, because no verified email was provided. Ask your site admin to configure the auth provider to include the user's verified email on sign-in.", lookupByExternalErr
		}

		act := &sgactor.Actor{
			SourcegraphOperator: acct.AccountSpec.ServiceType == auth.SourcegraphOperatorProviderType,
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
		user, err := users.CreateWithExternalAccount(ctx, op.UserProps, acct)

		switch {
		case database.IsUsernameExists(err):
			return 0, false, false, fmt.Sprintf("Username %q already exists, but no verified email matched %q", op.UserProps.Username, op.UserProps.Email), err
		case errcode.PresentationMessage(err) != "":
			return 0, false, false, errcode.PresentationMessage(err), err
		case err != nil:
			return 0, false, false, "Unable to create a new user account due to a unexpected error. Ask a site admin for help.", errors.Wrapf(err, "username: %q, email: %q", op.UserProps.Username, op.UserProps.Email)
		}
		act.UID = user.ID

		// Schedule a permission sync, since this is new user
		permssync.SchedulePermsSync(ctx, logger, db, permssync.ScheduleSyncOpts{
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

		// SECURITY: This args map is treated as a public argument in the LogEvent call below, so it must not contain
		// any sensitive data.
		args, err := json.Marshal(map[string]any{
			// NOTE: The conventional name should be "service_type", but keeping as-is for
			// backwards capability.
			"serviceType": acct.AccountSpec.ServiceType,
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
		//
		// TODO: Use EventRecorder from internal/telemetryrecorder instead.
		//lint:ignore SA1019 existing usage of deprecated functionality.
		err = usagestats.LogEvent(
			ctx,
			db,
			usagestats.Event{
				EventName:      eventName,
				UserID:         act.UID,
				Argument:       args,
				PublicArgument: args,
				Source:         "BACKEND",
			},
		)
		if err != nil {
			logger.Error(
				"failed to log event",
				sglog.String("eventName", eventName),
				sglog.Error(err),
			)
		}

		return user.ID, true, true, "", nil
	}()
	if err != nil {
		const eventName = "ExternalAuthSignupFailed"
		serviceTypeArg := json.RawMessage(fmt.Sprintf(`{"serviceType": %q}`, acct.AccountSpec.ServiceType))
		// TODO: Use EventRecorder from internal/telemetryrecorder instead.
		//lint:ignore SA1019 existing usage of deprecated functionality.
		if logErr := usagestats.LogBackendEvent(db, sgactor.FromContext(ctx).UID, deviceid.FromContext(ctx), eventName, serviceTypeArg, serviceTypeArg, featureflag.GetEvaluatedFlagSet(ctx), nil); logErr != nil {
			logger.Error(
				"failed to log event",
				sglog.String("eventName", eventName),
				sglog.Error(err),
			)
		}
		return newUserSaved, 0, safeErrMsg, err
	}

	// Update user properties, if they've changed
	if !newUserSaved {
		// Update user in our DB if their profile info changed on the issuer. (Except username and
		// email, which the user is somewhat likely to want to control separately on Sourcegraph.)
		user, err := db.Users().GetByID(ctx, userID)
		if err != nil {
			return newUserSaved, 0, "Unexpected error getting the Sourcegraph user account. Ask a site admin for help.", err
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
				return newUserSaved, 0, "Unexpected error updating the Sourcegraph user account with new user profile information from the external account. Ask a site admin for help.", err
			}
		}
	}

	// Create/update the external account and ensure it's associated with the user ID
	if !extAcctSaved {
		acct.UserID = userID
		_, err := externalAccountsStore.Upsert(ctx, acct)
		if err != nil {
			return newUserSaved, 0, "Unexpected error associating the external account with your Sourcegraph user. The most likely cause for this problem is that another Sourcegraph user is already linked with this external account. A site admin or the other user can unlink the account to fix this problem.", err
		}

		// Schedule a permission sync, since this is probably a new external account for the user
		permssync.SchedulePermsSync(ctx, logger, db, permssync.ScheduleSyncOpts{
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

	// Enable the cody-pro feature flag for new users who are on the "exempted from the minimum external account age" list
	dc := conf.Get().Dotcom
	verifiedEmails, err := db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{UserID: userID, OnlyVerified: true})
	fmt.Println("XXX verifiedEmails", verifiedEmails, dc, err)
	if dc != nil && err == nil {
		exempted := false
		for _, exemptedEmail := range dc.MinimumExternalAccountAgeExemptList {
			for _, verifiedEmail := range verifiedEmails {
				if verifiedEmail.Email == exemptedEmail {
					exempted = true
					break
				}
			}
			if exempted {
				break
			}
		}
		fmt.Println("XXX exempted", exempted)
		if exempted {
			_, err = db.FeatureFlags().CreateOverride(context.Background(), &featureflag.Override{FlagName: "cody-pro", Value: true, UserID: &userID})
			if err != nil {
				logger.Error("failed to create feature flag override", sglog.Error(err))
				// Don't fail, though.
			}
		}
	}

	return newUserSaved, userID, "", nil
}
