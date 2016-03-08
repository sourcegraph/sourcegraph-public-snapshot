package local

import (
	"errors"
	"strings"

	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-diff/diff"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

var Deltas sourcegraph.DeltasServer = &deltas{}

type deltas struct {
	// mockDiffFunc, if set, is called by (deltas).diff instead of the
	// main method body. It allows mocking (deltas).diff in tests.
	mockDiffFunc func(context.Context, sourcegraph.DeltaSpec) ([]*diff.FileDiff, *sourcegraph.Delta, error)
}

var _ sourcegraph.DeltasServer = (*deltas)(nil)

func (s *deltas) Get(ctx context.Context, ds *sourcegraph.DeltaSpec) (*sourcegraph.Delta, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Deltas.Get", ds.Base.URI); err != nil {
		return nil, err
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Deltas.Get", ds.Head.URI); err != nil {
		return nil, err
	}

	d := &sourcegraph.Delta{
		Base: ds.Base,
		Head: ds.Head,
	}

	get := func(repo **sourcegraph.Repo, repoRevSpec *sourcegraph.RepoRevSpec, commit **vcs.Commit) error {
		var err error
		*repo, err = svc.Repos(ctx).Get(ctx, &repoRevSpec.RepoSpec)
		if err != nil {
			return err
		}
		*commit, err = svc.Repos(ctx).GetCommit(ctx, repoRevSpec)
		if err != nil {
			return err
		}
		repoRevSpec.CommitID = string((*commit).ID)
		return nil
	}

	par := parallel.NewRun(2)
	par.Do(func() error { return get(&d.BaseRepo, &d.Base, &d.BaseCommit) })
	par.Do(func() error { return get(&d.HeadRepo, &d.Head, &d.HeadCommit) })
	if err := par.Wait(); err != nil {
		return d, err
	}

	// Try to compute merge-base.
	vcsBaseRepo, err := store.RepoVCSFromContext(ctx).Open(ctx, d.BaseRepo.URI)
	if err != nil {
		return d, err
	}

	if d.BaseRepo.URI != d.HeadRepo.URI {
		return d, errors.New("base and head repo must be identical")
	}

	id, err := vcsBaseRepo.MergeBase(vcs.CommitID(d.BaseCommit.ID), vcs.CommitID(d.HeadCommit.ID))
	if err != nil {
		return d, err
	}

	if d.BaseCommit.ID != id {
		ds2 := *ds
		// There is most likely a merge conflict here, so we update the
		// delta to contain the actual merge base used in this diff A...B
		ds2.Base.CommitID = string(id)
		if strings.HasPrefix(ds.Base.CommitID, ds.Base.Rev) {
			// If the Revision is not a branch, but the commit ID, clear it.
			ds2.Base.Rev = ""
		}
		return s.Get(ctx, &ds2)
	}

	return d, nil
}
