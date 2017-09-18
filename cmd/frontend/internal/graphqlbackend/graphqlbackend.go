package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/clearbit/clearbit-go/clearbit"
	graphql "github.com/neelance/graphql-go"
	gqlerrors "github.com/neelance/graphql-go/errors"
	"github.com/neelance/graphql-go/relay"
	"github.com/neelance/graphql-go/trace"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/clearbit"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/gobuildserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

var GraphQLSchema *graphql.Schema

var graphqlFieldHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "graphql",
	Name:      "field_seconds",
	Help:      "GraphQL field resolver latencies in seconds.",
	Buckets:   []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
}, []string{"type", "field", "error"})

func init() {
	prometheus.MustRegister(graphqlFieldHistogram)
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

// Various graphql tools expect at least one field to be present in the schema
// so we provide a dummy one here that is always nil.
func (er *EmptyResponse) AlwaysNil() *string {
	return nil
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

	if err := refreshRepo(ctx, repo); err != nil {
		return nil, err
	}

	return &repositoryResolver{repo: repo}, nil
}

var skipRefresh = false // set by tests

func refreshRepo(ctx context.Context, repo *sourcegraph.Repo) error {
	if skipRefresh {
		return nil
	}

	go func() {
		if err := localstore.Repos.UpdateRepoFieldsFromRemote(context.Background(), repo.ID); err != nil {
			log.Printf("failed to update repo %s from remote: %s", repo.URI, err)
		}
	}()

	if err := backend.Repos.RefreshIndex(ctx, repo.URI); err != nil {
		return err
	}

	return nil
}

func (r *rootResolver) Repositories(ctx context.Context, args *struct {
	Query string
}) ([]*repositoryResolver, error) {
	opt := &sourcegraph.RepoListOptions{Query: args.Query}
	opt.PerPage = 200
	return listRepos(ctx, opt)
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

	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! ResolveRev is responsible for ensuring 🚨
	// the user has permissions to access the repository.
	rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: repo.ID,
		Rev:  "",
	})

	var symbols []lsp.SymbolInformation
	params := lspext.WorkspaceSymbolParams{Symbol: lspext.SymbolDescriptor{"id": args.ID}}

	err = xlang.UnsafeOneShotClientRequest(ctx, args.Mode, lsp.DocumentURI("git://"+repoURI+"?"+rev.CommitID), "workspace/symbol", params, &symbols)
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

func (r *rootResolver) CurrentUser(ctx context.Context) (*currentUserResolver, error) {
	return currentUser(ctx)
}

// RevealCustomerCompany transforms a user's IP addresses into a company profile by using
// Clearbit's reveal API.
func (r *rootResolver) RevealCustomerCompany(ctx context.Context, args *struct{ IP string }) (*revealResolver, error) {
	c, err := clearbitutil.NewClient()
	if err != nil {
		return nil, err
	}

	reveal, _, err := c.Reveal.Find(clearbit.RevealFindParams{
		IP: args.IP,
	})
	if err != nil {
		return nil, err
	}

	return &revealResolver{
		ip:     reveal.IP,
		domain: reveal.Domain,
		fuzzy:  reveal.Fuzzy,
		company: &companyResolver{
			id:            reveal.Company.ID,
			name:          reveal.Company.Name,
			legalName:     reveal.Company.LegalName,
			domain:        reveal.Company.Domain,
			domainAliases: reveal.Company.DomainAliases,
			url:           reveal.Company.URL,
			site: &siteDetailsResolver{
				url:            reveal.Company.Site.URL,
				title:          reveal.Company.Site.Title,
				phoneNumbers:   reveal.Company.Site.PhoneNumbers,
				emailAddresses: reveal.Company.Site.EmailAddresses,
			},
			category: &companyCategoryResolver{
				sector:        reveal.Company.Category.Sector,
				industryGroup: reveal.Company.Category.IndustryGroup,
				industry:      reveal.Company.Category.Industry,
				subIndustry:   reveal.Company.Category.SubIndustry,
			},
			tags:        reveal.Company.Tags,
			description: reveal.Company.Description,
			foundedYear: string(reveal.Company.FoundedYear),
			location:    reveal.Company.Location,
			logo:        reveal.Company.Logo,
			tech:        reveal.Company.Tech,
		},
	}, nil
}
