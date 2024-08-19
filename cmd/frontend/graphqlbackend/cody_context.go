package graphqlbackend

import (
	"bytes"
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/conc/iter"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/codycontext"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type CodyContextResolver interface {
	ChatIntent(ctx context.Context, args ChatIntentArgs) (IntentResolver, error)
	ChatContext(ctx context.Context, args ChatContextArgs) (ChatContextResolver, error)
	RankContext(ctx context.Context, args RankContextArgs) (RankContextResolver, error)
	RecordContext(ctx context.Context, args RecordContextArgs) (*EmptyResponse, error)
	UrlMentionContext(ctx context.Context, args UrlMentionContextArgs) (UrlMentionContextResolver, error)
	// GetCodyContext is the existing Cody Enterprise context endpoint
	GetCodyContext(ctx context.Context, args GetContextArgs) ([]ContextResultResolver, error)
	GetCodyContextAlternatives(ctx context.Context, args GetContextArgs) (*ContextAlternativesResolver, error)
}

type GetContextArgs struct {
	Repos            []graphql.ID
	FilePatterns     *[]string
	Query            string
	CodeResultsCount int32
	TextResultsCount int32
}

type UrlMentionContextArgs struct {
	Url string
}

type UrlMentionContextResolver interface {
	Title() *string
	Content() string
}

func NewContextAlternativesResolver(db database.DB, gitserverClient gitserver.Client, contextAlternatives *codycontext.GetCodyContextResult) *ContextAlternativesResolver {
	return &ContextAlternativesResolver{
		db:                  db,
		gitserverClient:     gitserverClient,
		ContextAlternatives: contextAlternatives,
	}
}

type ContextAlternativesResolver struct {
	db                  database.DB
	gitserverClient     gitserver.Client
	ContextAlternatives *codycontext.GetCodyContextResult
}

func (c *ContextAlternativesResolver) ContextLists() []*ContextListResolver {
	var res []*ContextListResolver
	for _, contextList := range c.ContextAlternatives.ContextLists {
		res = append(res, &ContextListResolver{ContextList: &contextList, db: c.db, gitserverClient: c.gitserverClient})
	}
	return res
}

type ContextListResolver struct {
	db              database.DB
	gitserverClient gitserver.Client
	ContextList     *codycontext.ContextList
}

func (r *ContextListResolver) ContextItems(ctx context.Context) (_ []ContextResultResolver, err error) {
	tr, ctx := trace.New(ctx, "resolveChunks")
	defer tr.EndWithErr(&err)

	return iter.MapErr(r.ContextList.FileChunks, func(fileChunk *codycontext.FileChunkContext) (ContextResultResolver, error) {
		return r.fileChunkToResolver(ctx, fileChunk)
	})
}

func (r *ContextListResolver) Name(ctx context.Context) (string, error) {
	return r.ContextList.Name, nil
}

// The rough size of a file chunk in runes. The value 1024 is due to historical reasons -- Cody context was once based
// on embeddings, and we chunked files into ~1024 characters (aiming for 256 tokens, assuming each token takes 4
// characters on average).
//
// Ideally, the caller would pass a token 'budget' and we'd use a tokenizer and attempt to exactly match this budget.
const chunkSizeRunes = 1024

func (r *ContextListResolver) fileChunkToResolver(ctx context.Context, chunk *codycontext.FileChunkContext) (ContextResultResolver, error) {
	repoResolver := NewMinimalRepositoryResolver(r.db, r.gitserverClient, chunk.RepoID, chunk.RepoName)

	commitResolver := NewGitCommitResolver(r.db, r.gitserverClient, repoResolver, chunk.CommitID, nil)
	stat, err := r.gitserverClient.Stat(ctx, chunk.RepoName, chunk.CommitID, chunk.Path)
	if err != nil {
		return nil, err
	}

	gitTreeEntryResolver := NewGitTreeEntryResolver(r.db, r.gitserverClient, GitTreeEntryResolverOpts{
		Commit: commitResolver,
		Stat:   stat,
	})

	// Populate content ahead of time so we can do it concurrently
	content, err := gitTreeEntryResolver.Content(ctx, &GitTreeContentPageArgs{
		StartLine: pointers.Ptr(int32(chunk.StartLine)),
	})
	if err != nil {
		return nil, err
	}

	numLines := countLines(content, chunkSizeRunes)
	endLine := chunk.StartLine + numLines - 1 // subtract 1 because endLine is inclusive
	return NewFileChunkContextResolver(gitTreeEntryResolver, chunk.StartLine, endLine), nil
}

// countLines finds the number of lines corresponding to the number of runes. We 'round down'
// to ensure that we don't return more characters than our budget.
func countLines(content string, numRunes int) int {
	if len(content) == 0 {
		return 0
	}

	if content[len(content)-1] != '\n' {
		content += "\n"
	}

	runes := []rune(content)
	truncated := runes[:min(len(runes), numRunes)]
	in := []byte(string(truncated))
	return bytes.Count(in, []byte("\n"))
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
	return f.treeEntry.Content(ctx, &GitTreeContentPageArgs{
		StartLine: &f.startLine,
		EndLine:   &f.endLine,
	})
}

type ChatIntentArgs struct {
	Query         string
	InteractionID string
}

type ChatContextArgs struct {
	Query         string
	InteractionID string
	Repo          string
	ResultsCount  *int32
}

type RankContextArgs struct {
	Query                     string
	ContextItems              []InputContextItem
	RankOptions               *RankOptions
	TargetModel               *string
	TargetContextWindowTokens *int32
	Intent                    *string
	Command                   *string
	InteractionID             string
}

type RecordContextArgs struct {
	InteractionID       string
	UsedContextItems    []InputContextItem
	IgnoredContextItems []InputContextItem
}

type InputContextItem struct {
	Content   string
	Retriever string
	Score     *float64
	FileName  *string
	StartLine *int32
	EndLine   *int32
}

type RankOptions struct {
	Ranker string
}

type IntentResolver interface {
	Intent() string
	Score() float64
	SearchScore() float64
	EditScore() float64
}

type RankContextResolver interface {
	Ranker() string
	Used() []RankedItemResolver
	Ignored() []RankedItemResolver
}

type ChatContextResolver interface {
	ContextItems() []RetrieverContextItemResolver
	PartialErrors() []string
	StopReason() string
}

type RetrieverContextItemResolver interface {
	Item() ContextResultResolver
	Score() *float64
	Retriever() string
}

type RankedItemResolver interface {
	Index() int32
	Score() float64
}
