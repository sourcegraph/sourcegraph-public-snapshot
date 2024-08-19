package renderer

import (
	"context"

	"github.com/jomei/notionapi"
)

// MaxBlocksPerUpdate is the maximum number of blocks that can be added in a single request.
//
// See https://developers.notion.com/reference/patch-block-children
const MaxBlocksPerUpdate = 100

// MaxNestedBlockLevelsPerUpdate is the maximum number of nesting levels that can be added in a
// single request.
//
// See https://developers.notion.com/reference/patch-block-children
const MaxNestedBlockLevelsPerUpdate = 2

// MaxRichTextContentLength is the maximum length of a single rich text content object.
// See https://developers.notion.com/reference/request-limits#limits-for-property-values
const MaxRichTextContentLength = 2000

// BlockUpdater implements the desired handling for Notion blocks converted from
// Markdown. It should represent a single parent block, to which all children
// are added.
type BlockUpdater interface {
	// AddChildren should add the given children to the desired parent block.
	//
	// The caller calls it while respecting MaxBlocksPerUpdate and
	// MaxRichTextContentLength - implementations can assume the set of children
	// being added is of a reasonable size and adhere's to Notion's API limits.
	AddChildren(ctx context.Context, children []notionapi.Block) error
}
