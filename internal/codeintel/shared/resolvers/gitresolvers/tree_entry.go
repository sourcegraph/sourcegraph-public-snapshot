package gitresolvers

import (
	"context"
	"fmt"
	"io"
	stdpath "path"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type treeEntryResolver struct {
	commit    resolvers.GitCommitResolver
	path      string
	isDir     bool
	uriSuffix string

	gitserverClient gitserver.Client
}

func NewGitTreeEntryResolver(commit resolvers.GitCommitResolver, path string, isDir bool, gitserverClient gitserver.Client) resolvers.GitTreeEntryResolver {
	uriSuffix := ""
	if stdpath.Clean("/"+path) != "/" {
		blobOrTree := "blob"
		if isDir {
			blobOrTree = "tree"
		}

		uriSuffix = fmt.Sprintf("/-/%s/%s", blobOrTree, path)
	}

	return &treeEntryResolver{
		commit:          commit,
		path:            path,
		isDir:           isDir,
		uriSuffix:       uriSuffix,
		gitserverClient: gitserverClient,
	}
}

func (r *treeEntryResolver) Repository() resolvers.RepositoryResolver     { return r.commit.Repository() }
func (r *treeEntryResolver) Commit() resolvers.GitCommitResolver          { return r.commit }
func (r *treeEntryResolver) Path() string                                 { return r.path }
func (r *treeEntryResolver) Name() string                                 { return stdpath.Base(r.path) }
func (r *treeEntryResolver) URL() string                                  { return r.commit.URI() + r.uriSuffix }
func (r *treeEntryResolver) RecordID() string                             { return r.path }
func (r *treeEntryResolver) ToGitTree() (resolvers.GitBlobResolver, bool) { return r, r.isDir }
func (r *treeEntryResolver) ToGitBlob() (resolvers.GitTreeResolver, bool) { return r, !r.isDir }

func (r *treeEntryResolver) Content(ctx context.Context, args *resolvers.GitTreeContentPageArgs) (string, error) {
	content, err := GetFullContents(ctx,
		r.gitserverClient,
		api.RepoName(r.Repository().Name()),
		api.CommitID(r.commit.OID()),
		core.NewRepoRelPathUnchecked(r.path))
	if err != nil {
		return "", err
	}

	return joinSelection(strings.Split(string(content), "\n"), args.StartLine, args.EndLine), nil
}

func GetFullContents(ctx context.Context, gitserverClient gitserver.Client, repoName api.RepoName, commit api.CommitID, path core.RepoRelPath) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	fr, err := gitserverClient.NewFileReader(ctx, repoName, commit, path.RawValue())
	if err != nil {
		return nil, err
	}
	defer fr.Close()
	content, err := io.ReadAll(fr)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func joinSelection(lines []string, startLine, endLine *int32) string {
	// Trim from back
	if endLine != nil && *endLine <= int32(len(lines)) {
		lines = lines[:*endLine]
	}

	// Trim from front
	if startLine != nil && *startLine >= 0 {
		lines = lines[*startLine:]
	}

	// Collapse remaining lines
	return strings.Join(lines, "\n")
}
