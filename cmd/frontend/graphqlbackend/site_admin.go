package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RecoverUsersRequest struct {
	UserIDs []graphql.ID
}

func (r *schemaResolver) RecoverUsers(ctx context.Context, args *RecoverUsersRequest) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can recover users.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if len(args.UserIDs) == 0 {
		return nil, errors.New("must specify at least one user ID")
	}

	// a must be authenticated at this point, CheckCurrentUserIsSiteAdmin enforces it.
	a := actor.FromContext(ctx)

	ids := make([]int32, len(args.UserIDs))
	for index, user := range args.UserIDs {
		id, err := UnmarshalUserID(user)
		if err != nil {
			return nil, err
		}
		if a.UID == id {
			return nil, errors.New("unable to recover current user")
		}
		ids[index] = id
	}

	users, err := r.db.Users().RecoverUsersList(ctx, ids)
	if err != nil {
		return nil, err
	}

	if len(users) != len(ids) {
		missingUserIds := missingUserIds(ids, users)
		return nil, errors.Errorf("some users were not found, expected to recover %d users, but found only %d users. Missing user IDs: %s", len(ids), len(users), missingUserIds)
	}

	return &EmptyResponse{}, nil
}

func (r *schemaResolver) DeleteUser(ctx context.Context, args *struct {
	User graphql.ID
	Hard *bool
}) (*EmptyResponse, error) {
	return r.DeleteUsers(ctx, &struct {
		Users []graphql.ID
		Hard  *bool
	}{
		Users: []graphql.ID{args.User},
		Hard:  args.Hard,
	})
}

func (r *schemaResolver) DeleteUsers(ctx context.Context, args *struct {
	Users []graphql.ID
	Hard  *bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete users.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if len(args.Users) == 0 {
		return nil, errors.New("must specify at least one user ID")
	}

	// a must be authenticated at this point, CheckCurrentUserIsSiteAdmin enforces it.
	a := actor.FromContext(ctx)

	ids := make([]int32, len(args.Users))
	for index, user := range args.Users {
		id, err := UnmarshalUserID(user)
		if err != nil {
			return nil, err
		}
		if a.UID == id {
			return nil, errors.New("unable to delete current user")
		}
		ids[index] = id
	}

	logger := r.logger.Scoped("DeleteUsers").
		With(log.Int32s("users", ids))

	// Collect username, verified email addresses, and external accounts to be used
	// for revoking user permissions later, otherwise they will be removed from database
	// if it's a hard delete.
	users, err := r.db.Users().List(ctx, &database.UsersListOptions{
		UserIDs: ids,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list users by IDs")
	}
	if len(users) == 0 {
		logger.Info("requested users to delete do not exist")
	} else {
		logger.Debug("attempting to delete requested users")
	}

	accountsList := make([][]*extsvc.Accounts, len(users))
	var revokeUserPermissionsArgsList []*database.RevokeUserPermissionsArgs
	for index, user := range users {
		var accounts []*extsvc.Accounts

		extAccounts, err := r.db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{UserID: user.ID})
		if err != nil {
			return nil, errors.Wrap(err, "list external accounts")
		}
		for _, acct := range extAccounts {
			// If the delete target is a SOAP user, make sure the actor is also a SOAP
			// user - regular users should not be able to delete SOAP users.
			if acct.ServiceType == auth.SourcegraphOperatorProviderType {
				if !a.SourcegraphOperator {
					return nil, errors.Newf("%[1]q user %[2]d cannot be deleted by a non-%[1]q user",
						auth.SourcegraphOperatorProviderType, user.ID)
				}
			}

			accounts = append(accounts, &extsvc.Accounts{
				ServiceType: acct.ServiceType,
				ServiceID:   acct.ServiceID,
				AccountIDs:  []string{acct.AccountID},
			})
		}

		verifiedEmails, err := r.db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		})
		if err != nil {
			return nil, err
		}
		emailStrs := make([]string, len(verifiedEmails))
		for i := range verifiedEmails {
			emailStrs[i] = verifiedEmails[i].Email
		}
		accounts = append(accounts, &extsvc.Accounts{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			AccountIDs:  append(emailStrs, user.Username),
		})

		accountsList[index] = accounts

		revokeUserPermissionsArgsList = append(revokeUserPermissionsArgsList, &database.RevokeUserPermissionsArgs{
			UserID:   user.ID,
			Accounts: accounts,
		})
	}

	if args.Hard != nil && *args.Hard {
		if err := r.db.Users().HardDeleteList(ctx, ids); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Users().DeleteList(ctx, ids); err != nil {
			return nil, err
		}
	}

	// NOTE: Practically, we don't reuse the ID for any new users, and the situation of left-over pending permissions
	// is possible but highly unlikely. Therefore, there is no need to roll back user deletion even if this step failed.
	// This call is purely for the purpose of cleanup.
	// TODO: Add user deletion and this to a transaction. See SCIM's user_delete.go for an example.
	if err := r.db.Authz().RevokeUserPermissionsList(ctx, revokeUserPermissionsArgsList); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

func (r *schemaResolver) DeleteOrganization(ctx context.Context, args *struct {
	Organization graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: For On-premise, only site admins can soft delete orgs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	orgID, err := UnmarshalOrgID(args.Organization)
	if err != nil {
		return nil, err
	}

	if err := r.db.Orgs().Delete(ctx, orgID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type roleChangeEventArgs struct {
	By   int32  `json:"by"`
	For  int32  `json:"for"`
	From string `json:"from"`
	To   string `json:"to"`

	// Reason will be present only if the RoleChangeDenied event is logged, but will be set to an
	// empty string in other cases for a consistent experience of the clients that consume this
	// data.
	Reason string `json:"reason"`
}

var errRefuseToSetCurrentUserSiteAdmin = errors.New("refusing to set current user site admin status")

func (r *schemaResolver) SetUserIsSiteAdmin(ctx context.Context, args *struct {
	UserID    graphql.ID
	SiteAdmin bool
}) (response *EmptyResponse, err error) {
	// Set default values for event args.
	eventArgs := roleChangeEventArgs{
		From: "role_user",
		To:   "role_site_admin",
	}

	// Correct the values based on the value of SiteAdmin in the GraphQL mutation.
	if !args.SiteAdmin {
		eventArgs.From = "role_site_admin"
		eventArgs.To = "role_user"
	}

	affectedUserID, err := UnmarshalUserID(args.UserID)
	if err != nil {
		return nil, err
	}

	eventArgs.For = affectedUserID

	userResolver, err := CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}

	eventArgs.By = userResolver.user.ID

	// At the moment, we log only two types of events:
	// - RoleChangeDenied
	// - RoleChangeGranted
	//
	// Unless we want to log another event for RoleChangeAttempted as well, invoking
	// logRoleChangeAttempt before this point does not make sense since this is the first time in
	// the lifetime of this function when we have all the details required for eventArgs, especially
	// eventArgs.By which is used as the UserID in database.SecurityEvent - a required argument to
	// write an entry into the database.
	eventName := database.SecurityEventNameRoleChangeDenied
	defer logRoleChangeAttempt(ctx, r.db, &eventName, &eventArgs, &err)

	// 🚨 SECURITY: Only site admins can promote other users to site admin (or demote from site
	// admin).
	if err = auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if userResolver.ID() == args.UserID {
		return nil, errRefuseToSetCurrentUserSiteAdmin
	}

	if err = r.db.Users().SetIsSiteAdmin(ctx, affectedUserID, args.SiteAdmin); err != nil {
		return nil, err
	}

	eventName = database.SecurityEventNameRoleChangeGranted
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) InvalidateSessionsByID(ctx context.Context, args *struct {
	UserID graphql.ID
}) (*EmptyResponse, error) {
	return r.InvalidateSessionsByIDs(ctx, &struct{ UserIDs []graphql.ID }{UserIDs: []graphql.ID{args.UserID}})
}

func (r *schemaResolver) InvalidateSessionsByIDs(ctx context.Context, args *struct {
	UserIDs []graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only the site admin can invalidate the sessions of a user
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	if len(args.UserIDs) == 0 {
		return nil, errors.New("must specify at least one user ID")
	}
	userIDs := make([]int32, len(args.UserIDs))
	for index, id := range args.UserIDs {
		userID, err := UnmarshalUserID(id)
		if err != nil {
			return nil, err
		}
		userIDs[index] = userID
	}
	if err := session.InvalidateSessionsByIDs(ctx, r.db, userIDs); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func logRoleChangeAttempt(ctx context.Context, db database.DB, name *database.SecurityEventName, eventArgs *roleChangeEventArgs, parentErr *error) {
	// To avoid a panic, it's important to check for a nil parentErr before we dereference it.
	if parentErr != nil && *parentErr != nil {
		eventArgs.Reason = (*parentErr).Error()
	}

	db.SecurityEventLogs().LogSecurityEvent(ctx, *name, "", uint32(eventArgs.By), "", "BACKEND", eventArgs)
}

func missingUserIds(id, affectedIds []int32) []graphql.ID {
	maffectedIds := make(map[int32]struct{}, len(affectedIds))
	for _, x := range affectedIds {
		maffectedIds[x] = struct{}{}
	}
	var diff []graphql.ID
	for _, x := range id {
		if _, found := maffectedIds[x]; !found {
			strId := MarshalUserID(x)
			diff = append(diff, strId)
		}
	}
	return diff
}
