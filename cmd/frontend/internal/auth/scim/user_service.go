package scim

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Reusing the env variables from email invites because the intent is the same as the welcome email
var (
	disableEmailInvites, _   = strconv.ParseBool(env.Get("DISABLE_EMAIL_INVITES", "false", "Disable email invitations entirely."))
	debugEmailInvitesMock, _ = strconv.ParseBool(env.Get("DEBUG_EMAIL_INVITES_MOCK", "false", "Do not actually send email invitations, instead just print that we did."))
)

const (
	AttrUserName      = "userName"
	AttrDisplayName   = "displayName"
	AttrName          = "name"
	AttrNameFormatted = "formatted"
	AttrNameGiven     = "givenName"
	AttrNameMiddle    = "middleName"
	AttrNameFamily    = "familyName"
	AttrNickName      = "nickName"
	AttrEmails        = "emails"
	AttrExternalId    = "externalId"
	AttrActive        = "active"
)

// NewUserResourceHandler returns a new UserResourceHandler.
func NewUserResourceHandler(ctx context.Context, observationCtx *observation.Context, db database.DB) *ResourceHandler {
	userSCIMService := &UserSCIMService{
		db: db,
	}
	return &ResourceHandler{
		ctx:              ctx,
		observationCtx:   observationCtx,
		coreSchema:       userSCIMService.Schema(),
		schemaExtensions: userSCIMService.SchemaExtensions(),
		service:          userSCIMService,
	}
}

// createUserResourceType creates a SCIM resource type for users.
func createResourceType(name, endpoint, description string, resourceHandler *ResourceHandler) scim.ResourceType {
	return scim.ResourceType{
		ID:               optional.NewString(name),
		Name:             name,
		Endpoint:         endpoint,
		Description:      optional.NewString(description),
		Schema:           resourceHandler.service.Schema(),
		SchemaExtensions: resourceHandler.service.SchemaExtensions(),
		Handler:          resourceHandler,
	}
}

type UserSCIMService struct {
	db database.DB
}

func (u *UserSCIMService) getLogger() log.Logger {
	return log.Scoped("scim.user")
}

func (u *UserSCIMService) Get(ctx context.Context, id string) (scim.Resource, error) {
	user, err := getUserFromDB(ctx, u.db.Users(), id)
	if err != nil {
		return scim.Resource{}, err
	}
	return user.ToResource(), nil
}

func (u *UserSCIMService) GetAll(ctx context.Context, start int, count *int) (totalCount int, entities []scim.Resource, err error) {
	return getAllUsersFromDB(ctx, u.db.Users(), start, count)
}

func (u *UserSCIMService) Update(ctx context.Context, id string, applySCIMUpdates func(getResource func() scim.Resource) (updated scim.Resource, _ error)) (finalResource scim.Resource, _ error) {
	var resourceAfterUpdate scim.Resource
	err := u.db.WithTransact(ctx, func(tx database.DB) error {
		var txErr error
		user, txErr := getUserFromDB(ctx, tx.Users(), id)
		if txErr != nil {
			return txErr
		}

		// Capture a copy of the resource before applying updates so it can be compared to determine which
		// database updates are necessary
		resourceBeforeUpdate := user.ToResource()
		resourceAfterUpdate, txErr = applySCIMUpdates(user.ToResource)
		if txErr != nil {
			return txErr
		}

		updateUser := NewUserUpdate(tx, user)
		txErr = updateUser.Update(ctx, &resourceBeforeUpdate, &resourceAfterUpdate)
		return txErr
	})
	if err != nil {
		multiErr, ok := err.(errors.MultiError)
		if !ok || len(multiErr.Errors()) == 0 {
			return scim.Resource{}, err
		}
		return scim.Resource{}, multiErr.Errors()[len(multiErr.Errors())-1]
	}
	return resourceAfterUpdate, nil
}

func (u *UserSCIMService) Create(ctx context.Context, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// Extract external ID, primary email, username, and display name from attributes to variables
	primaryEmail, otherEmails := extractPrimaryEmail(attributes)
	if primaryEmail == "" {
		return scim.Resource{}, scimerrors.ScimErrorBadParams([]string{"emails missing"})
	}
	displayName := extractDisplayName(attributes)

	// Try to match emails to existing users
	allEmails := append([]string{primaryEmail}, otherEmails...)
	existingEmails, err := u.db.UserEmails().GetVerifiedEmails(ctx, allEmails...)
	if err != nil {
		return scim.Resource{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	existingUserIDs := make(map[int32]struct{})
	for _, email := range existingEmails {
		existingUserIDs[email.UserID] = struct{}{}
	}
	if len(existingUserIDs) > 1 {
		return scim.Resource{}, scimerrors.ScimError{Status: http.StatusConflict, Detail: "Emails match to multiple users"}
	}
	if len(existingUserIDs) == 1 {
		userID := int32(0)
		for id := range existingUserIDs {
			userID = id
		}
		// A user with the email(s) already exists → check if the user is not SCIM-controlled
		user, err := u.db.Users().GetByID(ctx, userID)
		if err != nil {
			return scim.Resource{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
		}
		if user == nil {
			return scim.Resource{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: "User not found"}
		}
		if user.SCIMControlled {
			// This user creation would fail based on the email address, so we'll return a conflict error
			return scim.Resource{}, scimerrors.ScimError{Status: http.StatusConflict, Detail: "User already exists based on email address"}
		}

		// The user exists, but is not SCIM-controlled, so we'll update the user with the new attributes,
		// and make the user SCIM-controlled (which is the same as a replace)
		return u.Update(ctx, strconv.Itoa(int(userID)), func(getResource func() scim.Resource) (updated scim.Resource, _ error) {
			now := time.Now()
			return scim.Resource{
				ID:         strconv.Itoa(int(userID)),
				ExternalID: getOptionalExternalID(attributes),
				Attributes: attributes,
				Meta: scim.Meta{
					Created:      &now,
					LastModified: &now,
				},
			}, nil
		})
	}

	// At this point we know that the user does not exist yet, so we'll create a new user

	// Make sure the username is unique, then create user with/without an external account ID
	var user *types.User
	err = u.db.WithTransact(ctx, func(tx database.DB) error {
		uniqueUsername, err := getUniqueUsername(ctx, tx.Users(), 0, extractStringAttribute(attributes, AttrUserName))
		if err != nil {
			return err
		}

		// Create user
		newUser := database.NewUser{
			Email:           primaryEmail,
			Username:        uniqueUsername,
			DisplayName:     displayName,
			EmailIsVerified: true,
		}
		accountSpec := extsvc.AccountSpec{
			ServiceType: "scim",
			ServiceID:   "scim",
			AccountID:   getUniqueExternalID(attributes),
		}
		accountData, err := toAccountData(attributes)
		if err != nil {
			return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
		}
		user, err = tx.Users().CreateWithExternalAccount(ctx, newUser,
			&extsvc.Account{
				AccountSpec: accountSpec,
				AccountData: accountData,
			})

		if err != nil {
			if dbErr, ok := containsErrCannotCreateUserError(err); ok {
				code := dbErr.Code()
				if code == database.ErrorCodeUsernameExists || code == database.ErrorCodeEmailExists {
					return scimerrors.ScimError{Status: http.StatusConflict, Detail: err.Error()}
				}
			}
			return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
		}
		return nil
	})
	if err != nil {
		multiErr, ok := err.(errors.MultiError)
		if !ok || len(multiErr.Errors()) == 0 {
			return scim.Resource{}, err
		}
		return scim.Resource{}, multiErr.Errors()[len(multiErr.Errors())-1]
	}

	// If there were additional emails provided, now that the user has been created
	// we can try to add and verify them each in a separate trx so that if it fails we can ignore
	// the error because they are not required.
	if len(otherEmails) > 0 {
		for _, email := range otherEmails {
			_ = u.db.WithTransact(ctx, func(tx database.DB) error {
				err := tx.UserEmails().Add(ctx, user.ID, email, nil)
				if err != nil {
					return err
				}
				return tx.UserEmails().SetVerified(ctx, user.ID, email, true)
			})
		}
	}

	// Attempt to send emails in the background.
	goroutine.Go(func() {
		_ = sendPasswordResetEmail(u.getLogger(), u.db, user, primaryEmail)
		_ = sendWelcomeEmail(primaryEmail, conf.ExternalURLParsed().String(), u.getLogger())
	})

	now := time.Now()
	return scim.Resource{
		ID:         strconv.Itoa(int(user.ID)),
		ExternalID: getOptionalExternalID(attributes),
		Attributes: attributes,
		Meta: scim.Meta{
			Created:      &now,
			LastModified: &now,
		},
	}, nil
}

func (u *UserSCIMService) Delete(ctx context.Context, id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.Wrap(err, "parse user ID")
	}
	user, err := findUser(ctx, u.db, idInt)
	if err != nil {
		return err
	}

	// If we found no user, we report “all clear” to match the spec
	if user.Username == "" {
		return nil
	}

	// Delete user and revoke user permissions
	err = u.db.WithTransact(ctx, func(tx database.DB) error {
		// Save username, verified emails, and external accounts to be used for revoking user permissions after deletion
		revokeUserPermissionsArgsList, err := getRevokeUserPermissionArgs(ctx, user, u.db)
		if err != nil {
			return err
		}

		if err := tx.Users().HardDelete(ctx, int32(idInt)); err != nil {
			return err
		}

		// NOTE: Practically, we don't reuse the ID for any new users, and the situation of left-over pending permissions
		// is possible but highly unlikely. Therefore, there is no need to roll back user deletion even if this step failed.
		// This call is purely for the purpose of cleanup.
		return tx.Authz().RevokeUserPermissionsList(ctx, []*database.RevokeUserPermissionsArgs{revokeUserPermissionsArgsList})
	})
	if err != nil {
		return errors.Wrap(err, "delete user")
	}

	return nil
}

// Helper functions used for Users

// getUserFromDB returns the user with the given ID.
// When it fails, it returns an error that's safe to return to the client as a SCIM error.
func getUserFromDB(ctx context.Context, store database.UserStore, idStr string) (*User, error) {
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return nil, scimerrors.ScimErrorResourceNotFound(idStr)
	}

	users, err := store.ListForSCIM(ctx, &database.UsersListOptions{
		UserIDs: []int32{int32(id)},
	})
	if err != nil {
		return nil, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	if len(users) == 0 {
		return nil, scimerrors.ScimErrorResourceNotFound(idStr)
	}

	return &User{UserForSCIM: *users[0]}, nil
}

func getAllUsersFromDB(ctx context.Context, store database.UserStore, startIndex int, count *int) (totalCount int, resources []scim.Resource, err error) {
	// Calculate offset
	var offset int
	if startIndex > 0 {
		offset = startIndex - 1
	}

	// Get users and convert them to SCIM resources
	opt := &database.UsersListOptions{}
	if count != nil {
		opt = &database.UsersListOptions{
			LimitOffset: &database.LimitOffset{Limit: *count, Offset: offset},
		}
	}
	users, err := store.ListForSCIM(ctx, opt)
	if err != nil {
		return
	}
	resources = make([]scim.Resource, 0, len(users))
	for _, user := range users {
		u := User{UserForSCIM: *user}
		resources = append(resources, u.ToResource())
	}

	// Get total count
	if count == nil {
		totalCount = len(users)
	} else {
		totalCount, err = store.CountForSCIM(ctx, &database.UsersListOptions{})
	}

	return
}

// findUser finds the user with the given ID. If the user does not exist, it returns an empty user.
func findUser(ctx context.Context, db database.DB, id int) (types.UserForSCIM, error) {
	users, err := db.Users().ListForSCIM(ctx, &database.UsersListOptions{
		UserIDs: []int32{int32(id)},
	})
	if err != nil {
		return types.UserForSCIM{}, errors.Wrap(err, "list users by IDs")
	}
	if len(users) == 0 {
		return types.UserForSCIM{}, nil
	}
	if users[0].SCIMAccountData == "" {
		return types.UserForSCIM{}, errors.New("cannot delete user because it doesn't seem to be SCIM-controlled")
	}
	user := *users[0]
	return user, nil
}

// Permissions

// getRevokeUserPermissionArgs returns a list of arguments for revoking user permissions.
func getRevokeUserPermissionArgs(ctx context.Context, user types.UserForSCIM, db database.DB) (*database.RevokeUserPermissionsArgs, error) {
	// Collect external accounts
	var accounts []*extsvc.Accounts
	extAccounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{UserID: user.ID})
	if err != nil {
		return nil, errors.Wrap(err, "list external accounts")
	}
	for _, acct := range extAccounts {
		accounts = append(accounts, &extsvc.Accounts{
			ServiceType: acct.ServiceType,
			ServiceID:   acct.ServiceID,
			AccountIDs:  []string{acct.AccountID},
		})
	}

	// Add Sourcegraph account
	accounts = append(accounts, &extsvc.Accounts{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountIDs:  append(user.Emails, user.Username),
	})

	return &database.RevokeUserPermissionsArgs{
		UserID:   user.ID,
		Accounts: accounts,
	}, nil
}

// Emails

// sendPasswordResetEmail sends a password reset email to the given user.
func sendPasswordResetEmail(logger log.Logger, db database.DB, user *types.User, primaryEmail string) bool {
	// Email user to ask to set up a password
	// This internally checks whether username/password login is enabled, whether we have an SMTP in place, etc.
	if disableEmailInvites {
		return true
	}
	if debugEmailInvitesMock {
		if logger != nil {
			logger.Info("password reset: mock pw reset email to Sourcegraph", log.String("sent", primaryEmail))
		}
		return true
	}
	_, err := auth.ResetPasswordURL(context.Background(), db, logger, user, primaryEmail, true)
	if err != nil {
		logger.Error("error sending password reset email", log.Error(err))
	}
	return false
}

// sendWelcomeEmail sends a welcome email to the given user.
func sendWelcomeEmail(email, siteURL string, logger log.Logger) error {
	if email != "" && conf.CanSendEmail() {
		if disableEmailInvites {
			return nil
		}
		if debugEmailInvitesMock {
			if logger != nil {
				logger.Info("email welcome: mock welcome to Sourcegraph", log.String("welcomed", email))
			}
			return nil
		}
		return txemail.Send(context.Background(), "user_welcome", txemail.Message{
			To:       []string{email},
			Template: emailTemplateEmailWelcomeSCIM,
			Data: struct {
				URL string
			}{
				URL: siteURL,
			},
		})
	}
	return nil
}

var emailTemplateEmailWelcomeSCIM = txemail.MustValidate(txtypes.Templates{
	Subject: `Welcome to Sourcegraph`,
	Text: `
Sourcegraph enables you to quickly understand, fix, and automate changes to your code.

You can use Sourcegraph to:
  - Search and navigate multiple repositories with cross-repository dependency navigation
  - Share links directly to lines of code to work more collaboratively together
  - Automate large-scale code changes with Batch Changes
  - Create code monitors to alert you about changes in code

Come experience the power of great code search.


{{.URL}}

Learn more about Sourcegraph:

https://sourcegraph.com
`,
	HTML: `
<p>Sourcegraph enables you to quickly understand, fix, and automate changes to your code.</p>

<p>
	You can use Sourcegraph to:<br/>
	<ul>
		<li>Search and navigate multiple repositories with cross-repository dependency navigation</li>
		<li>Share links directly to lines of code to work more collaboratively together</li>
		<li>Automate large-scale code changes with Batch Changes</li>
		<li>Create code monitors to alert you about changes in code</li>
	</ul>
</p>

<p><strong><a href="{{.URL}}">Come experience the power of great code search</a></strong></p>

<p><a href="https://sourcegraph.com">Learn more about Sourcegraph</a></p>
`,
})
