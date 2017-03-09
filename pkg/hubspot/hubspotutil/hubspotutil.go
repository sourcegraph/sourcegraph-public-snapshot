package hubspotutil

import (
	"errors"
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
)

// AfterPrivateCodeSignupFormID is a hardcoded identifier for the
// "After Private Code Signup" HubSpot form. To make changes to
// variables passed to this form, edit it in the HubSpot console.
const AfterPrivateCodeSignupFormID = "e9443a20-ef14-4858-971e-4925a3c6c45c"

// ChangeUserPlanFormID is an identifier for the "Change user plan form"
const ChangeUserPlanFormID = "198ad76b-a88c-4b79-b026-bf0588bb2f9f"

var client *hubspot.Client

func init() {
	// TODO(dan): replace this with an env variable (e.g. see mailchimputil.go)
	portalID := "2762526"
	if portalID != "" {
		client = hubspot.New(portalID)
	}
}

// Client returns a hubspot client, or an error if HUBSPOT_KEY is not set.
func Client() (*hubspot.Client, error) {
	if client == nil {
		return nil, errors.New("hubspotutil.Client: authorization key only available on production environments")
	}
	return client, nil
}

// GetFormID returns a valid HubSpot API form ID
func GetFormID(formName string) (string, error) {
	switch formName {
	case "AfterSignupForm":
		return AfterPrivateCodeSignupFormID, nil
	case "ChangeUserPlan":
		return ChangeUserPlanFormID, nil
	default:
		return "", fmt.Errorf("hubspotutil.GetFormID: '%s' is not a valid form", formName)
	}
}
