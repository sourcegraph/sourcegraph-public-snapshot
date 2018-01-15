package graphqlbackend

import (
	"context"
	"sort"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func (r *repositoryResolver) Branches(ctx context.Context, args *struct {
	connectionArgs
	Query *string
}) (*gitRefConnectionResolver, error) {
	gitRefTypeBranch := gitRefTypeBranch
	return r.GitRefs(ctx, &struct {
		connectionArgs
		Query *string
		Type  *string
	}{connectionArgs: args.connectionArgs, Query: args.Query, Type: &gitRefTypeBranch})
}

func (r *repositoryResolver) Tags(ctx context.Context, args *struct {
	connectionArgs
	Query *string
}) (*gitRefConnectionResolver, error) {
	gitRefTypeTag := gitRefTypeTag
	return r.GitRefs(ctx, &struct {
		connectionArgs
		Query *string
		Type  *string
	}{connectionArgs: args.connectionArgs, Query: args.Query, Type: &gitRefTypeTag})
}

func (r *repositoryResolver) GitRefs(ctx context.Context, args *struct {
	connectionArgs
	Query *string
	Type  *string
}) (*gitRefConnectionResolver, error) {
	vcsrepo, err := db.RepoVCS.Open(ctx, r.repo.ID)
	if err != nil {
		return nil, err
	}

	var branches []*vcs.Branch
	if args.Type == nil || *args.Type == gitRefTypeBranch {
		branches, err = vcsrepo.Branches(ctx, vcs.BranchesOptions{})
		if err != nil {
			return nil, err
		}

		// Sort branches by most recently committed.
		sort.Slice(branches, func(i, j int) bool {
			bi, bj := branches[i], branches[j]
			switch {
			case bi.Commit == nil:
				return false
			case bj.Commit == nil:
				return true
			default:
				return bi.Commit.Author.Date.After(bj.Commit.Author.Date)
			}
		})
	}

	var tags []*vcs.Tag
	if args.Type == nil || *args.Type == gitRefTypeTag {
		var err error
		tags, err = vcsrepo.Tags(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &gitRefConnectionResolver{
		first:    args.First,
		query:    args.Query,
		branches: branches,
		tags:     tags,
		repo:     r,
	}, nil
}

type gitRefConnectionResolver struct {
	first    *int32
	query    *string
	branches []*vcs.Branch
	tags     []*vcs.Tag

	repo *repositoryResolver
}

func (r *gitRefConnectionResolver) Nodes() []*gitRefResolver {
	// Combine branches and tags.
	refs := make([]*gitRefResolver, len(r.branches)+len(r.tags))
	for i, b := range r.branches {
		refs[i] = &gitRefResolver{name: "refs/heads/" + b.Name, repo: r.repo, target: gitObjectID(b.Head)}
	}
	for i, t := range r.tags {
		refs[i+len(r.branches)] = &gitRefResolver{name: "refs/tags/" + t.Name, repo: r.repo, target: gitObjectID(t.CommitID)}
	}

	if r.query != nil {
		query := strings.ToLower(*r.query)

		// Filter using query.
		filtered := refs[:0]
		for _, ref := range refs {
			if strings.Contains(strings.ToLower(strings.TrimPrefix(ref.name, gitRefPrefix(ref.name))), query) {
				filtered = append(filtered, ref)
			}
		}
		refs = filtered
	}

	// Paginate.
	if r.first != nil && len(refs) > int(*r.first) {
		refs = refs[:int(*r.first)]
	}

	return refs
}

func (r *gitRefConnectionResolver) TotalCount() int32 {
	return int32(len(r.Nodes()))
}
