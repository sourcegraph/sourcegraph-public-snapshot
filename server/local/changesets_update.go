package local

import (
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"

	"src.sourcegraph.com/sourcegraph/server/internal/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

func (s *changesets) Update(ctx context.Context, op *sourcegraph.ChangesetUpdateOp) (*sourcegraph.ChangesetEvent, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.Update"); err != nil {
		return nil, err
	}

	defer noCache(ctx)

	return store.ChangesetsFromContext(ctx).Update(ctx, &store.ChangesetUpdateOp{Op: op})
}

func (s *changesets) UpdateAffected(ctx context.Context, op *sourcegraph.ChangesetUpdateAffectedOp) (*sourcegraph.ChangesetEventList, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.UpdateAffected"); err != nil {
		return nil, err
	}

	defer noCache(ctx)

	if op == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "empty argument")
	}

	changesetsStore := store.ChangesetsFromContextOrNil(ctx)
	if changesetsStore == nil {
		return nil, grpc.Errorf(codes.Internal, "no changesets store in context")
	}

	// Get ChangesetUpdateOps for the affected changesets.
	updates, err := s.getAffected(ctx, op)
	if err != nil {
		return nil, err
	}

	// Execute all changeset updates.
	var res sourcegraph.ChangesetEventList
	for _, updateOp := range updates {
		if e, err := changesetsStore.Update(ctx, updateOp); err != nil {
			log15.Error("Changesets.UpdateAffected: cannot update changeset", "repo", updateOp.Op.Repo, "id", updateOp.Op.ID, "error", err)
		} else {
			res.Events = append(res.Events, e)
		}
	}

	return &res, nil
}

func (s *changesets) getAffected(ctx context.Context, op *sourcegraph.ChangesetUpdateAffectedOp) ([]*store.ChangesetUpdateOp, error) {
	repoVCS, err := store.RepoVCSFromContext(ctx).Open(ctx, op.Repo.URI)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "cannot open repo vcs %v: %v", op.Repo.URI, err)
	}

	changesetsStore := store.ChangesetsFromContextOrNil(ctx)
	if changesetsStore == nil {
		return nil, grpc.Errorf(codes.Internal, "no changesets store in context")
	}

	// Find open changesets that have the pushed branch as HEAD:
	havingHead, err := changesetsStore.List(ctx, &sourcegraph.ChangesetListOp{
		Repo: op.Repo.URI,
		Open: true,
		Head: op.Branch,
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, grpc.Errorf(codes.Internal, "cannot list changesets for head: %v", err)
	}

	// Find open changesets that have the pushed branch as BASE:
	havingBase, err := changesetsStore.List(ctx, &sourcegraph.ChangesetListOp{
		Repo: op.Repo.URI,
		Open: true,
		Base: op.Branch,
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, grpc.Errorf(codes.Internal, "cannot list changesets for base: %v", err)
	}

	isBranchDeleted := (op.Commit == EmptyCommitID)

	// Record all changeset updates to be executed.
	var updates []*store.ChangesetUpdateOp

	// For changesets with affected HEAD:
	// - If the branch was deleted, close changesets.
	// - If the branch was comitted into, update the changeset to reflect the new HEAD.
	for _, cs := range havingHead.Changesets {
		updateOp := store.ChangesetUpdateOp{
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo: cs.DeltaSpec.Base.RepoSpec,
				ID:   cs.ID,
			},
			Head: op.Commit,
		}
		if isBranchDeleted {
			base, err := repoVCS.ResolveBranch(cs.DeltaSpec.Base.Rev)
			if err != nil {
				log15.Error("Changesets.UpdateAffected: cannot resolve base branch", "rev", cs.DeltaSpec.Base.Rev, "error", err)
				continue
			}
			updateOp.Op.Close = true
			updateOp.Head = op.Last
			updateOp.Base = string(base)
		}
		updates = append(updates, &updateOp)
	}

	// For changesets with affected BASE:
	// - If the branch was deleted, close the changesets and save the last commit.
	// - If the branch contained the merge of the changeset, mark it as merged.
	mergedBranches := make(branchMap)
	isMerged := func(b string) bool { _, ok := mergedBranches[b]; return ok }
	if !isBranchDeleted {
		mergedBranches = mergedInto(repoVCS, op.Branch)
	}
	for _, cs := range havingBase.Changesets {
		isBranchMerged := isMerged(cs.DeltaSpec.Head.Rev)
		updateOp := store.ChangesetUpdateOp{
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo:  cs.DeltaSpec.Base.RepoSpec,
				ID:    cs.ID,
				Close: true,
			},
			Base: op.Last,
		}
		if !isBranchDeleted && isBranchMerged {
			head, err := repoVCS.ResolveBranch(cs.DeltaSpec.Head.Rev)
			if err != nil {
				log15.Error("Changesets.UpdateAffected: cannot resolve head branch", "rev", cs.DeltaSpec.Head.Rev, "error", err)
			}
			updateOp.Op.Merged = true
			updateOp.Head = string(head)
		}
		if isBranchDeleted || isBranchMerged {
			updates = append(updates, &updateOp)
		}
	}

	return updates, nil
}

// branchMap indexes a list of branches.
type branchMap map[string]struct{}

// mergedInto returns a branchMap of all branches that were merged into branch.
func mergedInto(repoVCS vcs.Repository, branch string) branchMap {
	bm := make(branchMap)
	branches, err := repoVCS.Branches(vcs.BranchesOptions{MergedInto: branch})
	if err != nil {
		log15.Error("Changesets: cannot retrieve branches", "error", err)
	}
	for _, b := range branches {
		if b.Name != branch {
			bm[b.Name] = struct{}{}
		}
	}
	return bm
}
