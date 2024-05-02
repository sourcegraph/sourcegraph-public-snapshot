package operationdocs

import (
	"context"
	"strings"

	notionmarkdown "github.com/sourcegraph/notionreposync/markdown"
	"github.com/sourcegraph/notionreposync/renderer"
)

type anchorLinkResolver struct{}

func (anchorLinkResolver) ResolveLink(link string) (string, error) {
	if strings.HasPrefix(link, "#") {
		// Notion rejects local anchor links, and making links to blocks by ID
		// that we don't know yet is very hard. We can need to figure something
		// out later. For now, strip all anchor links.
		return "", renderer.ErrDiscardLink
	}
	return link, nil
}

// NewNotionConverter creates preconfigured notionmarkdown.Processor for
// operational docs.
func NewNotionConverter(ctx context.Context, blocks renderer.BlockUpdater) notionmarkdown.Processor {
	return notionmarkdown.NewProcessor(ctx, blocks,
		renderer.WithLinkResolver(anchorLinkResolver{}))
}
