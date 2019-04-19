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

	return &Client{conn: conn}, conn.ConnectContext(ctx)
}

type getRawDiffRequest struct {
	DiffID int `json:"diffID"`
}

type getRawDiffResponse struct {
	Diff         *string `json:"result"`
	ErrorCode    *string `json:"error_code"`
	ErrorMessage *string `json:"error_info"`
}

func (c *Client) GetRawDiff(ctx context.Context, diffID int) (string, error) {
	req := getRawDiffRequest{DiffID: diffID}

	var res getRawDiffResponse
	err := c.conn.CallContext(ctx, "differential.getrawdiff", &req, &res)
	if err != nil {
		return "", err
	}

	if res.ErrorMessage != nil {
		return "", errors.Errorf("phabricator error: %s %s", *res.ErrorCode, *res.ErrorMessage)
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

type getDiffInfoRequest struct {
	IDs []int `json:"ids"`
}

type getDiffInfoResponse struct {
	// Infos is a map of string(diffID) -> DiffInfo
	// See this page for more information https://phabricator.sgdev.org/conduit/method/differential.querydiffs/
	Infos        *map[string]DiffInfo `json:"result"`
	ErrorCode    *string              `json:"error_code"`
	ErrorMessage *string              `json:"error_info"`
}

func (c *Client) GetDiffInfo(ctx context.Context, diffID int) (*DiffInfo, error) {
	req := getDiffInfoRequest{IDs: []int{diffID}}

	var res getDiffInfoResponse
	err := c.conn.CallContext(ctx, "differential.querydiffs", &req, &res)
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
