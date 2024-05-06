package msp

import (
	"context"

	"github.com/jomei/notionapi"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// resetNotionPage resets the given Notion page to the given title and removes
// all its children. For now, we reset the entire page and start from scratch
// each time. This will break Notion block links but it can't be helped, Notion
// is hard to work with.
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
	var pages int
	for {
		resp, err := client.Block.GetChildren(ctx, notionapi.BlockID(pageID), &notionapi.Pagination{
			StartCursor: cursor,
			PageSize:    100,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "page %d: failed to get children", pages)
		}
		for _, b := range resp.Results {
			// Don't treat child pages as blocks on this page, they are different
			// pages.
			if b.GetType() != notionapi.BlockTypeChildPage {
				blocks = append(blocks, b)
			}
		}

		if !resp.HasMore {
			break
		}
		pages += 1
		cursor = notionapi.Cursor(resp.NextCursor)
	}
	return blocks, nil
}

func deleteBlocks(ctx context.Context, client *notionapi.Client, blocks notionapi.Blocks) error {
	// WARNING: this cannot be paralellized, the Notion API will complain about
	// a page-save-conflict.
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
				"title": notionapi.TitleProperty{
					Type: notionapi.PropertyTypeTitle,
					Title: []notionapi.RichText{{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{Content: title},
					}},
				},
			},
		}); err != nil {
		return errors.Wrap(err, "failed to set page title")
	}
	if _, err := client.Block.AppendChildren(ctx, notionapi.BlockID(pageID), &notionapi.AppendBlockChildrenRequest{
		Children: []notionapi.Block{
			notionapi.DividerBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeDivider,
				},
			},
			notionapi.TableOfContentsBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeTableOfContents,
				},
			},
			notionapi.DividerBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeDivider,
				},
			},
		},
	}); err != nil {
		return errors.Wrap(err, "failed to add table of contents block")
	}
	return nil
}
