package mailchimp

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/go-querystring/query"
)

// Error represents a Mailchimp error and may be returned by any Client method.
type Error struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("mailchimp.Error{Type=%q, Title=%q, Status=%v, Detail=%q, Instance=%q}", e.Type, e.Title, e.Status, e.Detail, e.Instance)
}

// Client is a Mailchimp V3 API client.
//
// It is safe for use by multiple goroutines concurrently, and generally only
// one API client should be used throughout an application.
type Client struct {
	key, dc string
}

// New returns a new Mailchimp client using the given API key.
//
// New panics if the API key is not a syntactically correct API key (suffixed
// with "-" follow by datacenter ID).
func New(key string) *Client {
	return &Client{
		key: key,
		dc:  datacenter(key),
	}
}

func (c *Client) baseURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   c.dc + ".api.mailchimp.com",
		Path:   "3.0",
	}
}

func (c *Client) post(methodName, suffix string, body, result interface{}) error {
	return c.postPutPatch("POST", methodName, suffix, body, result)
}

func (c *Client) put(methodName, suffix string, body, result interface{}) error {
	return c.postPutPatch("PUT", methodName, suffix, body, result)
}

func (c *Client) patch(methodName, suffix string, body, result interface{}) error {
	return c.postPutPatch("PATCH", methodName, suffix, body, result)
}

func (c *Client) postPutPatch(httpMethodName, methodName, suffix string, body, result interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return wrapError(methodName, err)
	}

	u := c.baseURL()
	u.Path = path.Join(u.Path, suffix)
	req, err := http.NewRequest(httpMethodName, u.String(), bytes.NewReader(data))
	if err != nil {
		return wrapError(methodName, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("user", c.key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return wrapError(methodName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		e := &Error{}
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return wrapError(methodName, err)
		}
		return e
	}
	return wrapError(methodName, json.NewDecoder(resp.Body).Decode(result))
}

func (c *Client) get(methodName, suffix string, body, result interface{}) error {
	q, err := query.Values(result)
	if err != nil {
		return wrapError(methodName, err)
	}

	u := c.baseURL()
	u.Path = path.Join(u.Path, suffix)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return wrapError(methodName, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("user", c.key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return wrapError(methodName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		e := &Error{}
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return wrapError(methodName, err)
		}
		return e
	}
	return wrapError(methodName, json.NewDecoder(resp.Body).Decode(result))
}

func wrapError(methodName string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("mailchimp.%s: %v", methodName, err)
}

// Array returns the given list of strings as a plaintext-string usable within
// Mailchimp as an 'array' of sorts. Basically:
//
//  mailchimpArray("a", "b", "c") == "a,b,c,"
//
// The trailing comma is significant because it allows matching a singular
// element within mailchimp using the "contains" operator. For example:
//
//  mailchimpArray("Visual Studio Code", "Sublime") == "Visual Studio Code,Sublime,"
//
// To match users who have selected Visual Studio (not VSCode), we would use "contains"
// and "Visual Studio,". The trailing comma is important because otherwise the
// query would also match all people who have picked "Visual Studio Code", i.e.
// without a comma the "contains" operation is a substring search.
//
// To match users who signed up with the old single-value registration form,
// just use the following configuration:
//
//  "Subscribers match ___ of the following conditions": "all"
//  "contains": "Visual Studio,"
//  "is": "Visual Studio"
//
func Array(values []string) string {
	return strings.Join(values, ",") + ","
}

// SubscriberHash returns the subscriber hash for the given email address.
func SubscriberHash(emailAddress string) string {
	sum := md5.Sum([]byte(strings.ToLower(emailAddress)))
	return hex.EncodeToString(sum[:])
}

// datacenter returns the datacenter ID from a Mailchimp API key, which is in
// the form of:
//
//  0xdeadbeef0x00000000badbeefbad0-us8
//
func datacenter(key string) string {
	s := strings.Split(key, "-")
	if len(s) != 2 {
		panic("mailchimp: invalid API key, expected <key><dash><datacenter> format")
	}
	return s[1]
}
