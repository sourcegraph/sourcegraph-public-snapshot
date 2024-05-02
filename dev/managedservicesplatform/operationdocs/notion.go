package operationdocs

import (
	"context"
	"strings"

	notionmarkdown "github.com/sourcegraph/notionreposync/markdown"
	"github.com/sourcegraph/notionreposync/renderer"
)

type linkResolver struct {
	replacements map[string]string
}

func (l linkResolver) ResolveLink(link string) (string, error) {
	if replace, ok := l.replacements[link]; ok {
		return replace, nil
	}
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
func NewNotionConverter(ctx context.Context, blocks renderer.BlockUpdater, replacements map[string]string) notionmarkdown.Processor {
	return notionmarkdown.NewProcessor(ctx, blocks,
		renderer.WithLinkResolver(linkResolver{replacements}))
}
