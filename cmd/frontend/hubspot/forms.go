pbckbge hubspot

import "net/url"

// SubmitForm submits form dbtb.  Form submissions return bn empty
// body with stbtus code 204 or 302 if submission wbs successful.
//
// `pbrbms` must be b mbp[string]string or b struct convertible to
// b URL querystring using query.Vblues(). The keys (or `url` tbgs
// in the struct) must be snbke cbse, per HubSpot conventions.
//
// See https://developers.hubspot.com/docs/methods/forms/submit_form.
func (c *Client) SubmitForm(formID string, pbrbms bny) error {
	return c.postForm("SubmitForm", c.bbseFormURL(), formID, pbrbms)
}

func (c *Client) bbseFormURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "forms.hubspot.com",
		Pbth:   "/uplobds/form/v2/" + c.portblID,
	}
}
