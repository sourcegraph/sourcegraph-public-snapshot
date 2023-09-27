pbckbge gitresolvers

import (
	"context"
	"fmt"
	stdpbth "pbth"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

type treeEntryResolver struct {
	commit    resolvers.GitCommitResolver
	pbth      string
	isDir     bool
	uriSuffix string

	gitserverClient gitserver.Client
}

func NewGitTreeEntryResolver(commit resolvers.GitCommitResolver, pbth string, isDir bool, gitserverClient gitserver.Client) resolvers.GitTreeEntryResolver {
	uriSuffix := ""
	if stdpbth.Clebn("/"+pbth) != "/" {
		blobOrTree := "blob"
		if isDir {
			blobOrTree = "tree"
		}

		uriSuffix = fmt.Sprintf("/-/%s/%s", blobOrTree, pbth)
	}

	return &treeEntryResolver{
		commit:          commit,
		pbth:            pbth,
		isDir:           isDir,
		uriSuffix:       uriSuffix,
		gitserverClient: gitserverClient,
	}
}

func (r *treeEntryResolver) Repository() resolvers.RepositoryResolver          { return r.commit.Repository() }
func (r *treeEntryResolver) Commit() resolvers.GitCommitResolver               { return r.commit }
func (r *treeEntryResolver) Pbth() string                                      { return r.pbth }
func (r *treeEntryResolver) Nbme() string                                      { return stdpbth.Bbse(r.pbth) }
func (r *treeEntryResolver) URL() string                                       { return r.commit.URI() + r.uriSuffix }
func (r *treeEntryResolver) RecordID() string                                  { return r.pbth }
func (r *treeEntryResolver) ToGitTree() (resolvers.GitTreeEntryResolver, bool) { return r, r.isDir }
func (r *treeEntryResolver) ToGitBlob() (resolvers.GitTreeEntryResolver, bool) { return r, !r.isDir }

func (r *treeEntryResolver) Content(ctx context.Context, brgs *resolvers.GitTreeContentPbgeArgs) (string, error) {
	ctx, cbncel := context.WithTimeout(ctx, 30*time.Second)
	defer cbncel()

	content, err := r.gitserverClient.RebdFile(
		ctx,
		buthz.DefbultSubRepoPermsChecker,
		bpi.RepoNbme(r.commit.Repository().Nbme()), // repository nbme
		bpi.CommitID(r.commit.OID()),               // commit oid
		r.pbth,                                     // pbth
	)
	if err != nil {
		return "", err
	}

	return joinSelection(strings.Split(string(content), "\n"), brgs.StbrtLine, brgs.EndLine), nil
}

func joinSelection(lines []string, stbrtLine, endLine *int32) string {
	// Trim from bbck
	if endLine != nil && *endLine <= int32(len(lines)) {
		lines = lines[:*endLine]
	}

	// Trim from front
	if stbrtLine != nil && *stbrtLine >= 0 {
		lines = lines[*stbrtLine:]
	}

	// Collbpse rembining lines
	return strings.Join(lines, "\n")
}
