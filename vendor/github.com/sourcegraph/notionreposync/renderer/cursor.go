package renderer

import (
	"encoding/json"
	"fmt"

	"github.com/jomei/notionapi"
	"github.com/yuin/goldmark/ast"
)

type cursor struct {
	rootBlocks []notionapi.Block
	m          map[ast.Node]notionapi.Block
	cur        ast.Node
}

func (c *cursor) RichText() *notionapi.RichText {
	switch block := c.m[c.cur].(type) {
	case *notionapi.ParagraphBlock:
		return &block.Paragraph.RichText[len(block.Paragraph.RichText)-1]
	case *notionapi.BulletedListItemBlock:
		return &block.BulletedListItem.RichText[len(block.BulletedListItem.RichText)-1]
	case *notionapi.NumberedListItemBlock:
		return &block.NumberedListItem.RichText[len(block.NumberedListItem.RichText)-1]
	case *notionapi.Heading1Block:
		return &block.Heading1.RichText[len(block.Heading1.RichText)-1]
	case *notionapi.Heading2Block:
		return &block.Heading2.RichText[len(block.Heading2.RichText)-1]
	case *notionapi.Heading3Block:
		return &block.Heading3.RichText[len(block.Heading3.RichText)-1]
	case *notionapi.QuoteBlock:
		return &block.Quote.RichText[len(block.Quote.RichText)-1]
	case *notionapi.CalloutBlock:
		return &block.Callout.RichText[len(block.Callout.RichText)-1]
	case *notionapi.TableRowBlock:
		return &block.TableRow.Cells[len(block.TableRow.Cells)-1][len(block.TableRow.Cells[len(block.TableRow.Cells)-1])-1]
	default:
		panic(errUnknownBlock(block, nil))
	}
}

func (c *cursor) Block() notionapi.Block {
	return c.m[c.cur]
}

func (c *cursor) AppendRichText(rt *notionapi.RichText) {
	rts := []notionapi.RichText{*rt}

	// See https://developers.notion.com/reference/request-limits#limits-for-property-values
	if len(rt.Text.Content) > MaxRichTextContentLength {
		rts = []notionapi.RichText{}
		chunks := chunkText(rt.Text.Content)
		for _, chunk := range chunks {
			chunkRT := *rt
			chunkRT.Text.Content = chunk
			rts = append(rts, chunkRT)
		}
	}

	switch block := c.m[c.cur].(type) {
	case *notionapi.ParagraphBlock:
		block.Paragraph.RichText = append(block.Paragraph.RichText, rts...)
	case *notionapi.BulletedListItemBlock:
		block.BulletedListItem.RichText = append(block.BulletedListItem.RichText, rts...)
	case *notionapi.NumberedListItemBlock:
		block.NumberedListItem.RichText = append(block.NumberedListItem.RichText, rts...)
	case *notionapi.Heading1Block:
		block.Heading1.RichText = append(block.Heading1.RichText, rts...)
	case *notionapi.Heading2Block:
		block.Heading2.RichText = append(block.Heading2.RichText, rts...)
	case *notionapi.Heading3Block:
		block.Heading3.RichText = append(block.Heading3.RichText, rts...)
	case *notionapi.QuoteBlock:
		block.Quote.RichText = append(block.Quote.RichText, rts...)
	case *notionapi.CalloutBlock:
		block.Callout.RichText = append(block.Callout.RichText, rts...)
	case *notionapi.TableRowBlock:
		block.TableRow.Cells[len(block.TableRow.Cells)-1] = append(block.TableRow.Cells[len(block.TableRow.Cells)-1], rts...)
	default:
		panic(errUnknownBlock(block, rts))
	}
}

func (c *cursor) AppendBlock(b notionapi.Block) {
	if c.cur.Kind() == ast.KindDocument {
		c.rootBlocks = append(c.rootBlocks, b)
	} else {
		switch block := c.Block().(type) {
		case *notionapi.ParagraphBlock:
			block.Paragraph.Children = append(block.Paragraph.Children, b)
		case *notionapi.BulletedListItemBlock:
			block.BulletedListItem.Children = append(block.BulletedListItem.Children, b)
		case *notionapi.NumberedListItemBlock:
			block.NumberedListItem.Children = append(block.NumberedListItem.Children, b)
		case *notionapi.Heading1Block:
			block.Heading1.Children = append(block.Heading1.Children, b)
		case *notionapi.Heading2Block:
			block.Heading2.Children = append(block.Heading2.Children, b)
		case *notionapi.Heading3Block:
			block.Heading3.Children = append(block.Heading3.Children, b)
		case *notionapi.QuoteBlock:
			block.Quote.Children = append(block.Quote.Children, b)
		case *notionapi.CalloutBlock:
			block.Callout.Children = append(block.Callout.Children, b)
		case *notionapi.TableBlock:
			block.Table.Children = append(block.Table.Children, b)
		default:
			panic(errUnknownBlock(block, b))
		}
	}
}

func (c *cursor) Set(node ast.Node, block notionapi.Block) {
	c.m[node] = block
}

func (c *cursor) Descend(node ast.Node) {
	c.cur = node
}

func (c *cursor) Ascend() {
	for {
		if c.cur.Parent() != nil {
			c.cur = c.cur.Parent()
			if c.m[c.cur] != nil {
				return
			}
		} else {
			return
		}
	}
}

func errUnknownBlock(block notionapi.Block, data any) error {
	raw, _ := json.Marshal(data)
	return fmt.Errorf("unknown block type: %T (%s)", block, string(raw))
}
