// Package phabricator is a package to interact with a Phabricator instance and its Conduit API.
package phabricator

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/uber/gonduit"
	"github.com/uber/gonduit/core"
	"github.com/uber/gonduit/requests"
)

// A Client provides high level methods to a Phabricator Conduit API.
type Client struct {
	conn *gonduit.Conn
}

// NewClient returns an authenticated Client, using the given URL and
// token. If provided, cli will be used to perform the underlying HTTP requests.
func NewClient(ctx context.Context, url, token string, cli httpcli.Doer) (*Client, error) {
	conn, err := gonduit.DialContext(ctx, url, &core.ClientOptions{
		APIToken: token,
		Client:   cli,
	})

	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) GetRawDiff(ctx context.Context, diffID int) (diff string, err error) {
	type request struct {
		requests.Request
		DiffID int `json:"diffID"`
	}

	req := request{DiffID: diffID}
	err = c.conn.CallContext(ctx, "differential.getrawdiff", &req, &diff)
	if err != nil {
		return "", err
	}

	return diff, nil
}

// DiffInfo contains information for a diff such as the author
type DiffInfo struct {
	Message     string    `json:"description"`
	AuthorName  string    `json:"authorName"`
	AuthorEmail string    `json:"authorEmail"`
	DateCreated string    `json:"dateCreated"`
	Date        time.Time `json:"omitempty"`
}

func (c *Client) GetDiffInfo(ctx context.Context, diffID int) (*DiffInfo, error) {
	type request struct {
		requests.Request
		IDs []int `json:"ids"`
	}

	req := request{IDs: []int{diffID}}

	var res map[string]*DiffInfo
	err := c.conn.CallContext(ctx, "differential.querydiffs", &req, &res)
	if err != nil {
		return nil, err
	}

	info, ok := res[strconv.Itoa(diffID)]
	if !ok {
		return nil, errors.Errorf("phabricator error: no diff info found for diff %d", diffID)
	}

	date, err := ParseDate(info.DateCreated)
	if err != nil {
		return nil, err
	}

	info.Date = *date

	return info, nil
}

func ParseDate(secStr string) (*time.Time, error) {
	seconds, err := strconv.ParseInt(secStr, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "phabricator: could not parse date")
	}
	t := time.Unix(seconds, 0)
	return &t, nil
}
