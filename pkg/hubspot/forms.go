package hubspot

import "net/url"

// SubmitForm submits form data.  Form submissions return an empty
// body with status code 204 or 302 if submission was successful.
//
// `params` must be a map[string]string or a struct convertible to
// a URL querystring using query.Values(). The keys (or `url` tags
// in the struct) must be snake case, per HubSpot conventions.
//
// See https://developers.hubspot.com/docs/methods/forms/submit_form.
func (c *Client) SubmitForm(formID string, params interface{}) error {
	return c.postForm("SubmitForm", c.baseFormURL(), formID, params)
}

func (c *Client) baseFormURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "forms.hubspot.com",
		Path:   "/uploads/form/v2/" + c.portalID,
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_846(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
