package hubspot

// SubmitForm submits form data.  Form submissions return an empty
// body with status code 204 or 302 if submission was successful.
//
// See http://developers.hubspot.com/docs/methods/forms/submit_form.
func (c *Client) SubmitForm(formID string, params map[string]string) error {
	err := c.post("SubmitForm", formID, params)
	if err != nil {
		return err
	}
	return nil
}
