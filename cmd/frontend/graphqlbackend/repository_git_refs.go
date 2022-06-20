package graphqlbackend

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type refsArgs struct {
	graphqlutil.ConnectionArgs
	Query       *string
	Type        *string
	OrderBy     *string
	Interactive bool
}

func (r *RepositoryResolver) Branches(ctx context.Context, args *refsArgs) (*gitRefConnectionResolver, error) {
	t := gitRefTypeBranch
	args.Type = &t
	return r.GitRefs(ctx, args)
}

func (r *RepositoryResolver) Tags(ctx context.Context, args *refsArgs) (*gitRefConnectionResolver, error) {
	t := gitRefTypeTag
	args.Type = &t
	return r.GitRefs(ctx, args)
}

func (r *RepositoryResolver) GitRefs(ctx context.Context, args *refsArgs) (*gitRefConnectionResolver, error) {
	var branches []*gitdomain.Branch
	db := r.db
	if args.Type == nil || *args.Type == gitRefTypeBranch {
		var err error
		branches, err = git.ListBranches(ctx, db, r.RepoName(), git.BranchesOptions{
			// We intentionally do not ask for commits here since it requires
			// a separate git call per branch. We only need the git commits to
			// sort by author/commit date and there are few enough branches to
			// warrant doing it interactively.
			IncludeCommit: false,
		})
		if err != nil {
			return nil, err
		}

		// Filter before calls to GetCommit. This hopefully reduces the
		// working set enough that we can sort interactively.
		if args.Query != nil {
			query := strings.ToLower(*args.Query)

			filtered := branches[:0]
			for _, branch := range branches {
				if strings.Contains(strings.ToLower(branch.Name), query) {
					filtered = append(filtered, branch)
				}
			}
			branches = filtered
		}

		if args.OrderBy != nil && *args.OrderBy == gitRefOrderAuthoredOrCommittedAt {
			// Sort branches by most recently committed.

			ok, err := hydrateBranchCommits(ctx, db, r.RepoName(), args.Interactive, branches)
			if err != nil {
				return nil, err
			}

			if ok {
				date := func(c *gitdomain.Commit) time.Time {
					if c.Committer == nil {
						return c.Author.Date
					}
					if c.Committer.Date.After(c.Author.Date) {
						return c.Committer.Date
					}
					return c.Author.Date
				}
				sort.Slice(branches, func(i, j int) bool {
					bi, bj := branches[i], branches[j]
					if bi.Commit == nil {
						return false
					}
					if bj.Commit == nil {
						return true
					}
					di, dj := date(bi.Commit), date(bj.Commit)
					if di.Equal(dj) {
						return bi.Name < bj.Name
					}
					if di.After(dj) {
						return true
					}
					return false
				})
			}
		}
	}

	var tags []*gitdomain.Tag
	if args.Type == nil || *args.Type == gitRefTypeTag {
		var err error
		tags, err = gitserver.NewClient(r.db).ListTags(ctx, r.RepoName())
		if err != nil {
			return nil, err
		}
		if args.OrderBy != nil && *args.OrderBy == gitRefOrderAuthoredOrCommittedAt {
			// Tags are already sorted by creatordate.
		} else {
			// Sort tags by reverse alpha.
			sort.Slice(tags, func(i, j int) bool {
				return tags[i].Name > tags[j].Name
			})
		}
	}

	// Combine branches and tags.
	refs := make([]*GitRefResolver, len(branches)+len(tags))
	for i, b := range branches {
		refs[i] = &GitRefResolver{name: "refs/heads/" + b.Name, repo: r, target: GitObjectID(b.Head)}
	}
	for i, t := range tags {
		refs[i+len(branches)] = &GitRefResolver{name: "refs/tags/" + t.Name, repo: r, target: GitObjectID(t.CommitID)}
	}

	if args.Query != nil {
		query := strings.ToLower(*args.Query)

		// Filter using query.
		filtered := refs[:0]
		for _, ref := range refs {
			if strings.Contains(strings.ToLower(strings.TrimPrefix(ref.name, gitRefPrefix(ref.name))), query) {
				filtered = append(filtered, ref)
			}
		}
		refs = filtered
	}

	return &gitRefConnectionResolver{
		first: args.First,
		refs:  refs,
		repo:  r,
	}, nil
}

func hydrateBranchCommits(ctx context.Context, db database.DB, repo api.RepoName, interactive bool, branches []*gitdomain.Branch) (ok bool, err error) {
	parentCtx := ctx
	if interactive {
		if len(branches) > 1000 {
			return false, nil
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	for _, branch := range branches {
		branch.Commit, err = git.GetCommit(ctx, db, repo, branch.Head, gitserver.ResolveRevisionOptions{}, authz.DefaultSubRepoPermsChecker)
		if err != nil {
			if parentCtx.Err() == nil && ctx.Err() != nil {
				// reached interactive timeout
				return false, nil
			}
			return false, err
		}
	}

	return true, nil
}

type gitRefConnectionResolver struct {
	first *int32
	refs  []*GitRefResolver

	repo *RepositoryResolver
}

func (r *gitRefConnectionResolver) Nodes() []*GitRefResolver {
	var nodes []*GitRefResolver

	// Paginate.
	if r.first != nil && len(r.refs) > int(*r.first) {
		nodes = r.refs[:int(*r.first)]
	} else {
		nodes = r.refs
	}

	return nodes
}

func (r *gitRefConnectionResolver) TotalCount() int32 {
	return int32(len(r.refs))
}

func (r *gitRefConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(r.first != nil && int(*r.first) < len(r.refs))
}
