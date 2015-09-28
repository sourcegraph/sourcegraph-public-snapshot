package local

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"golang.org/x/net/context"

	"github.com/AaronO/go-git-http"
	ppretty "github.com/kr/pretty"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	vcstesting "sourcegraph.com/sourcegraph/go-vcs/vcs/testing"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

// TestChangesetHook_couldAffectChangesets tests if a list of events is correctly
// filtered by couldAffectChangesets.
func TestChangesetHook_couldAffectChangesets(t *testing.T) {
	for _, tc := range []struct {
		in  githttp.Event
		out bool
	}{
		{
			// contains an error
			githttp.Event{
				Error:  errors.New("some error"),
				Branch: "branch",
				Type:   githttp.PUSH,
				Last:   strings.Repeat("X", 40),
				Commit: strings.Repeat("Y", 40),
			}, false,
		}, {
			// is not a PUSH operation
			githttp.Event{
				Error:  nil,
				Branch: "branch",
				Type:   githttp.FETCH,
				Last:   strings.Repeat("X", 40),
				Commit: strings.Repeat("Y", 40),
			}, false,
		}, {
			// is not a PUSH operation
			githttp.Event{
				Error:  nil,
				Branch: "branch",
				Type:   githttp.TAG,
				Last:   strings.Repeat("X", 40),
				Commit: strings.Repeat("Y", 40),
			}, false,
		}, {
			// is a new branch
			githttp.Event{
				Error:  nil,
				Branch: "branch",
				Type:   githttp.PUSH,
				Last:   strings.Repeat("0", 40),
				Commit: strings.Repeat("Y", 40),
			}, false,
		}, {
			// invalid commit value
			githttp.Event{
				Error:  nil,
				Branch: "branch",
				Type:   githttp.PUSH,
				Last:   strings.Repeat("X", 40),
				Commit: "HEAD",
			}, false,
		}, {
			// push commit
			githttp.Event{
				Error:  nil,
				Branch: "branch",
				Type:   githttp.PUSH,
				Last:   strings.Repeat("X", 40),
				Commit: strings.Repeat("Y", 40),
			}, true,
		}, {
			// push branch deletion
			githttp.Event{
				Error:  nil,
				Branch: "branch",
				Type:   githttp.PUSH,
				Last:   strings.Repeat("X", 40),
				Commit: strings.Repeat("0", 40),
			}, true,
		},
	} {
		if out := couldAffectChangesets(tc.in); out != tc.out {
			t.Errorf("Expected %v for %v, got %v", tc.out, tc.in, out)
		}
	}
}

// TestChangesetHook_processEvent tests that for a certain chain of events
// the correct update operations are performed.
func TestChangesetHook_processEvent(t *testing.T) {
	tt := newChangesetHookTester(t)

	tt.withRepo(sourcegraph.RepoSpec{URI: "repo/path"})
	tt.withChangesets(
		tt.cs(1, "master", "feature"),
		tt.cs(2, "dev", "feature"),
	)
	tt.withMergedInto("master", "feature")

	tt.run([]githttp.Event{
		{
			// (1) pushing 'master' contains merge of 'feature'
			Type:   githttp.PUSH,
			Branch: "master",
			Last:   fakeCommit("old(master)"),
			Commit: fakeCommit("new(master+feature)"),
		}, {
			// (2) push deleted branch 'master'
			Type:   githttp.PUSH,
			Branch: "master",
			Last:   fakeCommit("old(master)"),
			Commit: fakeCommit("DELETED"),
		}, {
			// (3) push updated branch 'feature'
			Type:   githttp.PUSH,
			Branch: "feature",
			Last:   fakeCommit("old(feature)"),
			Commit: fakeCommit("new(feature)"),
		}, {
			// (4) push deleted branch 'feature'
			Type:   githttp.PUSH,
			Branch: "feature",
			Last:   fakeCommit("old(feature)"),
			Commit: fakeCommit("DELETED"),
		},
	})

	tt.expect([]*store.ChangesetUpdateOp{
		{
			// (1)
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo:   tt.repo.spec,
				ID:     1,
				Close:  true,
				Merged: true,
			},
			Base: fakeCommit("old(master)"),
			Head: fakeCommit("feature"),
		}, {
			// (2)
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo:  tt.repo.spec,
				ID:    1,
				Close: true,
			},
			Base: fakeCommit("old(master)"),
		}, {
			// (3)
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo: tt.repo.spec,
				ID:   1,
			},
			Head: fakeCommit("new(feature)"),
		}, {
			// (3)
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo: tt.repo.spec,
				ID:   2,
			},
			Head: fakeCommit("new(feature)"),
		}, {
			// (4)
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo:  tt.repo.spec,
				ID:    1,
				Close: true,
			},
			Base: fakeCommit("master"),
			Head: fakeCommit("old(feature)"),
		}, {
			// (4)
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo:  tt.repo.spec,
				ID:    2,
				Close: true,
			},
			Base: fakeCommit("dev"),
			Head: fakeCommit("old(feature)"),
		},
	})
}

// changesetHookTester is a utility that aids with testing this hook.
type changesetHookTester struct {
	t    *testing.T
	ctx  context.Context
	mock *mocks

	changesets changesetIndex
	repo       testRepoConfig

	out []*store.ChangesetUpdateOp
}

// testRepoConfig determines the configuration for mocking RepoVCS.
type testRepoConfig struct {
	// spec holds the RepoSpec these tests will use
	spec sourcegraph.RepoSpec

	// mergedInto maps branches to a list of branches that
	// should resolve as merged into it (for the mock).
	mergedInto map[string][]string
}

// branchChangesets maps branch names to a list of changesets that they
// appear in.
type branchChangesets map[string][]*sourcegraph.Changeset

func newChangesetHookTester(t *testing.T) changesetHookTester {
	ctx, mock := testContext()
	tester := changesetHookTester{
		t:    t,
		ctx:  ctx,
		mock: mock,
		repo: testRepoConfig{mergedInto: make(map[string][]string)},
	}
	return tester
}

// withRepo sets repository for the context of the test.
func (tt *changesetHookTester) withRepo(spec sourcegraph.RepoSpec) {
	tt.repo.spec = spec
}

// withMergedInto makes the next push into base contain the merge of head.
func (tt *changesetHookTester) withMergedInto(base, head string) {
	m := tt.repo.mergedInto
	if m[base] == nil {
		m[base] = []string{base}
	}
	m[base] = append(m[base], head)
}

// changesetIndex manages an index of changesets used to mock return
// values for List commands that request by-Base or by-Head.
type changesetIndex struct {
	// byHead maps a branch name to the list of changesets
	// where that branch is HEAD (for mocking).
	byHead branchChangesets

	// byBase maps a branch name to the list of changesets
	// where that branch is BASE (for mocking).
	byBase branchChangesets
}

// withChangesets sets mock behavior of the Changesets store.
func (tt *changesetHookTester) withChangesets(changesets ...*sourcegraph.Changeset) {
	tt.changesets = changesetIndex{make(branchChangesets), make(branchChangesets)}
	base, head := tt.changesets.byBase, tt.changesets.byHead
	for _, cs := range changesets {
		b, h := cs.DeltaSpec.Base.Rev, cs.DeltaSpec.Head.Rev
		if base[b] == nil {
			base[b] = make([]*sourcegraph.Changeset, 0)
		}
		if head[h] == nil {
			head[h] = make([]*sourcegraph.Changeset, 0)
		}
		base[b] = append(base[b], cs)
		head[h] = append(head[h], cs)
	}
}

// run runs a set of events and stores all update operations that occurred
// into a buffer.
func (tt *changesetHookTester) run(events []githttp.Event) {
	tt.configRepoStore()
	tt.configChangesetStore()

	var m sync.Mutex // protects tt.out
	tt.out = make([]*store.ChangesetUpdateOp, 0)
	updateChangeset = func(_ store.Changesets, _ context.Context, op *store.ChangesetUpdateOp) {
		m.Lock()
		defer m.Unlock()
		tt.out = append(tt.out, op)
	}

	hook := newChangesetHook(tt.ctx, tt.repo.spec.URI)
	for _, e := range events {
		hook.processEvent(e)
	}
}

// expect verifies that the given update operations have occurred. The order of
// operations need not (and will not) match due to concurrent execution.
func (tt *changesetHookTester) expect(ops []*store.ChangesetUpdateOp) {
	if len(ops) != len(tt.out) {
		tt.t.Errorf("got %d ops, expected %d ops", len(tt.out), len(ops))
	}
	p := ppretty.Formatter
outer:
	for _, op := range ops {
		for _, exp := range tt.out {
			if reflect.DeepEqual(exp, op) {
				continue outer
			}
		}
		tt.t.Errorf("ops mismatch error\n\nGOT:\n\n%# v\n\nEXPECTED:\n\n%# v\n", p(tt.out), p(ops))
		break
	}
}

func (tt *changesetHookTester) configRepoStore() {
	tt.mock.stores.RepoVCS.MockOpen(tt.t, tt.repo.spec.URI, vcstesting.MockRepository{
		Branches_: func(opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
			branches := tt.repo.mergedInto[opt.MergedInto]
			b := make([]*vcs.Branch, len(branches))
			for i, branch := range branches {
				b[i] = &vcs.Branch{Name: branch}
			}
			return b, nil
		},
		ResolveBranch_: func(name string) (vcs.CommitID, error) {
			return vcs.CommitID(fakeCommit(name)), nil
		},
	})
}

func (tt *changesetHookTester) configChangesetStore() {
	tt.mock.stores.Changesets.List_ = func(_ context.Context, op *sourcegraph.ChangesetListOp) (*sourcegraph.ChangesetList, error) {
		switch {
		case op.Base != "":
			return &sourcegraph.ChangesetList{
				Changesets: tt.changesets.byBase[op.Base],
			}, nil
		case op.Head != "":
			return &sourcegraph.ChangesetList{
				Changesets: tt.changesets.byHead[op.Head],
			}, nil
		}
		return nil, nil
	}
}

// fakeCommit returns a fake commit having b as prefix so it can be easily
// identified in test error output.
func fakeCommit(b string) string {
	if b == "DELETED" {
		return strings.Repeat("0", 40)
	}
	return b + strings.Repeat("+", 40-len(b)-1)
}

// cs creates a minimal changeset.
func (tt *changesetHookTester) cs(id int64, base, head string) *sourcegraph.Changeset {
	return &sourcegraph.Changeset{
		ID: id,
		DeltaSpec: &sourcegraph.DeltaSpec{
			Base: sourcegraph.RepoRevSpec{RepoSpec: tt.repo.spec, Rev: base},
			Head: sourcegraph.RepoRevSpec{RepoSpec: tt.repo.spec, Rev: head},
		},
	}
}
