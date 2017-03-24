package hubspotutil

import (
	"errors"
	"unicode"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
)

// FormNameToHubSpotID is a mapping from form names provided by backend or API
// requests to submit HubSpot forms
//
// HubSpot IDs are all defined in HubSpot "Forms" web console:
// https://app.hubspot.com/forms/2762526/
var FormNameToHubSpotID = map[string]string{
	// AfterPrivateCodeSignupForm represents the
	// "After Private Code Signup" form
	"AfterSignupForm": "e9443a20-ef14-4858-971e-4925a3c6c45c",
	// BetaSignupForm represents the "Beta Signup" form
	"BetaSignupForm": "105a5d66-64e8-4993-bb75-797fb725ab85",
	// ChangeUserPlan represents the "Change user plan form"
	"ChangeUserPlan": "198ad76b-a88c-4b79-b026-bf0588bb2f9f",
	// ZapBetaSignupForm is an identifier for the "Zap Beta Signup" form
	"ZapBetaSignupForm": "776431cf-bf0d-4318-9275-b14be48805ea",
}

// EventNameToHubSpotID is a mapping from event names provided by backend
// or API requests to track HubSpot events
//
// HubSpot Events and IDs are all defined in HubSpot "Events" web console:
// https://app.hubspot.com/reports/2762526/events
var EventNameToHubSpotID = map[string]string{
	// ZapAuthCompleted is an identifier for the "ZapAuthCompleted" event
	"ZapAuthCompleted": "000001981045",
}

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

// PrepareFormData does any required preprocessing for individual forms
func PrepareFormData(formName string, formData map[string]string) map[string]string {
	// Convert all form keys to snake case, per HubSpot requirements
	data := make(map[string]string)
	for key, value := range formData {
		if key == "hubSpotFormName" {
			continue
		}
		data[ToSnake(key)] = value
	}

	// Always set `company` and `lastname` fields to "Unknown" if not set by the user.
	// Salesforce requires these fields to be non-blank to do data ingestion from
	// HubSpot, so ensure they always have a value
	if formName == "AfterSignupForm" {
		if company := data["company"]; company == "" {
			data["company"] = "Unknown"
		}
		if lastname := data["lastname"]; lastname == "" {
			data["lastname"] = "Unknown"
		}
	}
	return data
}

// ToSnake convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
//
// Source: https://gist.github.com/elwinar/14e1e897fdbe4d3432e1
func ToSnake(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}
