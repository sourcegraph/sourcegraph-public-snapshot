package graphqlbackend

import "context"

// DocSitePageResolver is the resolver for the GraphQL field Query.docSitePage.
//
// It is set at init time.
var DocSitePageResolver func(context.Context, DocSitePageArgs) (DocSitePage, error)

type DocSitePageArgs struct {
	Path string
}

// DocSitePage is the interface for the GraphQL type DocSitePage.
type DocSitePage interface {
	Title() string
	ContentHTML() string
	IndexHTML() string
	FilePath() string
}

func (*schemaResolver) DocSitePage(ctx context.Context, args *DocSitePageArgs) (DocSitePage, error) {
	return DocSitePageResolver(ctx, *args)
}
