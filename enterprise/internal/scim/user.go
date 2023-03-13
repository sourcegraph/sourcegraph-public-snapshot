package scim

import (
	"bytes"
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
		schemaExtensions: []scim.SchemaExtension{},
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
func updateUser(ctx context.Context, tx database.DB, oldUser *types.UserForSCIM, updatedUserSCIMAttributes scim.ResourceAttributes, emailsModified bool) (err error) {
	usernameUpdate := ""
	// Get a copy of the user SCIM resources before updates were applied so we can diff them if needed
	beforeUpdateUserSCIMResources := convertUserToSCIMResource(oldUser)

	requestedUsername := extractStringAttribute(updatedUserSCIMAttributes, AttrUserName)
	if requestedUsername != oldUser.Username {
		usernameUpdate, err = getUniqueUsername(ctx, tx.Users(), requestedUsername)
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
	err = tx.Users().Update(ctx, oldUser.ID, userUpdate)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}

	accountData, err := toAccountData(updatedUserSCIMAttributes)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	err = tx.UserExternalAccounts().UpsertSCIMData(ctx, oldUser.ID, getUniqueExternalID(updatedUserSCIMAttributes), accountData)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}

	if emailsModified {
		currentEmails, err := tx.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{UserID: oldUser.ID, OnlyVerified: false})
		if err != nil {
			return err
		}
		diffs := diffEmails(beforeUpdateUserSCIMResources.Attributes, updatedUserSCIMAttributes, currentEmails)
		// First add any new email address
		for _, newEmail := range diffs.toAdd {
			err = tx.UserEmails().Add(ctx, oldUser.ID, newEmail, nil)
			if err != nil {
				return err
			}
			err = tx.UserEmails().SetVerified(ctx, oldUser.ID, newEmail, true)
			if err != nil {
				return err
			}
		}

		// Now verify any addresses that already existed but weren't verified
		for _, email := range diffs.toVerify {
			err = tx.UserEmails().SetVerified(ctx, oldUser.ID, email, true)
			if err != nil {
				return err
			}
		}

		// Now that all the new emails are added and verified set the primary email if it changed
		if diffs.setPrimaryEmailTo != nil {
			err = tx.UserEmails().SetPrimaryEmail(ctx, oldUser.ID, *diffs.setPrimaryEmailTo)
			if err != nil {
				return err
			}
		}

		// Finally remove any email addresses that no longer are needed
		for _, email := range diffs.toRemove {
			err = tx.UserEmails().Remove(ctx, oldUser.ID, email)
			if err != nil {
				return err
			}
		}
	}

	return
}

type emailDiffs struct {
	toRemove          []string
	toAdd             []string
	toVerify          []string
	setPrimaryEmailTo *string
}

//	diffEmails compares the email addresses from the user_emails table to their SCIM data before and after the current update
//	and determines what changes need to be made. It takes into account the current email addresses and verification status from the database
//
// (emailsInDB) to determine if emails need to be added, verified or removed, and if the primary email needs to be changed.
//
//		Parameters:
//		    beforeUpdateUserData - The SCIM resource attributes containing the user's email addresses prior to the update.
//		    afterUpdateUserData - The SCIM resource attributes containing the user's email addresses after the update.
//		    emailsInDB - The current email addresses and verification status for the user from the database.
//
//		Returns:
//		    emailDiffs - A struct containing the email changes that need to be made:
//		     toRemove - Email addresses that need to be removed.
//		     toAdd - Email addresses that need to be added.
//		     toVerify - Existing email addresses that should be marked as verified.
//	         setPrimaryEmailTo - The new primary email address if it changed, otherwise nil.
func diffEmails(beforeUpdateUserData, afterUpdateUserData scim.ResourceAttributes, emailsInDB []*database.UserEmail) emailDiffs {
	beforePrimary, beforeOthers := extractPrimaryEmail(beforeUpdateUserData)
	afterPrimary, afterOthers := extractPrimaryEmail(afterUpdateUserData)
	result := emailDiffs{}

	// Make a map of existing emails and verification status that we can use for lookup
	currentEmailVerificationStatus := map[string]bool{}
	for _, email := range emailsInDB {
		currentEmailVerificationStatus[email.Email] = email.VerifiedAt != nil
	}

	// Check if primary changed
	if !strings.EqualFold(beforePrimary, afterPrimary) && afterPrimary != "" {
		result.setPrimaryEmailTo = &afterPrimary
	}

	toMap := func(s string, others []string) map[string]bool {
		m := map[string]bool{}
		for _, v := range append([]string{s}, others...) {
			if v != "" { // don't include empty strings
				m[v] = true
			}
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
	startingEmails := toMap(beforePrimary, beforeOthers)
	endingEmails := toMap(afterPrimary, afterOthers)

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
func checkBodyNotEmpty(r *http.Request) (err error) {
	data, err := io.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}

		if err == nil {
			// Restore the original body so that it can be read by a next handler.
			r.Body = io.NopCloser(bytes.NewBuffer(data))
		}
	}(r.Body)

	if err != nil {
		return
	}
	if len(data) == 0 {
		return scimerrors.ScimErrorBadParams([]string{"request body is empty"})
	}
	return
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
