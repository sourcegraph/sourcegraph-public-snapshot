package client

import (
	"errors"
	"net/url"
)

// Registration is used to collect data needed by
// Enterprise to properly start the registration
// Process.
type Registration struct {
	ClusterID string
	Product   string // influxdb, chronograf, telegraf, kapicator
	// OPTIONAL: Enterprise will redirect the customer back to the
	// RedirectURL if it is specified.
	RedirectURL string
}

// IsValid returns an error if the Registration is not valid.
// This is necessary since there is no server-side validation
// of this data.
func (r Registration) IsValid() error {
	if r.ClusterID == "" || r.Product == "" {
		return errors.New("You must supply both a ClusterID and a Product!")
	}
	return nil
}

// RegistrationURL returns a URL based on the Registration
// data provided. The app can then use this URL to direct
// customers over to the Enterprise application to complete
// their registration.
func (c *Client) RegistrationURL(r Registration) (string, error) {
	err := r.IsValid()
	if err != nil {
		return "", err
	}

	u, _ := url.Parse(c.URL)
	u.Path = "/start"

	q := u.Query()
	q.Set("cluster_id", r.ClusterID)
	q.Set("product", r.Product)
	if r.RedirectURL != "" {
		q.Set("redirect_url", r.RedirectURL)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}
