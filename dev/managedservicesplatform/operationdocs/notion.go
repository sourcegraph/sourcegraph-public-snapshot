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
		// TODO Notion rejects local anchor links, we need to figure something
		// out later
		return "https://www.notion.so/help/create-links-and-backlinks", nil
	}
	return link, nil
}

func NewNotionConverter(ctx context.Context, blocks renderer.BlockUpdater) notionmarkdown.Processor {
	return notionmarkdown.NewProcessor(ctx, blocks,
		renderer.WithLinkResolver(anchorLinkResolver{}))
}
