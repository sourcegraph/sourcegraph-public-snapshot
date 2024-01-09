package hubspot

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CreateOrUpdateContact creates or updates a HubSpot contact (with email as primary key)
//
// The endpoint returns 200 with the contact's VID and an isNew field on success,
// or a 409 Conflict if we attempt to change a user's email address to a new one
// that is already taken
//
// http://developers.hubspot.com/docs/methods/contacts/create_or_update
func (c *Client) CreateOrUpdateContact(email string, params *ContactProperties) (*ContactResponse, error) {
	if c.accessToken == "" {
		return nil, errors.New("HubSpot API key must be provided.")
	}
	var resp ContactResponse
	err := c.postJSON("CreateOrUpdateContact", c.baseContactURL(email), newAPIValues(params), &resp)
	if err != nil {
		return &resp, err
	}
	if resp.IsNew {
		// Certain properties (such as first source URL) should only be sent when a contact is new. Although
		// the user's cookie value should not change, minimize risk of login via multiple browsers, clearing
		// of cookies, etc. by not sending these values on subsequent logins.
		err = c.postJSON("CreateOrUpdateContact", c.baseContactURL(email), firstTimeUserValues(params), &resp)
	}
	return &resp, err
}

func (c *Client) baseContactURL(email string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "api.hubapi.com",
		Path:   "/contacts/v1/contact/createOrUpdate/email/" + email + "/",
	}
}

// ContactProperties represent HubSpot user properties
type ContactProperties struct {
	UserID                       string `json:"user_id"`
	IsServerAdmin                bool   `json:"is_server_admin"`
	LatestPing                   int64  `json:"latest_ping"`
	AnonymousUserID              string `json:"anonymous_user_id"`
	DatabaseID                   int32  `json:"database_id"`
	HasAgreedToToS               bool   `json:"has_agreed_to_tos_and_pp"`
	VSCodyInstalledEmailsEnabled bool   `json:"vs_cody_installed_emails_enabled"`

	// The URL of the first page a user landed on their first session on a Sourcegraph site.
	FirstSourceURL string `json:"first_source_url"`

	// The URL of the first page a user landed on their latest session on a Sourcegraph site.
	LastSourceURL string `json:"last_source_url"`

	// The URL of the first page a user landed on the session when they signed up.
	SignupSessionSourceURL string `json:"signup_session_source_url"`

	// The referrer for a user on their first session on a Sourcegraph site.
	OriginalReferrer string `json:"original_referrer"`

	// The referrer for a user on their latest session on a Sourcegraph site.
	LastReferrer string `json:"most_recent_referrer_url"`

	// The referrer for a user on the session when they signed up.
	SignupSessionReferrer string `json:"signup_session_referrer"`

	// The UTM campaign associated with the current session.
	SessionUTMCampaign string `json:"utm_campaign"`

	// The UTM source associated with the current session.
	SessionUTMSource string `json:"utm_source"`

	// The UTM medium associated with the current session.
	SessionUTMMedium string `json:"utm_medium"`

	// The UTM term associated with the current session.
	SessionUTMTerm string `json:"utm_term"`

	// The UTM content associated with the current session.
	SessionUTMContent string `json:"utm_content"`

	// The Google Ads click ID
	GoogleClickID string `json:"gclid"`

	// The Microsoft Ads click ID
	MicrosoftClickID string `json:"msclkid"`
}

// ContactResponse represents HubSpot user properties returned
// after a CreateOrUpdate API call
type ContactResponse struct {
	VID   int32 `json:"vid"`
	IsNew bool  `json:"isNew"`
}

// newAPIValues converts a ContactProperties struct to a HubSpot API-compliant
// array of key-value pairs
func newAPIValues(h *ContactProperties) *apiProperties {
	apiProps := &apiProperties{}
	apiProps.set("user_id", h.UserID)
	apiProps.set("is_server_admin", h.IsServerAdmin)
	apiProps.set("latest_ping", h.LatestPing)
	apiProps.set("anonymous_user_id", h.AnonymousUserID)
	apiProps.set("database_id", h.DatabaseID)
	apiProps.set("has_agreed_to_tos_and_pp", h.HasAgreedToToS)
	apiProps.set("last_source_url", h.LastSourceURL)
	apiProps.set("signup_session_source_url", h.SignupSessionSourceURL)
	apiProps.set("most_recent_referrer_url", h.LastReferrer)
	apiProps.set("signup_session_referrer", h.SignupSessionReferrer)
	apiProps.set("utm_campaign", h.SessionUTMCampaign)
	apiProps.set("utm_source", h.SessionUTMSource)
	apiProps.set("utm_medium", h.SessionUTMMedium)
	apiProps.set("utm_term", h.SessionUTMTerm)
	apiProps.set("utm_content", h.SessionUTMContent)
	apiProps.set("gclid", h.GoogleClickID)
	apiProps.set("msclkid", h.MicrosoftClickID)
	return apiProps
}

func firstTimeUserValues(h *ContactProperties) *apiProperties {
	firstTimeUserProps := &apiProperties{}
	firstTimeUserProps.set("first_source_url", h.FirstSourceURL)
	firstTimeUserProps.set("original_referrer", h.OriginalReferrer)
	return firstTimeUserProps
}

// apiProperties represents a list of HubSpot API-compliant key-value pairs
type apiProperties struct {
	Properties []*apiProperty `json:"properties"`
}

type apiProperty struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}

func (h *apiProperties) set(property string, value any) {
	if h.Properties == nil {
		h.Properties = make([]*apiProperty, 0)
	}
	if value != reflect.Zero(reflect.TypeOf(value)).Interface() {
		h.Properties = append(h.Properties, &apiProperty{Property: property, Value: fmt.Sprintf("%v", value)})
	}
}
