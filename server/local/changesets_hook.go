package local

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/gitserver/gitpb"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

func init() {
	AddGitPostPushHook("changesets.hook", func(ctx context.Context, op *gitpb.ReceivePackOp, events []githttp.Event) {
		hook := newChangesetHook(ctx, op.Repo.URI)
		for _, e := range events {
			if couldAffectChangesets(e) {
				hook.processEvent(e)
			}
		}
	})
}

// changesetHook tracks each push to the server and looks for open changesets
// that could be affected, updating them as needed.
type changesetHook struct {
	ctx        context.Context
	changesets store.Changesets
	repo       vcs.Repository
	repoURI    string
}

func newChangesetHook(ctx context.Context, repoURI string) changesetHook {
	repo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoURI)
	if err != nil {
		log.Printf("error in postPushHook, cannot open repo '%s': %v", repoURI, err)
	}
	return changesetHook{
		ctx:        ctx,
		changesets: store.ChangesetsFromContext(ctx),
		repo:       repo,
		repoURI:    repoURI,
	}
}

// processEvent processes a githttp.Event and alters open changesets based
// on how they are affected by it.
func (h changesetHook) processEvent(e githttp.Event) {
	// find open changesets that have the pushed branch as HEAD:
	havingHead, err := h.changesets.List(h.ctx, &sourcegraph.ChangesetListOp{
		Repo: h.repoURI,
		Open: true,
		Head: e.Branch,
	})
	if err != nil && !os.IsNotExist(err) {
		log.Printf("error in postPushHook, cannot list changesets for head: %v", err)
	}
	h.updateHavingHead(havingHead.Changesets, e)

	// find open changesets that have the pushed branch as BASE:
	havingBase, err := h.changesets.List(h.ctx, &sourcegraph.ChangesetListOp{
		Repo: h.repoURI,
		Open: true,
		Base: e.Branch,
	})
	if err != nil && !os.IsNotExist(err) {
		log.Printf("error in postPushHook, cannot list changesets for base: %v", err)
	}
	h.updateHavingBase(havingBase.Changesets, e)
}

// updateHavingHead updates a list of changesets that have the passed event's
// branch as HEAD.
//
// For example:
// - If the branch was deleted, we close the changesets.
// - If the branch was comitted into, we update the changesets to reflect the new HEAD.
func (h changesetHook) updateHavingHead(list []*sourcegraph.Changeset, e githttp.Event) {
	var wg sync.WaitGroup
	for _, cs := range list {
		wg.Add(1)
		go func(cs *sourcegraph.Changeset) {
			defer wg.Done()
			switch {
			case zeroCommit(e.Commit): // branch deleted
				base, err := h.repo.ResolveBranch(cs.DeltaSpec.Base.Rev)
				if err != nil {
					log.Printf("error in postPushHook, cannot resolve base branch '%s': %v", cs.DeltaSpec.Base.Rev, err)
				}
				h.update(&store.ChangesetUpdateOp{
					Op: &sourcegraph.ChangesetUpdateOp{
						Repo:  cs.DeltaSpec.Base.RepoSpec,
						ID:    cs.ID,
						Close: true,
					},
					Head: e.Last,
					Base: string(base),
				})

			default: // regular commit
				h.update(&store.ChangesetUpdateOp{
					Op: &sourcegraph.ChangesetUpdateOp{
						Repo: cs.DeltaSpec.Base.RepoSpec,
						ID:   cs.ID,
					},
					Head: e.Commit,
				})
			}

		}(cs)
	}
	wg.Wait()
}

// updateHavingBase updates a list of changesets having the passed event's
// branch as BASE.
//
// For example:
// - If the branch was deleted, we close the changesets and save the last commit.
// - If the branch contained the merge of any changeset in the list, we mark them
// as merged.
func (h changesetHook) updateHavingBase(changesets []*sourcegraph.Changeset, e githttp.Event) {
	if len(changesets) == 0 {
		return
	}
	mergedBranches := make(branchMap)
	isMerged := func(b string) bool { _, ok := mergedBranches[b]; return ok }
	if !zeroCommit(e.Commit) {
		mergedBranches = h.mergedInto(e.Branch)
	}
	var wg sync.WaitGroup
	for _, cs := range changesets {
		wg.Add(1)
		go func(cs *sourcegraph.Changeset) {
			defer wg.Done()
			switch {
			case zeroCommit(e.Commit): // branch deleted
				h.update(&store.ChangesetUpdateOp{
					Op: &sourcegraph.ChangesetUpdateOp{
						Repo:  cs.DeltaSpec.Base.RepoSpec,
						ID:    cs.ID,
						Close: true,
					},
					Base: e.Last,
				})

			case isMerged(cs.DeltaSpec.Head.Rev): // contained merge
				head, err := h.repo.ResolveBranch(cs.DeltaSpec.Head.Rev)
				if err != nil {
					log.Printf("error in postPushHook, cannot resolve rev '%s': %v", cs.DeltaSpec.Head.Rev, err)
				}
				h.update(&store.ChangesetUpdateOp{
					Op: &sourcegraph.ChangesetUpdateOp{
						Repo:   cs.DeltaSpec.Base.RepoSpec,
						ID:     cs.ID,
						Close:  true,
						Merged: true,
					},
					Base: e.Last,
					Head: string(head),
				})
			}
		}(cs)
	}
	wg.Wait()
}

// branchMap indexes a list of branches.
type branchMap map[string]struct{}

// mergedInto returns a branchMap of all branches that were merged into branch.
func (h changesetHook) mergedInto(branch string) branchMap {
	bm := make(branchMap)
	branches, err := h.repo.Branches(vcs.BranchesOptions{MergedInto: branch})
	if err != nil {
		log.Printf("error in postPushHook, cannot retrieve branches: %v", err)
	}
	for _, b := range branches {
		if b.Name != branch {
			bm[b.Name] = struct{}{}
		}
	}
	return bm
}

func (h changesetHook) update(op *store.ChangesetUpdateOp) {
	updateChangeset(h.changesets, h.ctx, op)
}

var updateChangeset = func(cs store.Changesets, ctx context.Context, op *store.ChangesetUpdateOp) {
	if op == nil {
		return
	}
	if _, err := cs.Update(ctx, op); err != nil {
		if !strings.Contains(err.Error(), "working directory clean") {
			log.Printf("error in postPushHook, cannot update changeset ref: %v", err)
		}
	}
}

// couldAffectChangesets returns true if the event was error-free, a PUSH
// operation and not a new branch (which can not affect any existing changesets).
func couldAffectChangesets(e githttp.Event) bool {
	if e.Error != nil {
		return false
	}
	if e.Type != githttp.PUSH || e.Branch == "" {
		return false
	}
	if zeroCommit(e.Last) || !commitsValid(e.Commit, e.Last) {
		return false
	}
	return true
}

// zeroCommit returns true if this is a commit that no longer
// exists or has never existed (a branch was deleted or created).
func zeroCommit(c string) bool {
	return c == "0000000000000000000000000000000000000000"
}

// commitsValid returns true if all commits in the paramters are exactly 40
// characters long.
func commitsValid(commits ...string) bool {
	for _, c := range commits {
		if len(c) != 40 {
			return false
		}
	}
	return true
}
