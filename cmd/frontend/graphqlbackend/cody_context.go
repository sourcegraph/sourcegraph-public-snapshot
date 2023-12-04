package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type CodyContextResolver interface {
	GetCodyContext(ctx context.Context, args GetContextArgs) ([]ContextResultResolver, error)
}

type GetContextArgs struct {
	Repos            []graphql.ID
	Query            string
	CodeResultsCount int32
	TextResultsCount int32
}

type ContextResultResolver interface {
	ToFileChunkContext() (*FileChunkContextResolver, bool)
}

func NewFileChunkContextResolver(gitTreeEntryResolver *GitTreeEntryResolver, startLine, endLine int) *FileChunkContextResolver {
	return &FileChunkContextResolver{
		treeEntry: gitTreeEntryResolver,
		startLine: int32(startLine),
		endLine:   int32(endLine),
	}
}

type FileChunkContextResolver struct {
	treeEntry          *GitTreeEntryResolver
	startLine, endLine int32
}

var _ ContextResultResolver = (*FileChunkContextResolver)(nil)

func (f *FileChunkContextResolver) Blob() *GitTreeEntryResolver { return f.treeEntry }
func (f *FileChunkContextResolver) StartLine() int32            { return f.startLine }
func (f *FileChunkContextResolver) EndLine() int32              { return f.endLine }
func (f *FileChunkContextResolver) ToFileChunkContext() (*FileChunkContextResolver, bool) {
	return f, true
}

func (f *FileChunkContextResolver) ChunkContent(ctx context.Context) (string, error) {
	startLine, endLine := int32(f.startLine), int32(f.endLine)
	return f.treeEntry.Content(ctx, &GitTreeContentPageArgs{
		StartLine: &startLine,
		EndLine:   &endLine,
	})
}
