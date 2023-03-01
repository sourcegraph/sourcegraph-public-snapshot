package scim

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	AttrUserName      = "userName"
	AttrDisplayName   = "displayName"
	AttrName          = "name"
	AttrNameFormatted = "formatted"
	AttrNameGiven     = "givenName"
	AttrNameMiddle    = "middleName"
	AttrNameFamily    = "familyName"
	AttrEmails        = "emails"
)

// UserResourceHandler implements the scim.ResourceHandler interface for users.
type UserResourceHandler struct {
	ctx              context.Context
	observationCtx   *observation.Context
	db               database.DB
	coreSchema       schema.Schema
	schemaExtensions []scim.SchemaExtension
}

// NewUserResourceHandler returns a new UserResourceHandler.
func NewUserResourceHandler(ctx context.Context, observationCtx *observation.Context, db database.DB) *UserResourceHandler {
	return &UserResourceHandler{
		ctx:              ctx,
		observationCtx:   observationCtx,
		db:               db,
		coreSchema:       createCoreSchema(),
		schemaExtensions: createSchemaExtensions(),
	}
}

// getUserFromDB returns the user with the given ID.
// When it fails, it returns an error that's safe to return to the client as a SCIM error.
func getUserFromDB(ctx context.Context, store database.UserStore, idStr string) (*types.UserForSCIM, error) {
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return nil, scimerrors.ScimError{Status: http.StatusBadRequest, Detail: "invalid user id"}
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

	return users[0], nil
}

// createUserResourceType creates a SCIM resource type for users.
func createUserResourceType(userResourceHandler *UserResourceHandler) scim.ResourceType {
	return scim.ResourceType{
		ID:               optional.NewString("User"),
		Name:             "User",
		Endpoint:         "/Users",
		Description:      optional.NewString("User Account"),
		Schema:           userResourceHandler.coreSchema,
		SchemaExtensions: userResourceHandler.schemaExtensions,
		Handler:          userResourceHandler,
	}
}

// updateUser updates a user in the database. This is meant to be used in a transaction.
func updateUser(ctx context.Context, db database.DB, oldUser *types.UserForSCIM, newUser scim.Resource) (err error) {
	usernameUpdate := ""
	requestedUsername := extractStringAttribute(newUser.Attributes, AttrUserName)
	if requestedUsername != oldUser.Username {
		usernameUpdate, err = getUniqueUsername(ctx, db.Users(), requestedUsername)
		if err != nil {
			return scimerrors.ScimError{Status: http.StatusBadRequest, Detail: errors.Wrap(err, "invalid username").Error()}
		}
	}
	var displayNameUpdate *string
	var avatarURLUpdate *string
	userUpdate := database.UserUpdate{
		Username:    usernameUpdate,
		DisplayName: displayNameUpdate,
		AvatarURL:   avatarURLUpdate,
	}
	err = db.Users().Update(ctx, oldUser.ID, userUpdate)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}

	accountData, err := toAccountData(newUser.Attributes)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	// TODO: External ID currently can't be updated, can it be updated through SCIM?
	_, err = db.UserExternalAccounts().LookupUserAndSave(ctx, extsvc.AccountSpec{
		ServiceType: "scim",
		ServiceID:   "TODO", // TODO: Start using service IDs
		ClientID:    "",
		AccountID:   oldUser.SCIMExternalID,
	}, accountData)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}

	return
}

// getOptionalExternalID extracts the external identifier of the given attributes.
func getOptionalExternalID(attributes scim.ResourceAttributes) optional.String {
	if eID, ok := attributes["externalId"]; ok {
		if externalID, ok := eID.(string); ok {
			return optional.NewString(externalID)
		}
	}
	return optional.String{}
}

// extractStringAttribute extracts the username from the given attributes.
func extractStringAttribute(attributes scim.ResourceAttributes, name string) (username string) {
	if attributes[name] != nil {
		username = attributes[name].(string)
	}
	return
}

// getUniqueUsername returns a unique username based on the given requested username plus normalization,
// and adding a random suffix to make it unique in case there one without a suffix already exists in the DB.
// This is meant to be done inside a transaction so that the user creation/update is guaranteed to be
// coherent with the evaluation of this function.
func getUniqueUsername(ctx context.Context, tx database.UserStore, requestedUsername string) (string, error) {
	// Process requested username
	normalizedUsername, err := auth.NormalizeUsername(requestedUsername)
	if err != nil {
		// Empty username after normalization. Generate a random one, it's the best we can do.
		normalizedUsername, err = auth.AddRandomSuffix("")
		if err != nil {
			return "", scimerrors.ScimErrorBadParams([]string{"invalid username"})
		}
	}
	_, err = tx.GetByUsername(ctx, normalizedUsername)
	if err == nil { // Username exists, try to add random suffix
		normalizedUsername, err = auth.AddRandomSuffix(normalizedUsername)
		if err != nil {
			return "", scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not normalize username").Error()}
		}
	} else if !database.IsUserNotFoundErr(err) {
		return "", scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not check if username exists").Error()}
	}
	return normalizedUsername, nil
}

// checkBodyNotEmpty checks whether the request body is empty. If it is, it returns a SCIM error.
func checkBodyNotEmpty(r *http.Request) error {
	// Check whether the request body is empty.
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return scimerrors.ScimErrorBadParams([]string{"request body is empty"})
	}
	return nil
}
