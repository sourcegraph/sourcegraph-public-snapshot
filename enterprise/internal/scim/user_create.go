package scim

import (
	"net/http"
	"strconv"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
func (h *UserResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// Get external ID, primary email, username, and display name
	optionalExternalID := getOptionalExternalID(attributes)
	primaryEmail := extractPrimaryEmail(attributes)
	if primaryEmail == "" {
		return scim.Resource{}, scimerrors.ScimErrorBadParams([]string{"emails missing"})
	}
	displayName := extractDisplayName(attributes)

	// Process requested username
	requestedUsername := extractUsername(attributes)
	normalizedUsername, err := auth.NormalizeUsername(requestedUsername)
	if err != nil {
		normalizedUsername, err = auth.AddRandomSuffix("")
		if err != nil {
			return scim.Resource{}, scimerrors.ScimErrorBadParams([]string{"invalid username"})
		}
	}

	// Make sure the username is unique, then create user with/without an external account ID
	var user *types.User
	err = h.db.WithTransact(r.Context(), func(tx database.DB) error {
		_, err := tx.Users().GetByUsername(r.Context(), normalizedUsername)
		if err == nil { // Username exists, try to add random suffix
			normalizedUsername, err = auth.AddRandomSuffix(normalizedUsername)
			if err != nil {
				return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not normalize username").Error()}
			}
		} else if !database.IsUserNotFoundErr(err) {
			return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not check if username exists").Error()}
		}

		// Create user (with or without external ID)
		// TODO: Use NewSCIMUser instead of NewUser?
		newUser := database.NewUser{
			Email:           primaryEmail,
			Username:        normalizedUsername,
			DisplayName:     displayName,
			EmailIsVerified: true,
		}
		if optionalExternalID.Present() {
			accountSpec := extsvc.AccountSpec{
				ServiceType: "scim",
				// TODO: provide proper service ID
				ServiceID: "TODO",
				AccountID: optionalExternalID.Value(),
			}
			user, err = h.db.UserExternalAccounts().CreateUserAndSave(r.Context(), newUser, accountSpec, extsvc.AccountData{})
		} else {
			user, err = h.db.Users().Create(r.Context(), newUser)
		}
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
		return scim.Resource{}, err
	}

	var now = time.Now()

	return scim.Resource{
		ID:         strconv.Itoa(int(user.ID)),
		ExternalID: optionalExternalID,
		Attributes: attributes,
		Meta: scim.Meta{
			Created:      &now,
			LastModified: &now,
		},
	}, nil
}

// extractPrimaryEmail extracts the primary email address from the given attributes.
// Tries to get the (first) email address marked as primary, otherwise uses the first email address it finds.
func extractPrimaryEmail(attributes scim.ResourceAttributes) (primaryEmail string) {
	if attributes["emails"] == nil {
		return
	}
	emails := attributes["emails"].([]interface{})
	for _, emailRaw := range emails {
		email := emailRaw.(map[string]interface{})
		if email["primary"] == true {
			primaryEmail = email["value"].(string)
			break
		}
	}
	if primaryEmail == "" && len(emails) > 0 {
		primaryEmail = emails[0].(map[string]interface{})["value"].(string)
	}
	return
}

// extractUsername extracts the username from the given attributes.
func extractUsername(attributes scim.ResourceAttributes) (username string) {
	// TODO: Validate here?
	if attributes["userName"] != nil {
		username = attributes["userName"].(string)
	}
	return
}

// extractDisplayName extracts the user's display name from the given attributes.
func extractDisplayName(attributes scim.ResourceAttributes) (displayName string) {
	if attributes["displayName"] != nil {
		displayName = attributes["displayName"].(string)
	} else if attributes["name"] != nil {
		name := attributes["name"].(map[string]interface{})
		if name["formatted"] != nil {
			displayName = name["formatted"].(string)
		} else if name["givenName"] != nil && name["familyName"] != nil {
			displayName = name["givenName"].(string) + " " + name["familyName"].(string)
		}
	} else if attributes["userName"] != nil {
		displayName = attributes["userName"].(string)
	}
	return
}

// containsErrCannotCreateUserError returns true if the given error contains at least one database.ErrCannotCreateUser.
// It also returns the first such error.
func containsErrCannotCreateUserError(err error) (database.ErrCannotCreateUser, bool) {
	if err == nil {
		return database.ErrCannotCreateUser{}, false
	}
	if _, ok := err.(database.ErrCannotCreateUser); ok {
		return err.(database.ErrCannotCreateUser), true
	}

	// Handle multiError
	if multiErr, ok := err.(errors.MultiError); ok {
		for _, err := range multiErr.Errors() {
			if _, ok := err.(database.ErrCannotCreateUser); ok {
				return err.(database.ErrCannotCreateUser), true
			}
		}
	}

	return database.ErrCannotCreateUser{}, false
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
