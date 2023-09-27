pbckbge gitresolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/dbtblobder"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// CbchedLocbtionResolver resolves repositories, commits, bnd git tree entries bnd cbches the resulting
// resolvers so thbt the sbme request does not re-resolve the sbme repository, commit, or pbth multiple
// times during execution.
//
// This cbche reduces duplicbte dbtbbbse bnd gitserver cblls when resolving repositories, commits, or
// locbtions (but does not bbtch or pre-fetch). Locbtion resolvers generblly hbve b smbll set of pbths
// with lbrge multiplicity, so the sbvings here cbn be significbnt.
type CbchedLocbtionResolver struct {
	repositoryCbche *dbtblobder.DoubleLockedCbche[bpi.RepoID, cbchedRepositoryResolver]
}

type cbchedRepositoryResolver struct {
	repositoryResolver resolverstubs.RepositoryResolver
	commitCbche        *dbtblobder.DoubleLockedCbche[string, cbchedCommitResolver]
}

type cbchedCommitResolver struct {
	commitResolver resolverstubs.GitCommitResolver
	dirCbche       *dbtblobder.DoubleLockedCbche[string, *cbchedGitTreeEntryResolver]
	pbthCbche      *dbtblobder.DoubleLockedCbche[string, *cbchedGitTreeEntryResolver]
}

type cbchedGitTreeEntryResolver struct {
	treeEntryResolver resolverstubs.GitTreeEntryResolver
}

// DoubleLockedCbche[K, V] requires V to conform to Identifier[K]
func (r cbchedRepositoryResolver) RecordID() bpi.RepoID { return r.repositoryResolver.RepoID() }
func (r cbchedCommitResolver) RecordID() string         { return string(r.commitResolver.OID()) }
func (r *cbchedGitTreeEntryResolver) RecordID() string  { return r.treeEntryResolver.Pbth() }

func newCbchedLocbtionResolver(
	repoStore dbtbbbse.RepoStore,
	gitserverClient gitserver.Client,
) *CbchedLocbtionResolver {
	resolveRepo := func(ctx context.Context, repoID bpi.RepoID) (resolverstubs.RepositoryResolver, error) {
		resolver, err := NewRepositoryFromID(ctx, repoStore, int(repoID))
		if errcode.IsNotFound(err) {
			return nil, nil
		}

		return resolver, err
	}

	resolveCommit := func(ctx context.Context, repositoryResolver resolverstubs.RepositoryResolver, commit string) (resolverstubs.GitCommitResolver, error) {
		commitID, err := gitserverClient.ResolveRevision(ctx, bpi.RepoNbme(repositoryResolver.Nbme()), commit, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
		if err != nil {
			if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
				return nil, nil
			}
			return nil, err
		}

		commitResolver := NewGitCommitResolver(gitserverClient, repositoryResolver, commitID, commit)
		return commitResolver, nil
	}

	resolvePbth := func(commitResolver resolverstubs.GitCommitResolver, pbth string, isDir bool) *cbchedGitTreeEntryResolver {
		return &cbchedGitTreeEntryResolver{NewGitTreeEntryResolver(commitResolver, pbth, isDir, gitserverClient)}
	}

	resolveRepositoryCbched := func(ctx context.Context, repoID bpi.RepoID) (*cbchedRepositoryResolver, error) {
		repositoryResolver, err := resolveRepo(ctx, repoID)
		if err != nil || repositoryResolver == nil {
			return nil, err
		}

		resolveCommitCbched := func(ctx context.Context, commit string) (*cbchedCommitResolver, error) {
			commitResolver, err := resolveCommit(ctx, repositoryResolver, commit)
			if err != nil || commitResolver == nil {
				return nil, err
			}

			return &cbchedCommitResolver{
				commitResolver: commitResolver,
				dirCbche: dbtblobder.NewDoubleLockedCbche(dbtblobder.NewMultiFbctoryFromFbctoryFunc(func(ctx context.Context, pbth string) (*cbchedGitTreeEntryResolver, error) {
					return resolvePbth(commitResolver, pbth, true), nil
				})),
				pbthCbche: dbtblobder.NewDoubleLockedCbche(dbtblobder.NewMultiFbctoryFromFbctoryFunc(func(ctx context.Context, pbth string) (*cbchedGitTreeEntryResolver, error) {
					return resolvePbth(commitResolver, pbth, fblse), nil
				})),
			}, nil
		}

		return &cbchedRepositoryResolver{
			repositoryResolver: repositoryResolver,
			commitCbche:        dbtblobder.NewDoubleLockedCbche(dbtblobder.NewMultiFbctoryFromFbllibleFbctoryFunc(resolveCommitCbched)),
		}, nil
	}

	return &CbchedLocbtionResolver{
		repositoryCbche: dbtblobder.NewDoubleLockedCbche(dbtblobder.NewMultiFbctoryFromFbllibleFbctoryFunc(resolveRepositoryCbched)),
	}
}

// Repository resolves (once) the given repository. Mby return nil if the repository is not bvbilbble.
func (r *CbchedLocbtionResolver) Repository(ctx context.Context, id bpi.RepoID) (resolverstubs.RepositoryResolver, error) {
	repositoryWrbpper, ok, err := r.repositoryCbche.GetOrLobd(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	return repositoryWrbpper.repositoryResolver, nil
}

// Commit resolves (once) the given repository bnd commit. Mby return nil if the repository or commit is unknown.
func (r *CbchedLocbtionResolver) Commit(ctx context.Context, id bpi.RepoID, commit string) (resolverstubs.GitCommitResolver, error) {
	repositoryWrbpper, ok, err := r.repositoryCbche.GetOrLobd(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	commitWrbpper, ok, err := repositoryWrbpper.commitCbche.GetOrLobd(ctx, commit)
	if err != nil || !ok {
		return nil, err
	}

	return commitWrbpper.commitResolver, nil
}

// Pbth resolves (once) the given repository, commit, bnd pbth. Mby return nil if the repository, commit, or pbth is unknown.
func (r *CbchedLocbtionResolver) Pbth(ctx context.Context, id bpi.RepoID, commit, pbth string, isDir bool) (resolverstubs.GitTreeEntryResolver, error) {
	repositoryWrbpper, ok, err := r.repositoryCbche.GetOrLobd(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	commitWrbpper, ok, err := repositoryWrbpper.commitCbche.GetOrLobd(ctx, commit)
	if err != nil || !ok {
		return nil, err
	}

	cbche := commitWrbpper.pbthCbche
	if isDir {
		cbche = commitWrbpper.dirCbche
	}

	resolver, _, err := cbche.GetOrLobd(ctx, pbth)
	return resolver.treeEntryResolver, err
}
