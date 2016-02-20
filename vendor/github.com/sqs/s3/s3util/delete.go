package s3util

import (
	"io"
	"net/http"
	"time"
)

// Delete deletes the S3 object at url. An HTTP status other than 204 (No
// Content) is considered an error.
//
// If c is nil, Delete uses DefaultConfig.
func Delete(url string, c *Config) (io.ReadCloser, error) {
	if c == nil {
		c = DefaultConfig
	}
	r, _ := http.NewRequest("DELETE", url, nil)
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Sign(r, *c.Keys)
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return nil, newRespError(resp)
	}
	return resp.Body, nil
}
