package hubspot

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
)

// CreateOrUpdateContact creates or updates a HubSpot contact (with email as primary key)
//
// The endpoint returns 200 with the contact's VID and an isNew field on success,
// or a 409 Conflict if we attempt to change a user's email address to a new one
// that is already taken
//
// http://developers.hubspot.com/docs/methods/contacts/create_or_update
func (c *Client) CreateOrUpdateContact(email string, params *ContactProperties) error {
	payload, err := json.Marshal(newAPIValues(params))
	if err != nil {
		return err
	}
	return c.postJSON("CreateOrUpdateContact", c.baseContactURL(email), "", string(payload))
}

func (c *Client) baseContactURL(email string) *url.URL {
	q := url.Values{}
	q.Set("hapikey", c.hapiKey)

	return &url.URL{
		Scheme:   "https",
		Host:     "api.hubapi.com",
		Path:     "/contacts/v1/contact/createOrUpdate/email/" + email + "/",
		RawQuery: q.Encode(),
	}
}

// ContactProperties represent HubSpot user properties updated on
// signup or login
type ContactProperties struct {
	UserID         string `json:"user_id"`
	UID            string `json:"uid"`
	GitHubLink     string `json:"github_link"`
	LookerLink     string `json:"looker_link"`
	GitHubName     string `json:"github_name"`
	GitHubCompany  string `json:"github_company"`
	GitHubLocation string `json:"location"`
	// Per HubSpot API, dates should be formatted in milliseconds, in UTC
	// http://developers.hubspot.com/docs/faq/how-should-timestamps-be-formatted-for-hubspots-apis
	RegisteredAt      int64 `json:"registered_at"`
	IsPrivateCodeUser bool  `json:"is_private_code_user"`
}

// newAPIValues converts a ContactProperties struct to a HubSpot API-compliant
// array of key-value pairs
func newAPIValues(h *ContactProperties) *apiProperties {
	apiProps := &apiProperties{}
	apiProps.set("user_id", h.UserID)
	apiProps.set("uid", h.UID)
	apiProps.set("github_link", h.GitHubLink)
	apiProps.set("looker_link", h.LookerLink)
	apiProps.set("github_name", h.GitHubName)
	apiProps.set("github_company", h.GitHubCompany)
	apiProps.set("location", h.GitHubLocation)
	apiProps.set("registered_at", h.RegisteredAt)
	apiProps.set("is_private_code_user", h.IsPrivateCodeUser)
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

func (h *apiProperties) set(property string, value interface{}) {
	if h.Properties == nil {
		h.Properties = make([]*apiProperty, 0)
	}
	if value != reflect.Zero(reflect.TypeOf(value)).Interface() {
		h.Properties = append(h.Properties, &apiProperty{Property: property, Value: fmt.Sprintf("%v", value)})
	}
}
