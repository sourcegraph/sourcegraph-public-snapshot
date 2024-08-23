package markdown

import (
	"context"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"

	"github.com/sourcegraph/notionreposync/renderer"
)

type Processor struct{ md goldmark.Markdown }

// NewProcessor returns a new simple Markdown processor that can be used to
// process Markdown test, sending the resulting Notion document blocks to the
// given BlockUpdater.
func NewProcessor(ctx context.Context, blocks renderer.BlockUpdater, opts ...renderer.Option) Processor {
	r := renderer.NewNodeRenderer(ctx, blocks, opts...)
	return Processor{
		goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithRenderer(
				// Default renderer priority is 1000 in the docs, but the GFM extensions also
				// injects renderers for the tables, which are set to be 500, and would overrides
				// our custom implementation if we didn't instead set the priority to 200.
				goldmarkrenderer.NewRenderer(goldmarkrenderer.WithNodeRenderers(util.Prioritized(r, 200))),
			),
		),
	}
}

// ProcessMarkdown ingests the given Markdown source, sending converted to
// Notion blocks to the BlockUpdater given in NewProcessor(...)
func (c Processor) ProcessMarkdown(source []byte, opts ...parser.ParseOption) error {
	return c.md.Convert(
		source,
		io.Discard, // no destination - our renderer sends blocks to BlockUpdater
		opts...)
}
