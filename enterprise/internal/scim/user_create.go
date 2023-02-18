package scim

import (
	"net/http"
	"strconv"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
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
	username := extractUsername(attributes)
	displayName := extractDisplayName(attributes)

	// Create user (with or without external ID)
	// TODO: Use NewSCIMUser instead of NewUser?
	newUser := database.NewUser{
		Email:           primaryEmail,
		Username:        username,
		DisplayName:     displayName,
		EmailIsVerified: true,
	}
	var user *types.User
	var err error
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
		if dbErr, ok := containsDBError(err); ok {
			if code := dbErr.Code(); code == database.ErrorCodeUsernameExists || code == database.ErrorCodeEmailExists {
				return scim.Resource{}, scimerrors.ScimError{Status: http.StatusConflict, Detail: err.Error()}
			}
		}
		return scim.Resource{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
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

// containsDBError returns true if the given error contains at least one database.ErrCannotCreateUser.
// It also returns the first such error.
func containsDBError(err error) (database.ErrCannotCreateUser, bool) {
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
