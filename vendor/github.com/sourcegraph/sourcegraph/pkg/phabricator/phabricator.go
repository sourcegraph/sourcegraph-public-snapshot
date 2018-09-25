// phabricator is a package to interact with a Phabricator instance and its Conduit API.
package phabricator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	url   string
	token string
}

func NewClient(url, token string) *Client {
	return &Client{
		url:   url,
		token: token,
	}
}

func (c *Client) post(path string, payload url.Values, target interface{}) error {
	payload.Add("api.token", c.token)

	res, err := http.PostForm(c.buildURL(path), payload)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	return json.NewDecoder(res.Body).Decode(&target)
}

func (c *Client) buildURL(path string) string {
	return fmt.Sprintf("%s/api/%s", c.url, path)
}

type getRawDiffResult struct {
	Diff         *string `json:"result"`
	ErrorCode    *string `json:"error_code"`
	ErrorMessage *string `json:"error_info"`
}

func (c *Client) GetRawDiff(diffID int) (string, error) {
	payload := url.Values{}
	payload.Add("diffID", strconv.Itoa(diffID))

	var res getRawDiffResult

	path := "differential.getrawdiff"
	err := c.post(path, payload, &res)
	if err != nil {
		return "", err
	}

	if res.ErrorMessage != nil {
		return "", fmt.Errorf("phabricator error: %s %s", *res.ErrorCode, *res.ErrorMessage)
	}

	if res.Diff == nil {
		return "", errors.New("phabricator differential.getrawdiff is null")
	}

	return *res.Diff, nil
}

// DiffInfo contains information for a diff such as the author
type DiffInfo struct {
	Message     string    `json:"description"`
	AuthorName  string    `json:"authorName"`
	AuthorEmail string    `json:"authorEmail"`
	DateCreated string    `json:"dateCreated"`
	Date        time.Time `json:"omitempty"`
}

type getDiffInfoResult struct {
	// Infos is a map of string(diffID) -> DiffInfo
	// See this page for more information https://phabricator.sgdev.org/conduit/method/differential.querydiffs/
	Infos        *map[string]DiffInfo `json:"result"`
	ErrorCode    *string              `json:"error_code"`
	ErrorMessage *string              `json:"error_info"`
}

func (c *Client) GetDiffInfo(diffID int) (*DiffInfo, error) {
	payload := url.Values{}

	payload.Add("ids[0]", strconv.Itoa(diffID))

	var res getDiffInfoResult

	path := "differential.querydiffs"
	err := c.post(path, payload, &res)
	if err != nil {
		return nil, err
	}

	if res.ErrorMessage != nil {
		return nil, errors.Errorf("phabricator error: %s %s", *res.ErrorCode, *res.ErrorMessage)
	}

	if res.Infos == nil {
		return nil, errors.Errorf("phabricator error: no result for diff %d", diffID)
	}

	info, ok := (*res.Infos)[strconv.Itoa(diffID)]
	if !ok {
		return nil, errors.Errorf("phabricator error: no diff info found for diff %d", diffID)
	}

	date, err := ParseDate(info.DateCreated)
	if err != nil {
		return nil, err
	}

	info.Date = *date

	return &info, nil
}

func ParseDate(secStr string) (*time.Time, error) {
	seconds, err := strconv.ParseInt("1524682853", 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "phabricator: could not parse date")
	}
	t := time.Unix(seconds, 0)
	return &t, nil
}
