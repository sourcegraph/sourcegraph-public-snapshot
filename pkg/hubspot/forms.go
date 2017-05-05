package hubspot

import (
	"net/url"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// SubmitForm submits form data.  Form submissions return an empty
// body with status code 204 or 302 if submission was successful.
//
// See http://developers.hubspot.com/docs/methods/forms/submit_form.
func (c *Client) SubmitForm(formID string, form *sourcegraph.SubmittedForm) error {
	err := c.postForm("SubmitForm", c.baseFormURL(), formID, form)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) baseFormURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "forms.hubspot.com",
		Path:   "/uploads/form/v2/" + c.portalID,
	}
}
