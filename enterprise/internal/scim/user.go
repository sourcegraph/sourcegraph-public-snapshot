package scim

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type User struct {
	types.UserForSCIM
}

func (u *User) ToResource() scim.Resource {
	// Convert account data – if it doesn't exist, never mind
	attributes, err := fromAccountData(u.SCIMAccountData)
	if err != nil {
		first, middle, last := displayNameToPieces(u.DisplayName)
		// Failed to convert account data to SCIM resource attributes. Fall back to core user data.
		attributes = scim.ResourceAttributes{
			AttrActive:      u.Active,
			AttrUserName:    u.Username,
			AttrDisplayName: u.DisplayName,
			AttrName: map[string]interface{}{
				AttrNameFormatted: u.DisplayName,
				AttrNameGiven:     first,
				AttrNameMiddle:    middle,
				AttrNameFamily:    last,
			},
		}
		if u.SCIMExternalID != "" {
			attributes[AttrExternalId] = u.SCIMExternalID
		}
	}
	if attributes[AttrName] == nil {
		attributes[AttrName] = map[string]interface{}{}
	}

	// Fall back to username and primary email in the user object if not set in account data
	if attributes[AttrUserName] == nil || attributes[AttrUserName].(string) == "" {
		attributes[AttrUserName] = u.Username
	}
	if emails, ok := attributes[AttrEmails].([]interface{}); (!ok || len(emails) == 0) && u.Emails != nil && len(u.Emails) > 0 {
		attributes[AttrEmails] = []interface{}{
			map[string]interface{}{
				"value":   u.Emails[0],
				"primary": true,
			},
		}
	}

	return scim.Resource{
		ID:         strconv.FormatInt(int64(u.ID), 10),
		ExternalID: getOptionalExternalID(attributes),
		Attributes: attributes,
		Meta: scim.Meta{
			Created:      &u.CreatedAt,
			LastModified: &u.UpdatedAt,
		},
	}
}

// AccountData stores information about a user that we don't have fields for in the schema.
type AccountData struct {
	Username string `json:"username"`
}

// toAccountData converts the given “SCIM resource attributes” type to an AccountData type.
func toAccountData(attributes scim.ResourceAttributes) (extsvc.AccountData, error) {
	serializedAccountData, err := json.Marshal(attributes)
	if err != nil {
		return extsvc.AccountData{}, err
	}

	return extsvc.AccountData{
		AuthData: nil,
		Data:     extsvc.NewUnencryptedData(serializedAccountData),
	}, nil
}

// fromAccountData converts the given account data JSON to a “SCIM resource attributes” type.
func fromAccountData(scimAccountData string) (attributes scim.ResourceAttributes, err error) {
	err = json.Unmarshal([]byte(scimAccountData), &attributes)
	return
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

// displayNameToPieces splits a display name into first, middle, and last name.
func displayNameToPieces(displayName string) (first, middle, last string) {
	pieces := strings.Fields(displayName)
	switch len(pieces) {
	case 0:
		return "", "", ""
	case 1:
		return pieces[0], "", ""
	case 2:
		return pieces[0], "", pieces[1]
	default:
		return pieces[0], strings.Join(pieces[1:len(pieces)-1], " "), pieces[len(pieces)-1]
	}
}

// Errors

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
