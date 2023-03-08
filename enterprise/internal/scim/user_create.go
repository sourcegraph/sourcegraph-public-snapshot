package scim

import (
	"net/http"
	"strconv"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
func (h *UserResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// Extract external ID, primary email, username, and display name from attributes to variables
	primaryEmail, otherEmails := extractPrimaryEmail(attributes)
	if primaryEmail == "" {
		return scim.Resource{}, scimerrors.ScimErrorBadParams([]string{"emails missing"})
	}
	displayName := extractDisplayName(attributes)

	// Make sure the username is unique, then create user with/without an external account ID
	var user *types.User
	err := h.db.WithTransact(r.Context(), func(tx database.DB) error {
		uniqueUsername, err := getUniqueUsername(r.Context(), tx.Users(), extractStringAttribute(attributes, AttrUserName))
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
		user, err = tx.UserExternalAccounts().CreateUserAndSave(r.Context(), newUser, accountSpec, accountData)

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
			h.db.WithTransact(r.Context(), func(tx database.DB) error {
				err := tx.UserEmails().Add(r.Context(), user.ID, email, nil)
				if err != nil {
					return err
				}
				return tx.UserEmails().SetVerified(r.Context(), user.ID, email, true)
			})
		}
	}

	var now = time.Now()

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

// extractPrimaryEmail extracts the primary email address from the given attributes.
// Tries to get the (first) email address marked as primary, otherwise uses the first email address it finds.
func extractPrimaryEmail(attributes scim.ResourceAttributes) (primaryEmail string, otherEmails []string) {
	if attributes[AttrEmails] == nil {
		return
	}
	emails := attributes[AttrEmails].([]interface{})
	otherEmails = make([]string, 0, len(emails))
	for _, emailRaw := range emails {
		email := emailRaw.(map[string]interface{})
		if email["primary"] == true && primaryEmail == "" {
			primaryEmail = email["value"].(string)
			continue
		}
		otherEmails = append(otherEmails, email["value"].(string))
	}
	if primaryEmail == "" && len(otherEmails) > 0 {
		primaryEmail, otherEmails = otherEmails[0], otherEmails[1:]
	}
	return
}

// extractDisplayName extracts the user's display name from the given attributes.
// Ii defaults to the username if no display name is available.
func extractDisplayName(attributes scim.ResourceAttributes) (displayName string) {
	if attributes[AttrDisplayName] != nil {
		displayName = attributes[AttrDisplayName].(string)
	} else if attributes[AttrName] != nil {
		name := attributes[AttrName].(map[string]interface{})
		if name[AttrNameFormatted] != nil {
			displayName = name[AttrNameFormatted].(string)
		} else if name[AttrNameGiven] != nil && name[AttrNameFamily] != nil {
			if name[AttrNameMiddle] != nil {
				displayName = name[AttrNameGiven].(string) + " " + name[AttrNameMiddle].(string) + " " + name[AttrNameFamily].(string)
			} else {
				displayName = name[AttrNameGiven].(string) + " " + name[AttrNameFamily].(string)
			}
		}
	} else if attributes[AttrNickName] != nil {
		displayName = attributes[AttrNickName].(string)
	}
	// Fallback to username
	if displayName == "" {
		displayName = attributes[AttrUserName].(string)
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
