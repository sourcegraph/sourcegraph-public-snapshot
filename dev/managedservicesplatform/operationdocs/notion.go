package operationdocs

import (
	"context"
	"strings"

	notionmarkdown "github.com/sourcegraph/notionreposync/markdown"
	"github.com/sourcegraph/notionreposync/renderer"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/internal/markdown"
)

type linkResolver struct{}

func (linkResolver) ResolveLink(link string) (string, error) {
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
		renderer.WithLinkResolver(linkResolver{}))
}

func addNotionWarning(md *markdown.Builder) {
	md.Admonitionf(markdown.AdmonitionWarning,
		"Due to Notion limitations, we currently recreate the contents of each page entirely on updates. This means you %s share links to sections within these pages, as the linked blocks will not persist - %s tracks future improvements to this mechanism.",
		markdown.Bold("cannot"), markdown.Link("sourcegraph/notionreposync#7", "https://github.com/sourcegraph/notionreposync/issues/7"))

}
