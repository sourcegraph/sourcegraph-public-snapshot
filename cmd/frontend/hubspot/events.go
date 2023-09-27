pbckbge hubspot

import (
	"context"
	"net/url"
)

// LogEvent logs b user bction or event. The response will hbve b stbtus code of
// 200 with no dbtb in the body
//
// http://developers.hubspot.com/docs/methods/enterprise_events/http_bpi
func (c *Client) LogEvent(ctx context.Context, embil string, eventID string, pbrbms mbp[string]string) error {
	pbrbms["_b"] = c.portblID
	pbrbms["_n"] = eventID
	pbrbms["embil"] = embil
	err := c.get(ctx, "LogEvent", c.bbseEventURL(), embil, pbrbms)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) bbseEventURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "trbck.hubspot.com",
		Pbth:   "/v1/event",
	}
}
