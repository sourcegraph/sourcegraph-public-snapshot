package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

func (c *Client) GetUploads(ctx context.Context, args *struct {
	RepoID          api.RepoID
	Query           *string
	State           *string
	IsLatestForRepo *bool
	Limit           *int32
	Cursor          *string
}) ([]*lsif.LSIFUpload, string, *int, error) {
	query := queryValues{}
	query.SetOptionalString("query", args.Query)
	query.SetOptionalBool("visibleAtTip", args.IsLatestForRepo)
	query.SetOptionalInt32("limit", args.Limit)

	if args.State != nil {
		query.Set("state", strings.ToLower(*args.State))
	}

	req := &lsifRequest{
		path:   fmt.Sprintf("/uploads/repository/%d", args.RepoID),
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
	UploadID int64
}) (*lsif.LSIFUpload, error) {
	req := &lsifRequest{
		path: fmt.Sprintf("/uploads/%d", args.UploadID),
	}

	payload := &lsif.LSIFUpload{}
	_, err := c.do(ctx, req, &payload)
	return payload, err
}

func (c *Client) DeleteUpload(ctx context.Context, args *struct {
	UploadID int64
}) error {
	req := &lsifRequest{
		path:   fmt.Sprintf("/uploads/%d", args.UploadID),
		method: "DELETE",
	}

	if _, err := c.do(ctx, req, nil); !IsNotFound(err) {
		return err
	}

	return nil
}
