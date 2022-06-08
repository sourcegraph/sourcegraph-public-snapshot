package git

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// BranchesOptions specifies options for the list of branches returned by
// (Repository).Branches.
type BranchesOptions struct {
	// MergedInto will cause the returned list to be restricted to only
	// branches that were merged into this branch name.
	MergedInto string `json:"MergedInto,omitempty" url:",omitempty"`
	// IncludeCommit controls whether complete commit information is included.
	IncludeCommit bool `json:"IncludeCommit,omitempty" url:",omitempty"`
	// BehindAheadBranch specifies a branch name. If set to something other than blank
	// string, then each returned branch will include a behind/ahead commit counts
	// information against the specified base branch. If left blank, then branches will
	// not include that information and their Counts will be nil.
	BehindAheadBranch string `json:"BehindAheadBranch,omitempty" url:",omitempty"`
	// ContainsCommit filters the list of branches to only those that
	// contain a specific commit ID (if set).
	ContainsCommit string `json:"ContainsCommit,omitempty" url:",omitempty"`
}

// branchFilter is a filter for branch names.
// If not empty, only contained branch names are allowed. If empty, all names are allowed.
// The map should be made so it's not nil.
type branchFilter map[string]struct{}

// allows will return true if the current filter set-up validates against
// the passed string. If there are no filters, all strings pass.
func (f branchFilter) allows(name string) bool {
	if len(f) == 0 {
		return true
	}
	_, ok := f[name]
	return ok
}

// add adds a slice of strings to the filter.
func (f branchFilter) add(list []string) {
	for _, l := range list {
		f[l] = struct{}{}
	}
}

// ListBranches returns a list of all branches in the repository.
func ListBranches(ctx context.Context, db database.DB, repo api.RepoName, opt BranchesOptions) ([]*gitdomain.Branch, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: Branches")
	span.SetTag("Opt", opt)
	defer span.Finish()

	client := gitserver.NewClient(db)
	f := make(branchFilter)
	if opt.MergedInto != "" {
		b, err := branches(ctx, client, repo, "--merged", opt.MergedInto)
		if err != nil {
			return nil, err
		}
		f.add(b)
	}
	if opt.ContainsCommit != "" {
		b, err := branches(ctx, client, repo, "--contains="+opt.ContainsCommit)
		if err != nil {
			return nil, err
		}
		f.add(b)
	}

	refs, err := showRef(ctx, db, repo, "--heads")
	if err != nil {
		return nil, err
	}

	var branches []*gitdomain.Branch
	for _, ref := range refs {
		name := strings.TrimPrefix(ref.Name, "refs/heads/")
		if !f.allows(name) {
			continue
		}

		branch := &gitdomain.Branch{Name: name, Head: ref.CommitID}
		if opt.IncludeCommit {
			branch.Commit, err = getCommit(ctx, db, repo, ref.CommitID, gitserver.ResolveRevisionOptions{}, authz.DefaultSubRepoPermsChecker)
			if err != nil {
				return nil, err
			}
		}
		if opt.BehindAheadBranch != "" {
			branch.Counts, err = client.GetBehindAhead(ctx, repo, "refs/heads/"+opt.BehindAheadBranch, "refs/heads/"+name)
			if err != nil {
				return nil, err
			}
		}
		branches = append(branches, branch)
	}
	return branches, nil
}

// branches runs the `git branch` command followed by the given arguments and
// returns the list of branches if successful.
func branches(ctx context.Context, client gitserver.Client, repo api.RepoName, args ...string) ([]string, error) {
	cmd := client.GitCommand(repo, append([]string{"branch"}, args...)...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.Errorf("exec %v in %s failed: %v (output follows)\n\n%s", cmd.Args(), cmd.Repo(), err, out)
	}
	lines := strings.Split(string(out), "\n")
	lines = lines[:len(lines)-1]
	branches := make([]string, len(lines))
	for i, line := range lines {
		branches[i] = line[2:]
	}
	return branches, nil
}

type byteSlices [][]byte

func (p byteSlices) Len() int           { return len(p) }
func (p byteSlices) Less(i, j int) bool { return bytes.Compare(p[i], p[j]) < 0 }
func (p byteSlices) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// ListRefs returns a list of all refs in the repository.
func ListRefs(ctx context.Context, db database.DB, repo api.RepoName) ([]gitdomain.Ref, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: ListRefs")
	defer span.Finish()
	return showRef(ctx, db, repo)
}

func showRef(ctx context.Context, db database.DB, repo api.RepoName, args ...string) ([]gitdomain.Ref, error) {
	cmdArgs := append([]string{"show-ref"}, args...)
	cmd := gitserver.NewClient(db).GitCommand(repo, cmdArgs...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if gitdomain.IsRepoNotExist(err) {
			return nil, err
		}
		// Exit status of 1 and no output means there were no
		// results. This is not a fatal error.
		if cmd.ExitStatus() == 1 && len(out) == 0 {
			return nil, nil
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}

	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
	lines := bytes.Split(out, []byte("\n"))
	sort.Sort(byteSlices(lines)) // sort for consistency
	refs := make([]gitdomain.Ref, len(lines))
	for i, line := range lines {
		if len(line) <= 41 {
			return nil, errors.New("unexpectedly short (<=41 bytes) line in `git show-ref ...` output")
		}
		id := line[:40]
		name := line[41:]
		refs[i] = gitdomain.Ref{Name: string(name), CommitID: api.CommitID(id)}
	}
	return refs, nil
}
