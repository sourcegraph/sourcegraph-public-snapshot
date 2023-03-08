package scim

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	AttrNickName      = "nickName"
	AttrEmails        = "emails"
	AttrExternalId    = "externalId"
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

// getDBFunc can be used to toggle between two database.DB
type getDBFunc func(includeInTx bool) database.DB

// makeGetDB returns a function to toggle between 2 database.DB based on a boolean
func makeGetDB(tx, notTx database.DB) getDBFunc {
	return func(includeInTx bool) database.DB {
		if includeInTx {
			return tx
		}
		return notTx
	}
}

// updateUser updates a user in the database. This is meant to be used in a transaction.
func updateUser(ctx context.Context, getTx getDBFunc, oldUser *types.UserForSCIM, attributes scim.ResourceAttributes, emailsModified bool) (err error) {
	usernameUpdate := ""
	// Get a copy of how the user started before the update so we can diff them if needed
	startingUserSCIMResource := convertUserToSCIMResource(oldUser)

	requestedUsername := extractStringAttribute(attributes, AttrUserName)
	if requestedUsername != oldUser.Username {
		usernameUpdate, err = getUniqueUsername(ctx, getTx(true).Users(), requestedUsername)
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
	err = getTx(true).Users().Update(ctx, oldUser.ID, userUpdate)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}

	accountData, err := toAccountData(attributes)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	err = getTx(true).UserExternalAccounts().UpdateSCIMData(ctx, oldUser.ID, getUniqueExternalID(attributes), accountData)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}

	if emailsModified {
		currentEmails, err := getTx(true).UserEmails().ListByUser(ctx, database.UserEmailsListOptions{UserID: oldUser.ID, OnlyVerified: false})
		if err != nil {
			return err
		}
		updates := generateEmailUpdates(startingUserSCIMResource.Attributes, attributes, currentEmails)
		// First add any new email address and only error if this email is required to succeed
		for _, newEmail := range updates.toAdd {
			// We only need to ensure that the email update is successful if it is the primary email
			isPrimaryEmail := updates.resetPrimaryTo != nil && strings.EqualFold(*updates.resetPrimaryTo, newEmail)
			db := getTx(isPrimaryEmail)
			addErr := db.WithTransact(ctx, func(addTx database.DB) error {
				err := addTx.UserEmails().Add(ctx, oldUser.ID, newEmail, nil)
				if err != nil {
					return err
				}
				return addTx.UserEmails().SetVerified(ctx, oldUser.ID, newEmail, true)

			})
			// Only return this error if this was the new primary
			if addErr != nil && isPrimaryEmail {
				return addErr
			}
		}

		// Now verify any addresses that already existed and weren't verified
		for _, verifyEmail := range updates.toVerify {
			// We only need to ensure that this is successful if it is the primary email
			isPrimaryEmail := updates.resetPrimaryTo != nil && strings.EqualFold(*updates.resetPrimaryTo, verifyEmail)
			db := getTx(isPrimaryEmail)
			verifyErr := db.UserEmails().SetVerified(ctx, oldUser.ID, verifyEmail, true)
			// Only return this error if this was the new primary
			if verifyErr != nil && isPrimaryEmail {
				return verifyErr
			}
		}

		// Now that all the new emails are added and verified set the primary email
		// The primary would be included in the tx because it either already existed
		// or we required the add to succeed in the prior steps
		if updates.resetPrimaryTo != nil {
			db := getTx(true)
			setPrimaryErr := db.UserEmails().SetPrimaryEmail(ctx, oldUser.ID, *updates.resetPrimaryTo)
			if setPrimaryErr != nil {
				return setPrimaryErr
			}
		}

		// Finally remove any email address no need to error or fail the tx here
		for _, newEmail := range updates.toRemove {
			db := getTx(false)
			db.UserEmails().Remove(ctx, oldUser.ID, newEmail)
		}
	}

	return
}

type emailUpdates struct {
	toRemove       []string
	toAdd          []string
	toVerify       []string
	resetPrimaryTo *string
}

func generateEmailUpdates(startingUserData, endingUserData scim.ResourceAttributes, currentEmails []*database.UserEmail) emailUpdates {
	startingPrimary, startingOther := extractPrimaryEmail(startingUserData)
	endingPrimary, endingOther := extractPrimaryEmail(endingUserData)
	result := emailUpdates{}

	// Make a map of existing emails and verification status that we can use for lookup
	currentEmailVerificationStatus := map[string]bool{}
	for _, email := range currentEmails {
		currentEmailVerificationStatus[email.Email] = email.VerifiedAt != nil
	}

	// Check if primary changed
	if !strings.EqualFold(startingPrimary, endingPrimary) {
		result.resetPrimaryTo = &endingPrimary
	}

	toMap := func(s string, others []string) map[string]bool {
		m := map[string]bool{s: true}
		for _, v := range others {
			m[v] = true
		}
		return m
	}

	difference := func(setA, setB map[string]bool) []string {
		result := []string{}
		for a := range setA {
			if !setB[a] {
				result = append(result, a)
			}
		}
		return result
	}

	// Put the original and ending lists of emails into maps to easier comparison
	startingEmails := toMap(startingPrimary, startingOther)
	endingEmails := toMap(endingPrimary, endingOther)

	// Identify emails that were removed
	result.toRemove = difference(startingEmails, endingEmails)

	// Using our ending list of emails check if they already exist
	// If they don't exist we need to add & verify
	// If they do exist but aren't verified we need to verify them
	for email := range endingEmails {
		verified, alreadyExists := currentEmailVerificationStatus[email]
		switch {
		case alreadyExists && !verified:
			result.toVerify = append(result.toVerify, email)
		case !alreadyExists:
			result.toAdd = append(result.toAdd, email)
		}
	}
	return result
}

// getUniqueExternalID extracts the external identifier from the given attributes.
// If it's not present, it returns a unique identifier based on the primary email address of the user.
// We need this because the account ID must be unique across all SCIM accounts that we have on file.
func getUniqueExternalID(attributes scim.ResourceAttributes) string {
	if attributes[AttrExternalId] != nil {
		return attributes[AttrExternalId].(string)
	}
	primary, _ := extractPrimaryEmail(attributes)
	return "no-external-id-" + primary
}

// getOptionalExternalID extracts the external identifier of the given attributes.
func getOptionalExternalID(attributes scim.ResourceAttributes) optional.String {
	if eID, ok := attributes[AttrExternalId]; ok {
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

// convertUserToSCIMResource converts a Sourcegraph user to a SCIM resource.
func convertUserToSCIMResource(user *types.UserForSCIM) scim.Resource {
	// Convert account data – if it doesn't exist, never mind
	attributes, err := fromAccountData(user.SCIMAccountData)
	if err != nil {
		// Failed to convert account data to SCIM resource attributes. Fall back to core user data.
		attributes = scim.ResourceAttributes{
			AttrUserName:    user.Username,
			AttrDisplayName: user.DisplayName,
			AttrName:        map[string]interface{}{AttrNameFormatted: user.DisplayName},
		}
		if user.SCIMExternalID != "" {
			attributes[AttrExternalId] = user.SCIMExternalID
		}
	}
	if attributes[AttrName] == nil {
		attributes[AttrName] = map[string]interface{}{}
	}

	// Fall back to username and primary email in the user object if not set in account data
	if attributes[AttrUserName] == nil || attributes[AttrUserName].(string) == "" {
		attributes[AttrUserName] = user.Username
	}
	if emails, ok := attributes[AttrEmails].([]interface{}); (!ok || len(emails) == 0) && user.Emails != nil && len(user.Emails) > 0 {
		attributes[AttrEmails] = []interface{}{
			map[string]interface{}{
				"value":   user.Emails[0],
				"primary": true,
			},
		}
	}

	return scim.Resource{
		ID:         strconv.FormatInt(int64(user.ID), 10),
		ExternalID: getOptionalExternalID(attributes),
		Attributes: attributes,
		Meta: scim.Meta{
			Created:      &user.CreatedAt,
			LastModified: &user.UpdatedAt,
		},
	}
}
