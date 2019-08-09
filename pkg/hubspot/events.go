package hubspot

import "net/url"

// LogEvent logs a user action or event. The response will have a status code of
// 200 with no data in the body
//
// http://developers.hubspot.com/docs/methods/enterprise_events/http_api
func (c *Client) LogEvent(email string, eventID string, params map[string]string) error {
	params["_a"] = c.portalID
	params["_n"] = eventID
	params["email"] = email
	err := c.get("LogEvent", c.baseEventURL(), email, params)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) baseEventURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "track.hubspot.com",
		Path:   "/v1/event",
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_845(size int) error {
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
