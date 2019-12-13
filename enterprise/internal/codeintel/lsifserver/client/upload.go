package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

func (c *Client) GetUploadStats(ctx context.Context) (*lsif.LSIFUploadStats, error) {
	req := &lsifRequest{
		path: "/uploads/stats",
	}

	payload := &lsif.LSIFUploadStats{}
	_, err := c.do(ctx, req, &payload)
	return payload, err
}

func (c *Client) GetUploads(ctx context.Context, args *struct {
	State  string
	Query  *string
	Limit  *int32
	Cursor *string
}) ([]*lsif.LSIFUpload, string, *int, error) {
	query := queryValues{}
	query.SetOptionalString("query", args.Query)
	query.SetOptionalInt32("limit", args.Limit)

	req := &lsifRequest{
		path:   fmt.Sprintf("/uploads/%s", strings.ToLower(args.State)),
		cursor: args.Cursor,
		query:  query,
	}

	payload := struct {
		Uploads    []*lsif.LSIFUpload `json:"uploads"`
		TotalCount *int               `json:"totalCount"`
	}{
		Uploads: []*lsif.LSIFUpload{},
	}

	meta, err := c.do(ctx, req, &payload)
	if err != nil {
		return nil, "", nil, err
	}

	return payload.Uploads, meta.nextURL, payload.TotalCount, nil
}

func (c *Client) GetUpload(ctx context.Context, args *struct {
	UploadID string
}) (*lsif.LSIFUpload, error) {
	req := &lsifRequest{
		path: fmt.Sprintf("/uploads/%s", url.PathEscape(args.UploadID)),
	}

	payload := &lsif.LSIFUpload{}
	_, err := c.do(ctx, req, &payload)
	return payload, err
}

func (c *Client) DeleteUpload(ctx context.Context, args *struct {
	UploadID string
}) error {
	req := &lsifRequest{
		path:   fmt.Sprintf("/uploads/%s", url.PathEscape(args.UploadID)),
		method: "DELETE",
	}

	if _, err := c.do(ctx, req, nil); IsNotFound(err) {
		return err
	}

	return nil
}
