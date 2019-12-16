package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

func (c *Client) GetDumps(ctx context.Context, args *struct {
	RepoName        string
	Query           *string
	IsLatestForRepo *bool
	Limit           *int32
	Cursor          *string
}) ([]*lsif.LSIFDump, string, int, error) {
	query := queryValues{}
	query.SetOptionalString("query", args.Query)
	query.SetOptionalBool("visibleAtTip", args.IsLatestForRepo)
	query.SetOptionalInt32("limit", args.Limit)

	req := &lsifRequest{
		path:   fmt.Sprintf("/dumps/%s", url.PathEscape(args.RepoName)),
		cursor: args.Cursor,
		query:  query,
	}

	payload := struct {
		Dumps      []*lsif.LSIFDump `json:"dumps"`
		TotalCount int              `json:"totalCount"`
	}{}

	meta, err := c.do(ctx, req, &payload)
	if err != nil {
		return nil, "", 0, err
	}

	return payload.Dumps, meta.nextURL, payload.TotalCount, nil
}

func (c *Client) GetDump(ctx context.Context, args *struct {
	RepoName string
	DumpID   int64
}) (*lsif.LSIFDump, error) {
	req := &lsifRequest{
		path: fmt.Sprintf("/dumps/%s/%d", url.PathEscape(args.RepoName), args.DumpID),
	}

	payload := &lsif.LSIFDump{}
	_, err := c.do(ctx, req, &payload)
	return payload, err
}

func (c *Client) DeleteDump(ctx context.Context, args *struct {
	RepoName string
	DumpID   int64
}) error {
	req := &lsifRequest{
		path:   fmt.Sprintf("/dumps/%s/%d", url.PathEscape(args.RepoName), args.DumpID),
		method: "DELETE",
	}

	if _, err := c.do(ctx, req, nil); IsNotFound(err) {
		return err
	}

	return nil
}
