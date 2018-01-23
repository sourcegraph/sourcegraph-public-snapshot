package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	graphql "github.com/neelance/graphql-go"
	gqlerrors "github.com/neelance/graphql-go/errors"
	"github.com/neelance/graphql-go/relay"
	"github.com/neelance/graphql-go/trace"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/gobuildserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

// GraphQLSchema is the parsed Schema with the root resolver attached. It is
// exported since it is accessed in our httpapi.
var GraphQLSchema *graphql.Schema

var graphqlFieldHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "graphql",
	Name:      "field_seconds",
	Help:      "GraphQL field resolver latencies in seconds.",
	Buckets:   []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
}, []string{"type", "field", "error"})

// githubEnterpriseURLs is a map of GitHub Enterprise hosts to their full URLs.
// This is used for the purposes of generating external GitHub enterprise links.
var githubEnterpriseURLs = make(map[string]string)
var repoListConfigs = make(map[api.RepoURI]schema.Repository)

func init() {
	prometheus.MustRegister(graphqlFieldHistogram)
	githubConf := conf.Get().Github
	for _, c := range githubConf {
		gheURL, err := url.Parse(c.Url)
		if err != nil {
			log15.Error("error parsing GitHub config", "error", err)
		}
		githubEnterpriseURLs[gheURL.Host] = strings.TrimSuffix(c.Url, "/")
	}
	reposList := conf.Get().ReposList
	for _, r := range reposList {
		repoListConfigs[api.RepoURI(r.Path)] = r
	}
}

type prometheusTracer struct {
	trace.OpenTracingTracer
}

func (prometheusTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, trace.TraceFieldFinishFunc) {
	traceCtx, finish := trace.OpenTracingTracer{}.TraceField(ctx, label, typeName, fieldName, trivial, args)
	start := time.Now()
	return traceCtx, func(err *gqlerrors.QueryError) {
		graphqlFieldHistogram.WithLabelValues(typeName, fieldName, strconv.FormatBool(err != nil)).Observe(time.Since(start).Seconds())
		finish(err)
	}
}

func init() {
	var err error
	GraphQLSchema, err = graphql.ParseSchema(Schema, &schemaResolver{}, graphql.Tracer(prometheusTracer{}))
	if err != nil {
		panic(err)
	}
}

// EmptyResponse is a type that can be used in the return signature for graphql queries
// that don't require a return value.
type EmptyResponse struct{}

// AlwaysNil exists since various graphql tools expect at least one field to be
// present in the schema so we provide a dummy one here that is always nil.
func (er *EmptyResponse) AlwaysNil() *string {
	return nil
}

type node interface {
	ID() graphql.ID
}

type nodeResolver struct {
	node
}

func (r *nodeResolver) ToComment() (*commentResolver, bool) {
	n, ok := r.node.(*commentResolver)
	return n, ok
}

func (r *nodeResolver) ToGitRef() (*gitRefResolver, bool) {
	n, ok := r.node.(*gitRefResolver)
	return n, ok
}

func (r *nodeResolver) ToRepository() (*repositoryResolver, bool) {
	n, ok := r.node.(*repositoryResolver)
	return n, ok
}

func (r *nodeResolver) ToUser() (*userResolver, bool) {
	n, ok := r.node.(*userResolver)
	return n, ok
}

func (r *nodeResolver) ToOrg() (*orgResolver, bool) {
	n, ok := r.node.(*orgResolver)
	return n, ok
}

func (r *nodeResolver) ToGitCommit() (*gitCommitResolver, bool) {
	n, ok := r.node.(*gitCommitResolver)
	return n, ok
}

func (r *nodeResolver) ToSite() (*siteResolver, bool) {
	n, ok := r.node.(*siteResolver)
	return n, ok
}

func (r *nodeResolver) ToThread() (*threadResolver, bool) {
	n, ok := r.node.(*threadResolver)
	return n, ok
}

type schemaResolver struct{}

// DEPRECATED
func (r *schemaResolver) Root() *schemaResolver {
	return &schemaResolver{}
}

func (r *schemaResolver) Node(ctx context.Context, args *struct{ ID graphql.ID }) (*nodeResolver, error) {
	n, err := nodeByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return &nodeResolver{n}, nil
}

func nodeByID(ctx context.Context, id graphql.ID) (node, error) {
	switch relay.UnmarshalKind(id) {
	case "GitRef":
		return gitRefByID(ctx, id)
	case "Comment":
		return commentByID(ctx, id)
	case "Repository":
		return repositoryByID(ctx, id)
	case "User":
		return userByID(ctx, id)
	case "Org":
		return orgByID(ctx, id)
	case "GitCommit":
		return gitCommitByID(ctx, id)
	case "SavedQuery":
		return savedQueryByID(ctx, id)
	case "Site":
		return siteByGQLID(ctx, id)
	case "Thread":
		return threadByID(ctx, id)
	default:
		return nil, errors.New("invalid id")
	}
}

func (r *schemaResolver) Repository(ctx context.Context, args *struct{ URI string }) (*repositoryResolver, error) {
	if args.URI == "" {
		return nil, nil
	}

	repo, err := backend.Repos.GetByURI(ctx, api.RepoURI(args.URI))
	if err != nil {
		if err, ok := err.(backend.ErrRepoSeeOther); ok {
			return &repositoryResolver{repo: &types.Repo{}, redirectURL: &err.RedirectURL}, nil
		}
		if err, ok := err.(legacyerr.Error); ok && err.Code == legacyerr.NotFound {
			return nil, nil
		}
		return nil, err
	}

	if err := refreshRepo(ctx, repo); err != nil {
		return nil, err
	}

	return &repositoryResolver{repo: repo}, nil
}

func (r *schemaResolver) PhabricatorRepo(ctx context.Context, args *struct{ URI string }) (*phabricatorRepoResolver, error) {
	repo, err := db.Phabricator.GetByURI(ctx, api.RepoURI(args.URI))
	if err != nil {
		return nil, err
	}
	return &phabricatorRepoResolver{repo}, nil
}

var skipRefresh = false // set by tests

func refreshRepo(ctx context.Context, repo *types.Repo) error {
	if skipRefresh {
		return nil
	}
	return backend.Repos.RefreshIndex(ctx, repo.URI)
}

// Resolves symbols by a global symbol ID (use case for symbol URLs)
func (r *schemaResolver) Symbols(ctx context.Context, args *struct {
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

	repoURI := api.RepoURI(strings.TrimPrefix(cloneURL, "https://"))
	repo, err := db.Repos.GetByURI(ctx, repoURI)
	if err != nil {
		if err, ok := err.(legacyerr.Error); ok && err.Code == legacyerr.NotFound {
			return nil, nil
		}
		return nil, err
	}
	if err := backend.Repos.RefreshIndex(ctx, repoURI); err != nil {
		return nil, err
	}

	rev, err := backend.Repos.ResolveRev(ctx, repo.ID, "")
	if err != nil {
		return nil, err
	}

	var symbols []lsp.SymbolInformation
	params := lspext.WorkspaceSymbolParams{Symbol: lspext.SymbolDescriptor{"id": args.ID}}

	err = xlang.UnsafeOneShotClientRequest(ctx, args.Mode, lsp.DocumentURI("git://"+string(repoURI)+"?"+string(rev)), "workspace/symbol", params, &symbols)
	if err != nil {
		return nil, err
	}

	var resolvers []*symbolResolver
	for _, symbol := range symbols {
		uri, err := uri.Parse(string(symbol.Location.URI))
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

func (r *schemaResolver) CurrentUser(ctx context.Context) (*userResolver, error) {
	return currentUser(ctx)
}

func (r *schemaResolver) Packages(ctx context.Context, args *struct {
	Lang    string
	ID      *string
	Type    *string
	Name    *string
	Commit  *string
	BaseDir *string
	RepoURL *string
	Version *string
	Offset  *int32
	Limit   *int32
}) ([]*packageResolver, error) {
	var limit int32 = 10
	if args.Limit != nil {
		limit = *args.Limit
	}
	if limit > 100 {
		limit = 100
	}

	pkgQuery := packageMetadata{
		id:      args.ID,
		typ:     args.Type,
		name:    args.Name,
		commit:  args.Commit,
		baseDir: args.BaseDir,
		repoURL: args.RepoURL,
		version: args.Version,
	}.toPkgQuery()

	pkgs, err := backend.Pkgs.ListPackages(ctx, &api.ListPackagesOp{Lang: args.Lang, PkgQuery: pkgQuery, Limit: int(limit)})
	if err != nil {
		return nil, err
	}
	pkgResolvers := make([]*packageResolver, len(pkgs))
	for i, pkg := range pkgs {
		pkgResolvers[i] = &packageResolver{&pkg}
	}
	return pkgResolvers, nil
}

func (r *schemaResolver) Dependents(ctx context.Context, args *struct {
	Lang    string
	ID      *string
	Type    *string
	Name    *string
	Commit  *string
	BaseDir *string
	RepoURL *string
	Version *string
	Package *string
	Limit   *int32
}) ([]*dependencyResolver, error) {
	limit := int32(10)
	if args.Limit != nil {
		limit = *args.Limit
	}
	if limit > 100 {
		limit = 100
	}

	pkgQuery := packageMetadata{
		id:      args.ID,
		typ:     args.Type,
		name:    args.Name,
		commit:  args.Commit,
		baseDir: args.BaseDir,
		repoURL: args.RepoURL,
		version: args.Version,
		packag:  args.Package,
	}.toPkgQuery()

	deps, err := db.GlobalDeps.Dependencies(ctx, db.DependenciesOptions{Language: args.Lang, DepData: pkgQuery, Limit: int(limit)})
	if err != nil {
		return nil, err
	}

	depResolvers := make([]*dependencyResolver, len(deps))
	for i, dep := range deps {
		depResolvers[i] = &dependencyResolver{dep}
	}

	return depResolvers, nil
}

func (r *schemaResolver) UpdateDeploymentConfiguration(ctx context.Context, args *struct {
	Email           string
	EnableTelemetry bool
}) (*EmptyResponse, error) {
	configuration := &types.SiteConfig{
		Email:            args.Email,
		TelemetryEnabled: args.EnableTelemetry,
	}
	err := db.SiteConfig.UpdateConfiguration(ctx, configuration)
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
