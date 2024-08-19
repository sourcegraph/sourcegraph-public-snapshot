package notion

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jomei/notionapi"
	"github.com/sourcegraph/notionreposync/renderer"
)

type PageBlockUpdater struct {
	client *notionapi.Client
	pageID string
}

var _ renderer.BlockUpdater = (*PageBlockUpdater)(nil)

// NewPageBlockUpdater creates a new BlockUpdater for the given Notion page, which
// adds all children to the given pageID, to be provided to 'notionreposync/renderer'.
func NewPageBlockUpdater(client *notionapi.Client, pageID string) *PageBlockUpdater {
	return &PageBlockUpdater{
		client: client,
		pageID: pageID,
	}
}

// blockWithChild wraps a notionapi.Block, allowing to grab
// children without having to mess with the underlying types.
type blockWithChild struct {
	b notionapi.Block
}

// Children returns a pointer to the slice of children from the block it wraps.
// The pointer allows to possibly detach the children in the eventuality of
// splitting API calls.
func (b *blockWithChild) Children() *notionapi.Blocks {
	switch bl := b.b.(type) {
	case *notionapi.ParagraphBlock:
		return &bl.Paragraph.Children
	case *notionapi.Heading1Block:
		return &bl.Heading1.Children
	case *notionapi.Heading2Block:
		return &bl.Heading2.Children
	case *notionapi.Heading3Block:
		return &bl.Heading3.Children
	case *notionapi.CalloutBlock:
		return &bl.Callout.Children
	case *notionapi.QuoteBlock:
		return &bl.Quote.Children
	case *notionapi.BulletedListItemBlock:
		return &bl.BulletedListItem.Children
	case *notionapi.NumberedListItemBlock:
		return &bl.NumberedListItem.Children
	case *notionapi.CodeBlock:
		// Code blocks do not have children.
		return nil
	case *notionapi.TableBlock:
		return &bl.Table.Children
	default:
		panic(errUnknownBlock(b.b, nil))
	}
}

func errUnknownBlock(block notionapi.Block, data any) error {
	raw, _ := json.Marshal(data)
	return fmt.Errorf("unknown block type: %T (%s)", block, string(raw))
}

type blockBatch struct {
	tooDeep bool
	blocks  []notionapi.Block
}

// batchBlocks splits the given blocks into batches of blocks. If a block is too deep, it's
// added alone is its own batch and marked as exceeding depth limits.
//
// For example, the tree below below will be split into three batches:
// (a b (c (c1 (c2 (c3)))) d)
// would be batched as [[a b] [(c (c1 (c2 (c3))))] [d])
func batchBlocks(children []notionapi.Block) []blockBatch {
	batches := []blockBatch{}
	acc := blockBatch{}
	shouldFlush := false

	for _, child := range children {
		if getDepth(child) < renderer.MaxNestedBlockLevelsPerUpdate {
			acc.blocks = append(acc.blocks, child)
			shouldFlush = true
		} else {
			batches = append(batches, acc, blockBatch{tooDeep: true, blocks: []notionapi.Block{child}})
			shouldFlush = false
		}
	}

	// Ensure the last batch is included.
	if shouldFlush {
		batches = append(batches, acc)
	}

	return batches
}

// AddChildren appends blocks to the page root.
func (b *PageBlockUpdater) AddChildren(ctx context.Context, children []notionapi.Block) error {
	// As documented in renderer.BlockUpdater, we can trust that the given
	// children adheres to Notion API requirements in terms of block length, but as for the depth limits,
	// we need to perform API requests before pursuing, because we can't append a children to a block without
	// its ID before.

	batches := batchBlocks(children)
	for _, batch := range batches {
		if !batch.tooDeep {
			_, err := b.client.Block.AppendChildren(ctx, notionapi.BlockID(b.pageID), &notionapi.AppendBlockChildrenRequest{
				Children: batch.blocks,
			})
			if err != nil {
				return err
			}
		} else {
			// If we're dealing with a block that's too deep, we append the blocks one by one.
			//
			// It's tempting to instead split the tree into subtrees that are not individually
			// too deep to avoid going block by block, but the API responses will truncate children
			// as well, forcing us to make another request to get the children where we should
			// append the next subtree.
			//
			// With the example from above, assuming we append [a b c] and then attempt at
			// appending (c1 (c2 (c3))) the response might just include (c1 (c2)) while still
			// telling us that c2 has a child, to tell us that we need to make another API call
			// to get it if we wished to.
			//
			// So if we add children of c3, we would have to make that API call, which makes it
			// even more complicated.
			//
			// This is perfectly doable, it's not rocket science, but given the only case where
			// we get deeply nested blocks when importing Markdown is with lists, it's merely
			// an optimization that we can deal later with if we really want to speed things up.
			for _, child := range batch.blocks {
				if err := b.walkAppend(ctx, notionapi.BlockID(b.pageID), child); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// walkAppend recursively appends the given block and all its children to the given parent, making
// one request per block. This allows us to avoid having to deal with tree juggling in the case
// we try to append a block that's too deep.
//
// As the only time this happens is when dealing with nested lists, it's a pretty rare occurence
// and not worth the effort to save a few seconds of API calls, at least not at this stage.
func (b *PageBlockUpdater) walkAppend(ctx context.Context, parentID notionapi.BlockID, block notionapi.Block) error {
	cb := blockWithChild{block}
	children := cb.Children()
	// If there are children and if the current block is *not* a table, we go block by block.
	// Tables are a special case, because you can't append a table without its children.
	if children != nil && len(*children) > 0 && block.GetType() != notionapi.BlockTypeTableBlock {
		detachedChildren := *children
		*children = nil
		resp, err := b.client.Block.AppendChildren(ctx, parentID, &notionapi.AppendBlockChildrenRequest{
			Children: []notionapi.Block{block},
		})
		if err != nil {
			return err
		}

		if len(resp.Results) < 1 {
			return fmt.Errorf("expected at a block in API response, got none")
		}

		blockID := resp.Results[0].GetID()
		for _, child := range detachedChildren {
			if err := b.walkAppend(ctx, blockID, child); err != nil {
				return err
			}
		}
	} else {
		_, err := b.client.Block.AppendChildren(ctx, parentID, &notionapi.AppendBlockChildrenRequest{
			Children: []notionapi.Block{block},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func getDepth(block notionapi.Block) int {
	depth := 0
	for {
		// While the block interface has a GetHasChildren method, it's merely a getter for the HasChildren field,
		// which we do not fill while rendering a markdown document, because it's convoluted to know in advance
		// if a block has children or not by just looking at the AST.
		//
		// https://developers.notion.com/reference/patch-block-children indicates that the API ignores that field
		// if we were to

		switch b := block.(type) {
		case notionapi.ParagraphBlock:
			if len(b.Paragraph.Children) > 0 {
				block = b.Paragraph.Children[len(b.Paragraph.Children)-1]
			} else {
				return depth
			}
		case *notionapi.ParagraphBlock:
			if len(b.Paragraph.Children) > 0 {
				block = b.Paragraph.Children[len(b.Paragraph.Children)-1]
			} else {
				return depth
			}
		case notionapi.Heading1Block:
			if len(b.Heading1.Children) > 0 {
				block = b.Heading1.Children[len(b.Heading1.Children)-1]
			} else {
				return depth
			}
		case *notionapi.Heading1Block:
			if len(b.Heading1.Children) > 0 {
				block = b.Heading1.Children[len(b.Heading1.Children)-1]
			} else {
				return depth
			}
		case notionapi.Heading2Block:
			if len(b.Heading2.Children) > 0 {
				block = b.Heading2.Children[len(b.Heading2.Children)-1]
			} else {
				return depth
			}
		case *notionapi.Heading2Block:
			if len(b.Heading2.Children) > 0 {
				block = b.Heading2.Children[len(b.Heading2.Children)-1]
			} else {
				return depth
			}
		case notionapi.Heading3Block:
			if len(b.Heading3.Children) > 0 {
				block = b.Heading3.Children[len(b.Heading3.Children)-1]
			} else {
				return depth
			}
		case *notionapi.Heading3Block:
			if len(b.Heading3.Children) > 0 {
				block = b.Heading3.Children[len(b.Heading3.Children)-1]
			} else {
				return depth
			}
		case notionapi.CalloutBlock:
			if len(b.Callout.Children) > 0 {
				block = b.Callout.Children[len(b.Callout.Children)-1]
			} else {
				return depth
			}
		case *notionapi.CalloutBlock:
			if len(b.Callout.Children) > 0 {
				block = b.Callout.Children[len(b.Callout.Children)-1]
			} else {
				return depth
			}
		case notionapi.QuoteBlock:
			if len(b.Quote.Children) > 0 {
				block = b.Quote.Children[len(b.Quote.Children)-1]
			} else {
				return depth
			}
		case *notionapi.QuoteBlock:
			if len(b.Quote.Children) > 0 {
				block = b.Quote.Children[len(b.Quote.Children)-1]
			} else {
				return depth
			}
		case notionapi.BulletedListItemBlock:
			if len(b.BulletedListItem.Children) > 0 {
				block = b.BulletedListItem.Children[len(b.BulletedListItem.Children)-1]
			} else {
				return depth
			}
		case *notionapi.BulletedListItemBlock:
			if len(b.BulletedListItem.Children) > 0 {
				block = b.BulletedListItem.Children[len(b.BulletedListItem.Children)-1]
			} else {
				return depth
			}
		case notionapi.NumberedListItemBlock:
			if len(b.NumberedListItem.Children) > 0 {
				block = b.NumberedListItem.Children[len(b.NumberedListItem.Children)-1]
			} else {
				return depth
			}
		case *notionapi.NumberedListItemBlock:
			if len(b.NumberedListItem.Children) > 0 {
				block = b.NumberedListItem.Children[len(b.NumberedListItem.Children)-1]
			} else {
				return depth
			}
		case notionapi.TableBlock:
			if len(b.Table.Children) > 0 {
				block = b.Table.Children[len(b.Table.Children)-1]
			} else {
				return depth
			}
		case *notionapi.TableBlock:
			if len(b.Table.Children) > 0 {
				block = b.Table.Children[len(b.Table.Children)-1]
			} else {
				return depth
			}
		default:
			return depth
		}

		depth++
	}
}
