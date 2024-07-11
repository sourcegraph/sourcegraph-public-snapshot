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
	// The cookie name for tracking the last visited URL within the current session has been updated to 'first_page_seen_url',
	// however, in the HubSpot integration, we are still using the legacy HubSpot contact property 'first_source_url'
	// for backward compatibility with existing HubSpot configurations and workflows.
	FirstSourceURL string `json:"first_source_url"`

	// The URL of the last page the user saw across multiple cookie duration sessions.
	// Same logic as 'LastsourceUrl' property but storing values across various cookie duration sessions.
	LastPageSeenShort string `json:"last_page_seen_short"`
	LastPageSeenMid   string `json:"last_page_seen_mid"`
	LastPageSeenLong  string `json:"last_page_seen_long"`

	// The URL of the first page a user landed on their latest session on a Sourcegraph site.
	// The cookie name for tracking the last visited URL within the current session has been updated to 'last_page_seen_url',
	// however, in the HubSpot integration, we are still using the legacy HubSpot contact property 'last_source_url'
	// for backward compatibility with existing HubSpot configurations and workflows.
	LastSourceURL string `json:"last_source_url"`

	// The URL of the first page a user landed on the session when they signed up.
	SignupSessionSourceURL string `json:"signup_session_source_url"`

	// The referrer for a user on their latest session on a Sourcegraph site.
	MostRecentReferrerUrl string `json:"most_recent_referrer_url"`

	// The referrer across multiple cookie duration sessions.
	MostRecentReferrerUrlShort string `json:"most_recent_referrer_url_short"`
	MostRecentReferrerUrlMid   string `json:"most_recent_referrer_url_mid"`
	MostRecentReferrerUrlLong  string `json:"most_recent_referrer_url_long"`

	// The referrer for a user on the session when they signed up.
	SignupSessionReferrer string `json:"signup_session_referrer"`

	// The Private App updates the "recent_utm_campaign" HubSpot Contact Property,
	// while Google Tag Manager (GTM) updates the "utm_campaign" property.
	// We handle these properties separately to ensure data integrity and avoid overwriting each other's values.
	// The UTM campaign associated with the current session.
	SessionUTMCampaign string `json:"recent_utm_campaign"`

	// The UTM campaign across multiple cookie duration sessions.
	UtmCampaignShort string `json:"utm_campaign_short"`
	UtmCampaignMid   string `json:"utm_campaign_mid"`
	UtmCampaignLong  string `json:"utm_campaign_long"`

	// The Private App updates the "recent_utm_source" HubSpot Contact Property,
	// while the Google Tag Manager (GTM) updates the "utm_source" property.
	// We handle these properties separately to ensure data integrity and avoid overwriting each other's values.
	// The UTM source associated with the current session.
	SessionUTMSource string `json:"recent_utm_source"`

	// The UTM source across multiple cookie duration sessions.
	UtmSourceShort string `json:"utm_source_short"`
	UtmSourceMid   string `json:"utm_source_mid"`
	UtmSourceLong  string `json:"utm_source_long"`

	// The Private App updates the "recent_utm_medium" HubSpot Contact Property,
	// while the Google Tag Manager (GTM) updates the "utm_medium" property.
	// We handle these properties separately to ensure data integrity and avoid overwriting each other's values.
	// The UTM medium associated with the current session.
	SessionUTMMedium string `json:"recent_utm_medium"`

	// The UTM medium across various cookie sessions.
	UtmMediumShort string `json:"utm_medium_short"`
	UtmMediumMid   string `json:"utm_medium_mid"`
	UtmMediumLong  string `json:"utm_medium_long"`

	// The UTM term associated with the current session.
	// The Private App updates the "recent_utm_term" HubSpot Contact Property,
	// while the Google Tag Manager (GTM) updates the "utm_term".
	// We handle these properties separately to ensure data integrity and avoid overwriting each other's values.
	SessionUTMTerm string `json:"recent_utm_term"`

	// The UTM term across multiple cookie duration sessions.
	UtmTermShort string `json:"utm_term_short"`
	UtmTermMid   string `json:"utm_term_mid"`
	UtmTermLong  string `json:"utm_term_long"`

	// The Private App updates the "recent_utm_contecnt" HubSpot Contact Property,
	// while the Google Tag Manager (GTM) updates the "utm_content" property.
	// We handle these properties separately to ensure data integrity and avoid overwriting each other's values.
	// The UTM content associated with the current session.
	SessionUTMContent string `json:"recent_utm_content"`

	// The UTM content across multiple cookie duration sessions.
	UtmContentShort string `json:"utm_content_short"`
	UtmContentMid   string `json:"utm_content_mid"`
	UtmContentLong  string `json:"utm_content_long"`

	// The Google Ads click ID
	GoogleClickID string `json:"gclid"`

	// The Microsoft Ads click ID
	MicrosoftClickID string `json:"msclkid"`
}

// ContactResponse represents HubSpot user properties returned
// after a CreateOrUpdate API call
type ContactResponse struct {
	VID   uint64 `json:"vid"`
	IsNew bool   `json:"isNew"`
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
	apiProps.set("first_source_url", h.FirstSourceURL)
	apiProps.set("last_source_url", h.LastSourceURL)
	apiProps.set("last_page_seen_url_short", h.LastPageSeenShort)
	apiProps.set("last_page_seen_url_mid", h.LastPageSeenMid)
	apiProps.set("last_page_seen_url_long", h.LastPageSeenLong)
	apiProps.set("signup_session_source_url", h.SignupSessionSourceURL)
	apiProps.set("most_recent_referrer_url", h.MostRecentReferrerUrl)
	apiProps.set("most_recent_referrer_url_short", h.MostRecentReferrerUrlShort)
	apiProps.set("most_recent_referrer_url_mid", h.MostRecentReferrerUrlMid)
	apiProps.set("most_recent_referrer_url_long", h.MostRecentReferrerUrlLong)
	apiProps.set("signup_session_referrer", h.SignupSessionReferrer)
	apiProps.set("recent_utm_campaign", h.SessionUTMCampaign)
	apiProps.set("utm_campaign_short", h.UtmCampaignShort)
	apiProps.set("utm_campaign_mid", h.UtmCampaignMid)
	apiProps.set("utm_campaign_long", h.UtmCampaignLong)
	apiProps.set("recent_utm_source", h.SessionUTMSource)
	apiProps.set("utm_source_short", h.UtmSourceShort)
	apiProps.set("utm_source_mid", h.UtmSourceMid)
	apiProps.set("utm_source_long", h.UtmSourceLong)
	apiProps.set("recent_utm_medium", h.SessionUTMMedium)
	apiProps.set("utm_medium_short", h.UtmMediumShort)
	apiProps.set("utm_medium_mid", h.UtmMediumMid)
	apiProps.set("utm_medium_long", h.UtmMediumLong)
	apiProps.set("recent_utm_term", h.SessionUTMTerm)
	apiProps.set("utm_term_short", h.UtmTermShort)
	apiProps.set("utm_term_mid", h.UtmTermMid)
	apiProps.set("utm_term_long", h.UtmTermLong)
	apiProps.set("recent_utm_content", h.SessionUTMContent)
	apiProps.set("utm_content_short", h.UtmContentShort)
	apiProps.set("utm_content_mid", h.UtmContentMid)
	apiProps.set("utm_content_long", h.UtmContentLong)
	apiProps.set("gclid", h.GoogleClickID)
	apiProps.set("msclkid", h.MicrosoftClickID)
	return apiProps
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
