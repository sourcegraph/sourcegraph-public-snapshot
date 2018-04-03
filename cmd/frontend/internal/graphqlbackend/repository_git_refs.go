package graphqlbackend

import (
	"context"
	"sort"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
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
	vcsrepo := backend.Repos.CachedVCS(r.repo)

	var branches []*vcs.Branch
	if args.Type == nil || *args.Type == gitRefTypeBranch {
		var err error
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
			case !bi.Commit.Author.Date.IsZero():
				return bi.Commit.Author.Date.After(bj.Commit.Author.Date)
			default:
				return bi.Name < bj.Name
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
		// Sort tags by reverse alpha.
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Name > tags[j].Name
		})
	}

	// Combine branches and tags.
	refs := make([]*gitRefResolver, len(branches)+len(tags))
	for i, b := range branches {
		refs[i] = &gitRefResolver{name: "refs/heads/" + b.Name, repo: r, target: gitObjectID(b.Head)}
	}
	for i, t := range tags {
		refs[i+len(branches)] = &gitRefResolver{name: "refs/tags/" + t.Name, repo: r, target: gitObjectID(t.CommitID)}
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

type gitRefConnectionResolver struct {
	first *int32
	refs  []*gitRefResolver

	repo *repositoryResolver
}

func (r *gitRefConnectionResolver) Nodes() []*gitRefResolver {
	var nodes []*gitRefResolver

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

func (r *gitRefConnectionResolver) PageInfo() *pageInfo {
	return &pageInfo{hasNextPage: r.first != nil && int(*r.first) < len(r.refs)}
}
