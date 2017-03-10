package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"strings"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"

	"sourcegraph.com/sourcegraph/sourcegraph/api"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/gobuildserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

var GraphQLSchema *graphql.Schema

func init() {
	var err error
	GraphQLSchema, err = graphql.ParseSchema(api.Schema, &schemaResolver{})
	if err != nil {
		panic(err)
	}
}

type node interface {
	ID() graphql.ID
}

type nodeResolver struct {
	node
}

func (r *nodeResolver) ToRepository() (*repositoryResolver, bool) {
	n, ok := r.node.(*repositoryResolver)
	return n, ok
}

func (r *nodeResolver) ToCommit() (*commitResolver, bool) {
	n, ok := r.node.(*commitResolver)
	return n, ok
}

type schemaResolver struct{}

func (r *schemaResolver) Root() *rootResolver {
	return &rootResolver{}
}

func (r *schemaResolver) Node(ctx context.Context, args *struct{ ID graphql.ID }) (*nodeResolver, error) {
	switch relay.UnmarshalKind(args.ID) {
	case "Repository":
		n, err := repositoryByID(ctx, args.ID)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{n}, nil
	case "Commit":
		n, err := commitByID(ctx, args.ID)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{n}, nil
	default:
		return nil, errors.New("invalid id")
	}
}

type rootResolver struct{}

func (r *rootResolver) Repository(ctx context.Context, args *struct{ URI string }) (*repositoryResolver, error) {
	if args.URI == "" {
		return nil, nil
	}

	repo, err := localstore.Repos.GetByURI(ctx, args.URI)
	if err != nil {
		if err, ok := err.(legacyerr.Error); ok && err.Code == legacyerr.NotFound {
			return nil, nil
		}
		return nil, err
	}

	if err := backend.Repos.RefreshIndex(ctx, repo.URI); err != nil {
		return nil, err
	}

	return &repositoryResolver{repo: repo}, nil
}

func (r *rootResolver) Repositories(ctx context.Context, args *struct{ Query string }) ([]*repositoryResolver, error) {
	opt := &sourcegraph.RepoListOptions{Query: args.Query}
	opt.PerPage = 20
	return listRepos(ctx, opt)
}

func (r *rootResolver) RemoteRepositories(ctx context.Context) ([]*repositoryResolver, error) {
	return listRepos(ctx, &sourcegraph.RepoListOptions{RemoteOnly: true})
}

func listRepos(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*repositoryResolver, error) {
	reposList, err := backend.Repos.List(ctx, opt)

	if err != nil {
		return nil, err
	}

	var l []*repositoryResolver
	for _, repo := range reposList.Repos {
		l = append(l, &repositoryResolver{
			repo: repo,
		})
	}

	return l, nil
}

func (r *rootResolver) RemoteStarredRepositories(ctx context.Context) ([]*repositoryResolver, error) {
	ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "RemoteStarredRepositories")
	repos, err := github.ListStarredRepos(ctx, nil)
	if err != nil {
		return nil, err
	}

	var s []*repositoryResolver
	for _, repo := range repos {
		s = append(s, &repositoryResolver{
			repo: repo,
		})
	}

	return s, nil
}

// Resolves symbols by a global symbol ID (use case for symbol URLs)
func (r *rootResolver) Symbols(ctx context.Context, args *struct {
	ID   string
	Mode string
}) ([]*symbolResolver, error) {

	if args.Mode != "go" {
		return []*symbolResolver{}, nil
	}

	importPath := strings.Split(args.ID, "/-/")[0]
	cloneURL, err := gobuildserver.ResolveImportPathCloneURL(importPath)
	if err != nil {
		return nil, err
	}

	if cloneURL == "" || !strings.HasPrefix(cloneURL, "https://github.com") {
		return nil, fmt.Errorf("non-github clone URL resolved for import path %s", importPath)
	}

	repoURI := strings.TrimPrefix(cloneURL, "https://")
	repo, err := localstore.Repos.GetByURI(ctx, repoURI)
	if err != nil {
		if err, ok := err.(legacyerr.Error); ok && err.Code == legacyerr.NotFound {
			return nil, nil
		}
		return nil, err
	}
	if err := backend.Repos.RefreshIndex(ctx, repoURI); err != nil {
		return nil, err
	}

	// Check that the user has permission to read this repo. Calling
	// Repos.ResolveRev will fail if the user does not have access to the
	// specified repo.
	//
	// SECURITY NOTE: The LSP client proxy DOES NOT check
	// permissions. It accesses the gitserver directly and relies on
	// its callers to check permissions.
	checkedUserHasReadAccessToRepo := false // safeguard to make sure we don't accidentally delete the check below
	var rev *sourcegraph.ResolvedRev
	{
		// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! ResolveRev is responsible for ensuring 🚨
		// the user has permissions to access the repository.
		rev, err = backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
			Repo: repo.ID,
			Rev:  "",
		})
		if err != nil {
			return nil, err
		}
		checkedUserHasReadAccessToRepo = true
	}

	if !checkedUserHasReadAccessToRepo {
		return nil, fmt.Errorf("authorization check failed")
	}

	var symbols []lsp.SymbolInformation
	params := lspext.WorkspaceSymbolParams{Symbol: lspext.SymbolDescriptor{"id": args.ID}}

	err = xlang.UnsafeOneShotClientRequest(ctx, args.Mode, "git://"+repoURI+"?"+rev.CommitID, "workspace/symbol", params, &symbols)
	if err != nil {
		return nil, err
	}

	var resolvers []*symbolResolver
	for _, symbol := range symbols {
		uri, err := uri.Parse(symbol.Location.URI)
		if err != nil {
			return nil, err
		}
		resolvers = append(resolvers, &symbolResolver{
			path:      uri.Fragment,
			line:      int32(symbol.Location.Range.Start.Line),
			character: int32(symbol.Location.Range.Start.Character),
			repo:      repo,
		})
	}

	return resolvers, nil
}

func (r *rootResolver) CurrentUser(ctx context.Context) (*currentUserResolver, error) {
	return currentUser(ctx)
}
