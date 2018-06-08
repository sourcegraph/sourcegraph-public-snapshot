package git

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

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

// Branches returns a list of all branches in the repository.
func (r *Repository) Branches(ctx context.Context, opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Branches")
	span.SetTag("Opt", opt)
	defer span.Finish()

	f := make(branchFilter)
	if opt.MergedInto != "" {
		b, err := r.branches(ctx, "--merged", opt.MergedInto)
		if err != nil {
			return nil, err
		}
		f.add(b)
	}
	if opt.ContainsCommit != "" {
		b, err := r.branches(ctx, "--contains="+opt.ContainsCommit)
		if err != nil {
			return nil, err
		}
		f.add(b)
	}

	refs, err := r.showRef(ctx, "--heads")
	if err != nil {
		return nil, err
	}

	var branches []*vcs.Branch
	for _, ref := range refs {
		name := strings.TrimPrefix(ref[1], "refs/heads/")
		id := api.CommitID(ref[0])
		if !f.allows(name) {
			continue
		}

		branch := &vcs.Branch{Name: name, Head: id}
		if opt.IncludeCommit {
			branch.Commit, err = r.getCommit(ctx, id)
			if err != nil {
				return nil, err
			}
		}
		if opt.BehindAheadBranch != "" {
			branch.Counts, err = r.BehindAhead(ctx, "refs/heads/"+opt.BehindAheadBranch, "refs/heads/"+name)
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
func (r *Repository) branches(ctx context.Context, args ...string) ([]string, error) {
	cmd := r.command("git", append([]string{"branch"}, args...)...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec %v in %s failed: %v (output follows)\n\n%s", cmd.Args, cmd.Repo, err, out)
	}
	lines := strings.Split(string(out), "\n")
	lines = lines[:len(lines)-1]
	branches := make([]string, len(lines))
	for i, line := range lines {
		branches[i] = line[2:]
	}
	return branches, nil
}

// BehindAhead returns the behind/ahead commit counts information for right vs. left (both Git
// revspecs).
func (r *Repository) BehindAhead(ctx context.Context, left, right string) (*vcs.BehindAhead, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: BehindAhead")
	defer span.Finish()

	if err := checkSpecArgSafety(left); err != nil {
		return nil, err
	}
	if err := checkSpecArgSafety(right); err != nil {
		return nil, err
	}

	cmd := r.command("git", "rev-list", "--count", "--left-right", fmt.Sprintf("%s...%s", left, right))
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, err
	}
	behindAhead := strings.Split(strings.TrimSuffix(string(out), "\n"), "\t")
	b, err := strconv.ParseUint(behindAhead[0], 10, 0)
	if err != nil {
		return nil, err
	}
	a, err := strconv.ParseUint(behindAhead[1], 10, 0)
	if err != nil {
		return nil, err
	}
	return &vcs.BehindAhead{Behind: uint32(b), Ahead: uint32(a)}, nil
}

// Tags returns a list of all tags in the repository.
func (r *Repository) Tags(ctx context.Context) ([]*vcs.Tag, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Tags")
	defer span.Finish()

	// Support both lightweight tags and tag objects. For creatordate, use an %(if) to prefer the
	// taggerdate for tag objects, otherwise use the commit's committerdate (instead of just always
	// using committerdate).
	cmd := r.command("git", "tag", "--list", "--sort", "-creatordate", "--format", "%(if)%(*objectname)%(then)%(*objectname)%(else)%(objectname)%(end)%00%(refname:short)%00%(if)%(creatordate:unix)%(then)%(creatordate:unix)%(else)%(*creatordate:unix)%(end)")
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("listing git tags in %s failed: %s. Output was:\n\n%s", cmd.Repo, err, out)
	}

	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
	if len(out) == 0 {
		return nil, nil // no tags
	}
	lines := bytes.Split(out, []byte("\n"))
	tags := make([]*vcs.Tag, len(lines))
	for i, line := range lines {
		parts := bytes.SplitN(line, []byte("\x00"), 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid git tag list output line: %q", line)
		}
		date, err := strconv.ParseInt(string(parts[2]), 10, 64)
		if err != nil {
			return nil, err
		}
		tags[i] = &vcs.Tag{
			Name:        string(parts[1]),
			CommitID:    api.CommitID(parts[0]),
			CreatorDate: time.Unix(date, 0).UTC(),
		}
	}
	return tags, nil
}

type byteSlices [][]byte

func (p byteSlices) Len() int           { return len(p) }
func (p byteSlices) Less(i, j int) bool { return bytes.Compare(p[i], p[j]) < 0 }
func (p byteSlices) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (r *Repository) showRef(ctx context.Context, arg string) ([][2]string, error) {
	cmd := r.command("git", "show-ref", arg)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return nil, err
		}
		// Exit status of 1 and no output means there were no
		// results. This is not a fatal error.
		if cmd.ExitStatus == 1 && len(out) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("exec `git show-ref %s` in %s failed: %s. Output was:\n\n%s", arg, cmd.Repo, err, out)
	}

	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
	lines := bytes.Split(out, []byte("\n"))
	sort.Sort(byteSlices(lines)) // sort for consistency
	refs := make([][2]string, len(lines))
	for i, line := range lines {
		if len(line) <= 41 {
			return nil, errors.New("unexpectedly short (<=41 bytes) line in `git show-ref ...` output")
		}
		id := line[:40]
		name := line[41:]
		refs[i] = [2]string{string(id), string(name)}
	}
	return refs, nil
}
