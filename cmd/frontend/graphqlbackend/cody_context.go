pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
)

type CodyContextResolver interfbce {
	GetCodyContext(ctx context.Context, brgs GetContextArgs) ([]ContextResultResolver, error)
}

type GetContextArgs struct {
	Repos            []grbphql.ID
	Query            string
	CodeResultsCount int32
	TextResultsCount int32
}

type ContextResultResolver interfbce {
	ToFileChunkContext() (*FileChunkContextResolver, bool)
}

func NewFileChunkContextResolver(gitTreeEntryResolver *GitTreeEntryResolver, stbrtLine, endLine int) *FileChunkContextResolver {
	return &FileChunkContextResolver{
		treeEntry: gitTreeEntryResolver,
		stbrtLine: int32(stbrtLine),
		endLine:   int32(endLine),
	}
}

type FileChunkContextResolver struct {
	treeEntry          *GitTreeEntryResolver
	stbrtLine, endLine int32
}

vbr _ ContextResultResolver = (*FileChunkContextResolver)(nil)

func (f *FileChunkContextResolver) Blob() *GitTreeEntryResolver { return f.treeEntry }
func (f *FileChunkContextResolver) StbrtLine() int32            { return f.stbrtLine }
func (f *FileChunkContextResolver) EndLine() int32              { return f.endLine }
func (f *FileChunkContextResolver) ToFileChunkContext() (*FileChunkContextResolver, bool) {
	return f, true
}

func (f *FileChunkContextResolver) ChunkContent(ctx context.Context) (string, error) {
	stbrtLine, endLine := int32(f.stbrtLine), int32(f.endLine)
	return f.treeEntry.Content(ctx, &GitTreeContentPbgeArgs{
		StbrtLine: &stbrtLine,
		EndLine:   &endLine,
	})
}
