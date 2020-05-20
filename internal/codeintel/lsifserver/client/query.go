package client

import (
	"context"
	"fmt"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

func (c *Client) Exists(ctx context.Context, args *struct {
	RepoID api.RepoID
	Commit api.CommitID
	Path   string
}) ([]*lsif.LSIFUpload, error) {
	query := queryValues{}
	query.SetInt("repositoryId", int64(args.RepoID))
	query.Set("commit", string(args.Commit))
	query.Set("path", args.Path)

	req := &lsifRequest{
		path:       "/exists",
		query:      query,
		routingKey: fmt.Sprintf("%d:%s", args.RepoID, args.Commit),
	}

	payload := struct {
		Uploads []*lsif.LSIFUpload `json:"uploads"`
	}{}

	_, err := c.do(ctx, req, &payload)
	if err != nil {
		return nil, err
	}

	return payload.Uploads, nil
}

func (c *Client) Definitions(ctx context.Context, args *struct {
	RepoID    api.RepoID
	Commit    api.CommitID
	Path      string
	Line      int32
	Character int32
	UploadID  int64
}) ([]*lsif.LSIFLocation, string, error) {
	return c.locationQuery(ctx, &struct {
		Operation string
		RepoID    api.RepoID
		Commit    api.CommitID
		Path      string
		Line      int32
		Character int32
		UploadID  int64
		Limit     *int32
		Cursor    *string
	}{
		Operation: "definitions",
		RepoID:    args.RepoID,
		Commit:    args.Commit,
		Path:      args.Path,
		Line:      args.Line,
		Character: args.Character,
		UploadID:  args.UploadID,
	})
}

func (c *Client) References(ctx context.Context, args *struct {
	RepoID    api.RepoID
	Commit    api.CommitID
	Path      string
	Line      int32
	Character int32
	UploadID  int64
	Limit     *int32
	Cursor    *string
}) ([]*lsif.LSIFLocation, string, error) {
	return c.locationQuery(ctx, &struct {
		Operation string
		RepoID    api.RepoID
		Commit    api.CommitID
		Path      string
		Line      int32
		Character int32
		UploadID  int64
		Limit     *int32
		Cursor    *string
	}{
		Operation: "references",
		RepoID:    args.RepoID,
		Commit:    args.Commit,
		Path:      args.Path,
		Line:      args.Line,
		Character: args.Character,
		UploadID:  args.UploadID,
		Limit:     args.Limit,
		Cursor:    args.Cursor,
	})
}

func (c *Client) locationQuery(ctx context.Context, args *struct {
	Operation string
	RepoID    api.RepoID
	Commit    api.CommitID
	Path      string
	Line      int32
	Character int32
	UploadID  int64
	Limit     *int32
	Cursor    *string
}) ([]*lsif.LSIFLocation, string, error) {
	query := queryValues{}
	query.SetInt("repositoryId", int64(args.RepoID))
	query.Set("commit", string(args.Commit))
	query.Set("path", args.Path)
	query.SetInt("line", int64(args.Line))
	query.SetInt("character", int64(args.Character))
	query.SetInt("uploadId", int64(args.UploadID))
	query.SetOptionalInt32("limit", args.Limit)

	req := &lsifRequest{
		path:       fmt.Sprintf("/%s", args.Operation),
		cursor:     args.Cursor,
		query:      query,
		routingKey: fmt.Sprintf("%d:%s", args.RepoID, args.Commit),
	}

	payload := struct {
		Locations []*lsif.LSIFLocation
	}{}

	meta, err := c.do(ctx, req, &payload)
	if err != nil {
		return nil, "", err
	}

	return payload.Locations, meta.nextURL, nil
}

func (c *Client) Hover(ctx context.Context, args *struct {
	RepoID    api.RepoID
	Commit    api.CommitID
	Path      string
	Line      int32
	Character int32
	UploadID  int64
}) (string, lsp.Range, error) {
	query := queryValues{}
	query.SetInt("repositoryId", int64(args.RepoID))
	query.Set("commit", string(args.Commit))
	query.Set("path", args.Path)
	query.SetInt("line", int64(args.Line))
	query.SetInt("character", int64(args.Character))
	query.SetInt("uploadId", int64(args.UploadID))

	req := &lsifRequest{
		path:       "/hover",
		query:      query,
		routingKey: fmt.Sprintf("%d:%s", args.RepoID, args.Commit),
	}

	payload := struct {
		Text  string    `json:"text"`
		Range lsp.Range `json:"range"`
	}{}

	_, err := c.do(ctx, req, &payload)
	if err != nil {
		return "", lsp.Range{}, err
	}

	return payload.Text, payload.Range, nil
}
