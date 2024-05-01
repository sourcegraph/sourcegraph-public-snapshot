package msp

import (
	"context"

	"github.com/jomei/notionapi"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// resetNotionPage resets the given Notion page to the given title and removes
// all its children.
func resetNotionPage(ctx context.Context, client *notionapi.Client, pageID, pageTitle string) error {
	blocks, err := listPageBlocks(ctx, client, pageID)
	if err != nil {
		return errors.Wrap(err, "failed to list page blocks")
	}
	if err := deleteBlocks(ctx, client, blocks); err != nil {
		return errors.Wrap(err, "failed to delete blocks")
	}
	if err := setPageTitle(ctx, client, pageID, pageTitle); err != nil {
		return errors.Wrap(err, "failed to set page title")
	}
	return nil
}

func listPageBlocks(ctx context.Context, client *notionapi.Client, pageID string) (notionapi.Blocks, error) {
	var blocks notionapi.Blocks
	var cursor notionapi.Cursor
	for {
		resp, err := client.Block.GetChildren(ctx, notionapi.BlockID(pageID), &notionapi.Pagination{
			StartCursor: cursor,
			PageSize:    100,
		})
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, resp.Results...)

		if !resp.HasMore {
			break
		}
		cursor = notionapi.Cursor(resp.NextCursor)
	}
	return blocks, nil
}

func deleteBlocks(ctx context.Context, client *notionapi.Client, blocks notionapi.Blocks) error {
	for _, block := range blocks {
		_, err := client.Block.Delete(ctx, block.GetID())
		if err != nil {
			return err
		}
	}
	return nil
}

func setPageTitle(ctx context.Context, client *notionapi.Client, pageID string, title string) error {
	if _, err := client.Page.Update(ctx,
		notionapi.PageID(pageID),
		&notionapi.PageUpdateRequest{
			Properties: notionapi.Properties{
				"Title": notionapi.TitleProperty{
					Title: []notionapi.RichText{
						{Text: &notionapi.Text{Content: title}},
					},
				},
			},
		}); err != nil {
		return err
	}
	return nil
}
