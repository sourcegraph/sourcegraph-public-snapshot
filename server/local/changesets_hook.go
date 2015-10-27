package local

import (
	"os"
	"strings"
	"sync"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif/githooks"
	"src.sourcegraph.com/sourcegraph/store"
)

func init() {
	events.Listeners = append(events.Listeners, &changesetHookListener{})
}

type changesetHookListener struct{}

func (g *changesetHookListener) Scopes() []string {
	return []string{"app:changes"}
}

func (g *changesetHookListener) Start(ctx context.Context) {
	callback := func(p githooks.Payload) {
		e := p.Event
		if e.Error != nil || e.Branch == "" || !commitsValid(e.Commit, e.Last) {
			return
		}
		hook := newChangesetHook(ctx, p.Repo.URI)
		hook.processEvent(p)
	}

	events.Subscribe(githooks.GitPushEvent, callback)
	events.Subscribe(githooks.GitDeleteEvent, callback)
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
		log15.Warn("changesetHook: cannot open repo", "repo", repoURI, "error", err)
	}
	return changesetHook{
		ctx:        ctx,
		changesets: store.ChangesetsFromContext(ctx),
		repo:       repo,
		repoURI:    repoURI,
	}
}

// processEvent processes a githooks.Payload and alters open changesets based
// on how they are affected by it.
func (h changesetHook) processEvent(p githooks.Payload) {
	e := p.Event
	// find open changesets that have the pushed branch as HEAD:
	havingHead, err := h.changesets.List(h.ctx, &sourcegraph.ChangesetListOp{
		Repo: h.repoURI,
		Open: true,
		Head: e.Branch,
	})
	if err != nil && !os.IsNotExist(err) {
		log15.Warn("changesetHook: cannot list changesets for head", "error", err)
	}
	h.updateHavingHead(havingHead.Changesets, p)

	// find open changesets that have the pushed branch as BASE:
	havingBase, err := h.changesets.List(h.ctx, &sourcegraph.ChangesetListOp{
		Repo: h.repoURI,
		Open: true,
		Base: e.Branch,
	})
	if err != nil && !os.IsNotExist(err) {
		log15.Warn("changesetHook: cannot list changesets for base", "error", err)
	}
	h.updateHavingBase(havingBase.Changesets, p)
}

// updateHavingHead updates a list of changesets that have the passed event's
// branch as HEAD.
//
// For example:
// - If the branch was deleted, we close the changesets.
// - If the branch was comitted into, we update the changesets to reflect the new HEAD.
func (h changesetHook) updateHavingHead(list []*sourcegraph.Changeset, p githooks.Payload) {
	e := p.Event
	var wg sync.WaitGroup
	for _, cs := range list {
		wg.Add(1)
		go func(cs *sourcegraph.Changeset) {
			defer wg.Done()
			switch p.Type {
			case githooks.GitDeleteEvent: // branch deleted
				base, err := h.repo.ResolveBranch(cs.DeltaSpec.Base.Rev)
				if err != nil {
					log15.Warn("changesetHook: cannot resolve base branch", "base", cs.DeltaSpec.Base.Rev, "error", err)
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
func (h changesetHook) updateHavingBase(changesets []*sourcegraph.Changeset, p githooks.Payload) {
	e := p.Event
	if len(changesets) == 0 {
		return
	}
	mergedBranches := make(branchMap)
	isMerged := func(b string) bool { _, ok := mergedBranches[b]; return ok }
	if p.Type != githooks.GitDeleteEvent {
		mergedBranches = h.mergedInto(e.Branch)
	}
	var wg sync.WaitGroup
	for _, cs := range changesets {
		wg.Add(1)
		go func(cs *sourcegraph.Changeset) {
			defer wg.Done()
			switch {
			case p.Type == githooks.GitDeleteEvent: // branch deleted
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
					log15.Warn("changesetHook: cannot resolve rev", "rev", cs.DeltaSpec.Head.Rev, "error", err)
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
		log15.Warn("changesetHook: cannot retrieve branches", "error", err)
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
			log15.Warn("changesetHook: cannot update changeset ref", "error", err)
		}
	}
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
