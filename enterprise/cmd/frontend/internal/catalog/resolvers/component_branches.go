package resolvers

import (
	"context"
	"log"
	"sort"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func (r *componentResolver) Branches(ctx context.Context, args *graphqlutil.ConnectionArgs) (gql.GitRefConnectionResolver, error) {
	// TODO(sqs): This will miss branches whose commit is very old. Instead, the impl should be
	// changed to iterate through all branches and find commits that affect the component's files.

	limit := int(args.GetFirst())

	slocs, err := r.sourceLocations(ctx)
	if err != nil {
		return nil, err
	}

	var allRefs []*gql.GitRefResolver
	for _, sloc := range slocs {
		matches, err := getBranchesForRepo(ctx, sloc.repoName, joinPathPrefixRegexps(sloc.paths), limit)
		if err != nil {
			return nil, err
		}

		ignoreRef := func(refName string) bool {
			return strings.HasPrefix(refName, "refs/heads/main-dry-run/")
		}

		type refData struct {
			*gql.GitRefResolver
			index int // support sorting by the original order returned from `git log`
		}
		refsByName := map[string]refData{}
		for i, m := range matches {
			for _, sourceRef := range m.SourceRefs {
				if ignoreRef(sourceRef) {
					continue
				}
				if _, ok := refsByName[sourceRef]; !ok {
					refsByName[sourceRef] = refData{
						GitRefResolver: gql.NewGitRefResolver(sloc.repo, sourceRef, gql.GitObjectID(m.Oid)),
						index:          i,
					}
				}
			}
		}

		refs := make([]*gql.GitRefResolver, 0, len(refsByName))
		for _, refResolver := range refsByName {
			refs = append(refs, refResolver.GitRefResolver)
		}
		sort.Slice(refs, func(i, j int) bool {
			return refsByName[refs[i].Name()].index < refsByName[refs[j].Name()].index
		})
		allRefs = append(allRefs, refs...)
	}

	return gql.NewGitRefConnectionResolver(args.First, allRefs), nil
}

func getBranchesForRepo(ctx context.Context, repo api.RepoName, pathRegexp string, limit int) ([]protocol.CommitMatch, error) {
	req := &protocol.SearchRequest{
		Repo: repo,
		Revisions: []protocol.RevisionSpecifier{
			{RevSpec: "^HEAD"},
			{RefGlob: "refs/heads/"},
		},
		Limit: limit,
		Query: &protocol.DiffModifiesFile{Expr: pathRegexp},
	}

	var allMatches []protocol.CommitMatch
	onMatches := func(matches []protocol.CommitMatch) {
		allMatches = append(allMatches, matches...)
	}
	limitHit, err := gitserver.DefaultClient.Search(ctx, req, onMatches)
	if err != nil {
		return nil, err
	}
	if limitHit {
		log.Println("TODO(sqs): limitHit")
	}
	return allMatches, nil
}
