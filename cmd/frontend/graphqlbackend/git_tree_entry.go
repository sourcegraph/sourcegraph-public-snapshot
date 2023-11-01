package graphqlbackend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/binary"
	"github.com/sourcegraph/sourcegraph/internal/cloneurls"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitTreeEntryResolver resolves an entry in a Git tree in a repository. The entry can be any Git
// object type that is valid in a tree.
//
// Prefer using the constructor, NewGitTreeEntryResolver.
type GitTreeEntryResolver struct {
	db              database.DB
	gitserverClient gitserver.Client
	commit          *GitCommitResolver

	contentOnce      sync.Once
	fullContentBytes []byte
	contentErr       error
	// stat is this tree entry's file info. Its Name method must return the full path relative to
	// the root, not the basename.
	stat fs.FileInfo
}

type GitTreeEntryResolverOpts struct {
	Commit *GitCommitResolver
	Stat   fs.FileInfo
}

type GitTreeContentPageArgs struct {
	StartLine *int32
	EndLine   *int32
}

func NewGitTreeEntryResolver(db database.DB, gitserverClient gitserver.Client, opts GitTreeEntryResolverOpts) *GitTreeEntryResolver {
	return &GitTreeEntryResolver{
		db:              db,
		commit:          opts.Commit,
		stat:            opts.Stat,
		gitserverClient: gitserverClient,
	}
}

func (r *GitTreeEntryResolver) Path() string { return r.stat.Name() }
func (r *GitTreeEntryResolver) Name() string { return path.Base(r.stat.Name()) }

func (r *GitTreeEntryResolver) ToGitTree() (*GitTreeEntryResolver, bool) { return r, r.IsDirectory() }
func (r *GitTreeEntryResolver) ToGitBlob() (*GitTreeEntryResolver, bool) { return r, !r.IsDirectory() }

func (r *GitTreeEntryResolver) ToVirtualFile() (*VirtualFileResolver, bool) { return nil, false }
func (r *GitTreeEntryResolver) ToBatchSpecWorkspaceFile() (BatchWorkspaceFileResolver, bool) {
	return nil, false
}

func (r *GitTreeEntryResolver) TotalLines(ctx context.Context) (int32, error) {
	// If it is a binary, return 0
	binary, err := r.Binary(ctx)
	if err != nil || binary {
		return 0, err
	}

	// Call content so that r.fullContentBytes is populated.
	_, err = r.Content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return 0, err
	}

	return int32(lineCount(r.fullContentBytes)), nil
}

func (r *GitTreeEntryResolver) ByteSize(ctx context.Context) (int32, error) {
	// We only care about the full content length here, so we just need content to be set.
	_, err := r.Content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return 0, err
	}

	return int32(len(r.fullContentBytes)), nil
}

func (r *GitTreeEntryResolver) Content(ctx context.Context, args *GitTreeContentPageArgs) (string, error) {
	r.contentOnce.Do(func() {
		r.fullContentBytes, r.contentErr = r.gitserverClient.ReadFile(
			ctx,
			r.commit.repoResolver.RepoName(),
			api.CommitID(r.commit.OID()),
			r.Path(),
		)
	})

	return string(pageContent(r.fullContentBytes, int32ToIntPtr(args.StartLine), int32ToIntPtr(args.EndLine))), r.contentErr
}

func int32ToIntPtr(p *int32) *int {
	if p == nil {
		return nil
	}
	val := int(*p)
	return &val
}

// pageContent returns a subslice of content for the range of startLine:endLine.
// If startLine is unset, it is set to 1 (start of file).
// If endLine is unset, it is set to the end of the file.
// endLine must not be before startLine, otherwise the whole content is returned.
// startLine and endLine are 1-indexed!
func pageContent(content []byte, startLine, endLine *int) []byte {
	// Trivial case: No pagination.
	if startLine == nil && endLine == nil {
		return content
	}

	totalLineCount := lineCount(content)

	var (
		startCursor int
		endCursor   int = totalLineCount
	)

	if startLine != nil {
		if *startLine < 1 {
			startCursor = 0
		} else if *startLine > totalLineCount {
			startCursor = totalLineCount
		} else {
			startCursor = *startLine - 1
		}
	}

	if endLine != nil {
		if *endLine > 0 {
			endCursor = *endLine
		}
	}

	if endCursor < startCursor {
		return content
	}

	start := nthIndex(content, '\n', startCursor)
	if start == -1 {
		start = 0
	}
	end := nthIndex(content, '\n', endCursor)
	if end == -1 {
		end = len(content)
	}
	if end < start {
		return content
	}

	return content[start:end]
}

func lineCount(in []byte) int {
	c := bytes.Count(in, []byte("\n"))
	// Final newline doesn't mark a new line.
	if in[len(in)-1] != '\n' {
		return c + 1
	}
	return c
}

func nthIndex(in []byte, sep byte, n int) (idx int) {
	if n < 0 {
		return -1
	}
	if len(in) < 1 {
		return -1
	}

	start := 0
	for i := 0; i < n; i++ {
		idx := bytes.IndexByte(in[start:], sep)
		if idx == -1 {
			return idx
		}
		start = start + idx + 1
	}
	return start
}

func (r *GitTreeEntryResolver) RichHTML(ctx context.Context, args *GitTreeContentPageArgs) (string, error) {
	content, err := r.Content(ctx, args)
	if err != nil {
		return "", err
	}

	return richHTML(content, path.Ext(r.Path()))
}

func (r *GitTreeEntryResolver) Binary(ctx context.Context) (bool, error) {
	// We only care about the full content length here, so we just need r.fullContentLines to be set.
	_, err := r.Content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return false, err
	}

	return binary.IsBinary(r.fullContentBytes), nil
}

func (r *GitTreeEntryResolver) Highlight(ctx context.Context, args *HighlightArgs) (*HighlightedFileResolver, error) {
	// Currently, pagination + highlighting is not supported, throw out an error if it is attempted.
	if (args.StartLine != nil || args.EndLine != nil) && args.Format != "HTML_PLAINTEXT" {
		return nil, errors.New("pagination is not supported with formats other than HTML_PLAINTEXT, don't " +
			"set startLine or endLine with other formats")
	}

	content, err := r.Content(ctx, &GitTreeContentPageArgs{StartLine: args.StartLine, EndLine: args.EndLine})
	if err != nil {
		return nil, err
	}

	return highlightContent(ctx, args, content, r.Path(), highlight.Metadata{
		RepoName: r.commit.repoResolver.Name(),
		Revision: string(r.commit.oid),
	})
}

func (r *GitTreeEntryResolver) Commit() *GitCommitResolver { return r.commit }

func (r *GitTreeEntryResolver) Repository() *RepositoryResolver { return r.commit.repoResolver }

func (r *GitTreeEntryResolver) URL(ctx context.Context) (string, error) {
	return r.url(ctx).String(), nil
}

func (r *GitTreeEntryResolver) url(ctx context.Context) *url.URL {
	submodule := r.Submodule()
	if submodule == nil {
		return r.urlPath(r.commit.repoRevURL())
	}

	tr, ctx := trace.New(ctx, "GitTreeEntryResolver.url", attribute.Bool("submodule", true))
	defer tr.End()
	submoduleURL := submodule.URL()
	if strings.HasPrefix(submoduleURL, "../") {
		submoduleURL = path.Join(r.Repository().Name(), submoduleURL)
	}
	repoName, err := cloneURLToRepoName(ctx, r.db, submoduleURL)
	if err != nil {
		log15.Error("Failed to resolve submodule repository name from clone URL", "cloneURL", submodule.URL(), "err", err)
		return &url.URL{}
	}
	return &url.URL{Path: "/" + repoName + "@" + submodule.Commit()}
}

func (r *GitTreeEntryResolver) CanonicalURL() string {
	canonicalUrl := r.commit.canonicalRepoRevURL()
	return r.urlPath(canonicalUrl).String()
}

func (r *GitTreeEntryResolver) ChangelistURL(ctx context.Context) (*string, error) {
	repo := r.Repository()
	source, err := repo.SourceType(ctx)
	if err != nil {
		return nil, err
	}

	if *source != PerforceDepotSourceType {
		return nil, nil
	}

	cl, err := r.commit.PerforceChangelist(ctx)
	if err != nil {
		return nil, err
	}

	// This is an oddity. We have checked above that this repository is a perforce depot. Then this
	// commit of this blob must also have a changelist ID associated with it.
	//
	// If we ever hit this check, this is a bug and the error should be propagated out.
	if cl == nil {
		return nil, errors.Newf(
			"failed to retrieve changelist from commit %q in repo %q",
			string(r.commit.OID()),
			string(repo.RepoName()),
		)
	}

	u := r.urlPath(cl.cidURL()).String()
	return &u, nil
}

func (r *GitTreeEntryResolver) urlPath(prefix *url.URL) *url.URL {
	// Dereference to copy to avoid mutating the input
	u := *prefix
	if r.IsRoot() {
		return &u
	}

	typ := "blob"
	if r.IsDirectory() {
		typ = "tree"
	}

	u.Path = path.Join(u.Path, "-", typ, r.Path())
	return &u
}

func (r *GitTreeEntryResolver) IsDirectory() bool { return r.stat.Mode().IsDir() }

func (r *GitTreeEntryResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}
	return externallink.FileOrDir(ctx, r.db, r.gitserverClient, repo, r.commit.inputRevOrImmutableRev(), r.Path(), r.stat.Mode().IsDir())
}

func (r *GitTreeEntryResolver) RawZipArchiveURL() string {
	return globals.ExternalURL().ResolveReference(&url.URL{
		Path:     path.Join(r.Repository().URL(), "-/raw/", r.Path()),
		RawQuery: "format=zip",
	}).String()
}

func (r *GitTreeEntryResolver) Submodule() *gitSubmoduleResolver {
	if submoduleInfo, ok := r.stat.Sys().(gitdomain.Submodule); ok {
		return &gitSubmoduleResolver{submodule: submoduleInfo}
	}
	return nil
}

func cloneURLToRepoName(ctx context.Context, db database.DB, cloneURL string) (_ string, err error) {
	tr, ctx := trace.New(ctx, "cloneURLToRepoName")
	defer tr.EndWithErr(&err)

	repoName, err := cloneurls.RepoSourceCloneURLToRepoName(ctx, db, cloneURL)
	if err != nil {
		return "", err
	}
	if repoName == "" {
		return "", errors.Errorf("no matching code host found for %s", cloneURL)
	}
	return string(repoName), nil
}

func CreateFileInfo(path string, isDir bool) fs.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

func (r *GitTreeEntryResolver) IsSingleChild(ctx context.Context) (bool, error) {
	if !r.IsDirectory() {
		return false, nil
	}

	entries, err := r.gitserverClient.ReadDir(ctx, r.commit.repoResolver.RepoName(), api.CommitID(r.commit.OID()), path.Dir(r.Path()), false)
	if err != nil {
		return false, err
	}

	return len(entries) == 1, nil
}

func (r *GitTreeEntryResolver) LSIF(ctx context.Context, args *struct{ ToolName *string }) (resolverstubs.GitBlobLSIFDataResolver, error) {
	var toolName string
	if args.ToolName != nil {
		toolName = *args.ToolName
	}

	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	return EnterpriseResolvers.codeIntelResolver.GitBlobLSIFData(ctx, &resolverstubs.GitBlobLSIFDataArgs{
		Repo:      repo,
		Commit:    api.CommitID(r.Commit().OID()),
		Path:      r.Path(),
		ExactPath: !r.stat.IsDir(),
		ToolName:  toolName,
	})
}

func (r *GitTreeEntryResolver) LocalCodeIntel(ctx context.Context) (*JSONValue, error) {
	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	payload, err := symbols.DefaultClient.LocalCodeIntel(ctx, types.RepoCommitPath{
		Repo:   string(repo.Name),
		Commit: string(r.commit.oid),
		Path:   r.Path(),
	})
	if err != nil {
		return nil, err
	}

	jsonValue, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &JSONValue{Value: string(jsonValue)}, nil
}

func (r *GitTreeEntryResolver) SymbolInfo(ctx context.Context, args *symbolInfoArgs) (*symbolInfoResolver, error) {
	if args == nil {
		return nil, errors.New("expected arguments to symbolInfo")
	}

	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	start := types.RepoCommitPathPoint{
		RepoCommitPath: types.RepoCommitPath{
			Repo:   string(repo.Name),
			Commit: string(r.commit.oid),
			Path:   r.Path(),
		},
		Point: types.Point{
			Row:    int(args.Line),
			Column: int(args.Character),
		},
	}

	fmt.Println("Calling GitTreeEntryResolver.SymbolInfo")
	result, err := symbols.DefaultClient.SymbolInfo(ctx, start)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return &symbolInfoResolver{symbolInfo: result}, nil
}

func (r *GitTreeEntryResolver) LFS(ctx context.Context) (*lfsResolver, error) {
	content, err := r.Content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return nil, err
	}
	return parseLFSPointer(content), nil
}

func (r *GitTreeEntryResolver) Ownership(ctx context.Context, args ListOwnershipArgs) (OwnershipConnectionResolver, error) {
	if _, ok := r.ToGitBlob(); ok {
		return EnterpriseResolvers.ownResolver.GitBlobOwnership(ctx, r, args)
	}
	if _, ok := r.ToGitTree(); ok {
		return EnterpriseResolvers.ownResolver.GitTreeOwnership(ctx, r, args)
	}
	return nil, nil
}

type OwnershipStatsArgs struct {
	Reasons *[]OwnershipReasonType
}

func (r *GitTreeEntryResolver) OwnershipStats(ctx context.Context) (OwnershipStatsResolver, error) {
	if _, ok := r.ToGitTree(); !ok {
		return nil, nil
	}
	return EnterpriseResolvers.ownResolver.GitTreeOwnershipStats(ctx, r)
}

type symbolInfoArgs struct {
	Line      int32
	Character int32
}

type symbolInfoResolver struct{ symbolInfo *types.SymbolInfo }

func (r *symbolInfoResolver) Definition(ctx context.Context) (*symbolLocationResolver, error) {
	return &symbolLocationResolver{location: r.symbolInfo.Definition}, nil
}

func (r *symbolInfoResolver) Hover(ctx context.Context) (*string, error) {
	return r.symbolInfo.Hover, nil
}

type symbolLocationResolver struct {
	location types.RepoCommitPathMaybeRange
}

func (r *symbolLocationResolver) Repo() string   { return r.location.Repo }
func (r *symbolLocationResolver) Commit() string { return r.location.Commit }
func (r *symbolLocationResolver) Path() string   { return r.location.Path }
func (r *symbolLocationResolver) Line() int32 {
	if r.location.Range == nil {
		return 0
	}
	return int32(r.location.Range.Row)
}

func (r *symbolLocationResolver) Character() int32 {
	if r.location.Range == nil {
		return 0
	}
	return int32(r.location.Range.Column)
}

func (r *symbolLocationResolver) Length() int32 {
	if r.location.Range == nil {
		return 0
	}
	return int32(r.location.Range.Length)
}

func (r *symbolLocationResolver) Range() (*lineRangeResolver, error) {
	if r.location.Range == nil {
		return nil, nil
	}
	return &lineRangeResolver{rnge: r.location.Range}, nil
}

type lineRangeResolver struct {
	rnge *types.Range
}

func (r *lineRangeResolver) Line() int32      { return int32(r.rnge.Row) }
func (r *lineRangeResolver) Character() int32 { return int32(r.rnge.Column) }
func (r *lineRangeResolver) Length() int32    { return int32(r.rnge.Length) }

type fileInfo struct {
	path  string
	size  int64
	isDir bool
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return f.size }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() any           { return any(nil) }
