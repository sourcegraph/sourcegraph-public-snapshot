package graphqlbackend

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-go/trace"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var graphqlFieldHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "graphql",
	Name:      "field_seconds",
	Help:      "GraphQL field resolver latencies in seconds.",
	Buckets:   []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
}, []string{"type", "field", "error"})

var codeIntelSearchHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "graphql",
	Name:      "code_intel_search_seconds",
	Help:      "Code intel search latencies in seconds.",
	Buckets:   []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
}, []string{"exact", "error"})

func init() {
	prometheus.MustRegister(graphqlFieldHistogram)
	prometheus.MustRegister(codeIntelSearchHistogram)
}

type prometheusTracer struct {
	trace.OpenTracingTracer
}

func (prometheusTracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, trace.TraceQueryFinishFunc) {
	traceCtx, finish := trace.OpenTracingTracer{}.TraceQuery(ctx, queryString, operationName, variables, varTypes)

	// Note: We don't care about the error here, we just extract the username if
	// we get a non-nil user object.
	currentUser, _ := CurrentUser(ctx)
	var currentUserName string
	if currentUser != nil {
		currentUserName = currentUser.Username()
	}

	// Requests made by our JS frontend and other internal things will have a concrete name attached to the
	// request which allows us to (softly) differentiate it from end-user API requests. For example,
	// /.api/graphql?Foobar where Foobar is the name of the request we make. If there is not a request name,
	// then it is an interesting query to log in the event it is harmful and a site admin needs to identify
	// it and the user issuing it.
	requestName := sgtrace.GraphQLRequestName(ctx)
	lvl := log15.Debug
	if requestName == "unknown" {
		lvl = log15.Info
	}
	lvl("serving GraphQL request", "name", requestName, "user", currentUserName)
	if requestName == "unknown" {
		log.Printf(`logging complete query for unnamed GraphQL request above name=%s user=%s:
QUERY
-----
%s

VARIABLES
---------
%v

`, requestName, currentUserName, queryString, variables)
	}
	return traceCtx, finish
}

func (prometheusTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, trace.TraceFieldFinishFunc) {
	traceCtx, finish := trace.OpenTracingTracer{}.TraceField(ctx, label, typeName, fieldName, trivial, args)
	start := time.Now()
	return traceCtx, func(err *gqlerrors.QueryError) {
		isErrStr := strconv.FormatBool(err != nil)
		graphqlFieldHistogram.WithLabelValues(typeName, fieldName, isErrStr).Observe(time.Since(start).Seconds())

		origin := sgtrace.RequestOrigin(ctx)
		if origin != "unknown" && (fieldName == "search" || fieldName == "lsif") {
			isExact := strconv.FormatBool(fieldName == "lsif")
			codeIntelSearchHistogram.WithLabelValues(isExact, isErrStr).Observe(time.Since(start).Seconds())
		}
		finish(err)
	}
}

func NewSchema(a8n A8NResolver, codeIntel CodeIntelResolver, authz AuthzResolver) (*graphql.Schema, error) {
	resolver := &schemaResolver{
		A8NResolver:       defaultA8NResolver{},
		AuthzResolver:     defaultAuthzResolver{},
		CodeIntelResolver: defaultCodeIntelResolver{},
	}
	if a8n != nil {
		resolver.A8NResolver = a8n
	}
	if codeIntel != nil {
		EnterpriseResolvers.codeIntelResolver = codeIntel
		resolver.CodeIntelResolver = codeIntel
	}
	if authz != nil {
		EnterpriseResolvers.authzResolver = authz
		resolver.AuthzResolver = authz
	}

	return graphql.ParseSchema(
		Schema,
		resolver,
		graphql.Tracer(prometheusTracer{}),
	)
}

// EmptyResponse is a type that can be used in the return signature for graphql queries
// that don't require a return value.
type EmptyResponse struct{}

// AlwaysNil exists since various graphql tools expect at least one field to be
// present in the schema so we provide a dummy one here that is always nil.
func (er *EmptyResponse) AlwaysNil() *string {
	return nil
}

type Node interface {
	ID() graphql.ID
}

type NodeResolver struct {
	Node
}

func (r *NodeResolver) ToAccessToken() (*accessTokenResolver, bool) {
	n, ok := r.Node.(*accessTokenResolver)
	return n, ok
}

func (r *NodeResolver) ToCampaign() (CampaignResolver, bool) {
	n, ok := r.Node.(CampaignResolver)
	return n, ok
}

func (r *NodeResolver) ToCampaignPlan() (CampaignPlanResolver, bool) {
	n, ok := r.Node.(CampaignPlanResolver)
	return n, ok
}

func (r *NodeResolver) ToExternalChangeset() (ExternalChangesetResolver, bool) {
	n, ok := r.Node.(ExternalChangesetResolver)
	return n, ok
}

func (r *NodeResolver) ToChangesetEvent() (ChangesetEventResolver, bool) {
	n, ok := r.Node.(ChangesetEventResolver)
	return n, ok
}

func (r *NodeResolver) ToDiscussionComment() (*discussionCommentResolver, bool) {
	n, ok := r.Node.(*discussionCommentResolver)
	return n, ok
}

func (r *NodeResolver) ToDiscussionThread() (*discussionThreadResolver, bool) {
	n, ok := r.Node.(*discussionThreadResolver)
	return n, ok
}

func (r *NodeResolver) ToProductLicense() (ProductLicense, bool) {
	n, ok := r.Node.(ProductLicense)
	return n, ok
}

func (r *NodeResolver) ToProductSubscription() (ProductSubscription, bool) {
	n, ok := r.Node.(ProductSubscription)
	return n, ok
}

func (r *NodeResolver) ToExternalAccount() (*externalAccountResolver, bool) {
	n, ok := r.Node.(*externalAccountResolver)
	return n, ok
}

func (r *NodeResolver) ToExternalService() (*externalServiceResolver, bool) {
	n, ok := r.Node.(*externalServiceResolver)
	return n, ok
}

func (r *NodeResolver) ToGitRef() (*GitRefResolver, bool) {
	n, ok := r.Node.(*GitRefResolver)
	return n, ok
}

func (r *NodeResolver) ToRepository() (*RepositoryResolver, bool) {
	n, ok := r.Node.(*RepositoryResolver)
	return n, ok
}

func (r *NodeResolver) ToUser() (*UserResolver, bool) {
	n, ok := r.Node.(*UserResolver)
	return n, ok
}

func (r *NodeResolver) ToOrg() (*OrgResolver, bool) {
	n, ok := r.Node.(*OrgResolver)
	return n, ok
}

func (r *NodeResolver) ToOrganizationInvitation() (*organizationInvitationResolver, bool) {
	n, ok := r.Node.(*organizationInvitationResolver)
	return n, ok
}

func (r *NodeResolver) ToGitCommit() (*GitCommitResolver, bool) {
	n, ok := r.Node.(*GitCommitResolver)
	return n, ok
}

func (r *NodeResolver) ToRegistryExtension() (RegistryExtension, bool) {
	if NodeToRegistryExtension == nil {
		return nil, false
	}
	return NodeToRegistryExtension(r.Node)
}

func (r *NodeResolver) ToSavedSearch() (*savedSearchResolver, bool) {
	n, ok := r.Node.(*savedSearchResolver)
	return n, ok
}

func (r *NodeResolver) ToSite() (*siteResolver, bool) {
	n, ok := r.Node.(*siteResolver)
	return n, ok
}

func (r *NodeResolver) ToLSIFUpload() (LSIFUploadResolver, bool) {
	n, ok := r.Node.(LSIFUploadResolver)
	return n, ok
}

// schemaResolver handles all GraphQL queries for Sourcegraph. To do this, it
// uses subresolvers which are globals. Enterprise-only resolvers are assigned
// to a field of EnterpriseResolvers.
type schemaResolver struct {
	A8NResolver
	AuthzResolver
	CodeIntelResolver
}

// EnterpriseResolvers holds the instances of resolvers which are enabled only
// in enterprise mode. These resolver instances are nil when running as OSS.
var EnterpriseResolvers = struct {
	codeIntelResolver CodeIntelResolver
	authzResolver     AuthzResolver
}{
	codeIntelResolver: defaultCodeIntelResolver{},
	authzResolver:     defaultAuthzResolver{},
}

// DEPRECATED
func (r *schemaResolver) Root() *schemaResolver {
	return &schemaResolver{}
}

func (r *schemaResolver) Node(ctx context.Context, args *struct{ ID graphql.ID }) (*NodeResolver, error) {
	n, err := r.nodeByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return &NodeResolver{n}, nil
}

func (r *schemaResolver) nodeByID(ctx context.Context, id graphql.ID) (Node, error) {
	switch relay.UnmarshalKind(id) {
	case "AccessToken":
		return accessTokenByID(ctx, id)
	case "Campaign":
		return r.CampaignByID(ctx, id)
	case "CampaignPlan":
		return r.CampaignPlanByID(ctx, id)
	case "ExternalChangeset":
		return r.ChangesetByID(ctx, id)
	case "ChangesetPlan":
		return r.ChangesetPlanByID(ctx, id)
	case "DiscussionComment":
		return discussionCommentByID(ctx, id)
	case "DiscussionThread":
		return discussionThreadByID(ctx, id)
	case "ProductLicense":
		if f := ProductLicenseByID; f != nil {
			return f(ctx, id)
		}
		return nil, errors.New("not implemented")
	case "ProductSubscription":
		if f := ProductSubscriptionByID; f != nil {
			return f(ctx, id)
		}
		return nil, errors.New("not implemented")
	case "ExternalAccount":
		return externalAccountByID(ctx, id)
	case externalServiceIDKind:
		return externalServiceByID(ctx, id)
	case "GitRef":
		return gitRefByID(ctx, id)
	case "Repository":
		return repositoryByID(ctx, id)
	case "User":
		return UserByID(ctx, id)
	case "Org":
		return OrgByID(ctx, id)
	case "OrganizationInvitation":
		return orgInvitationByID(ctx, id)
	case "GitCommit":
		return gitCommitByID(ctx, id)
	case "RegistryExtension":
		return RegistryExtensionByID(ctx, id)
	case "SavedSearch":
		return savedSearchByID(ctx, id)
	case "Site":
		return siteByGQLID(ctx, id)
	case "LSIFUpload":
		return r.LSIFUploadByID(ctx, id)
	default:
		return nil, errors.New("invalid id")
	}
}

func (r *schemaResolver) Repository(ctx context.Context, args *struct {
	Name     *string
	CloneURL *string
	// TODO(chris): Remove URI in favor of Name.
	URI *string
}) (*RepositoryResolver, error) {
	// Deprecated query by "URI"
	if args.URI != nil && args.Name == nil {
		args.Name = args.URI
	}
	resolver, err := r.RepositoryRedirect(ctx, &struct {
		Name     *string
		CloneURL *string
	}{args.Name, args.CloneURL})
	if err != nil {
		return nil, err
	}
	return resolver.repo, nil
}

type RedirectResolver struct {
	url string
}

func (r *RedirectResolver) URL() string {
	return r.url
}

type repositoryRedirect struct {
	repo     *RepositoryResolver
	redirect *RedirectResolver
}

func (r *repositoryRedirect) ToRepository() (*RepositoryResolver, bool) {
	return r.repo, r.repo != nil
}

func (r *repositoryRedirect) ToRedirect() (*RedirectResolver, bool) {
	return r.redirect, r.redirect != nil
}

func (r *schemaResolver) RepositoryRedirect(ctx context.Context, args *struct {
	Name     *string
	CloneURL *string
}) (*repositoryRedirect, error) {
	var name api.RepoName
	if args.Name != nil {
		// Query by name
		name = api.RepoName(*args.Name)
	} else if args.CloneURL != nil {
		// Query by git clone URL
		var err error
		name, err = reposourceCloneURLToRepoName(ctx, *args.CloneURL)
		if err != nil {
			return nil, err
		}
		if name == "" {
			// Clone URL could not be mapped to a code host
			return nil, nil
		}
	} else {
		return nil, errors.New("neither name nor cloneURL given")
	}

	repo, err := backend.Repos.GetByName(ctx, name)
	if err != nil {
		if err, ok := err.(backend.ErrRepoSeeOther); ok {
			return &repositoryRedirect{redirect: &RedirectResolver{url: err.RedirectURL}}, nil
		}
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &repositoryRedirect{repo: &RepositoryResolver{repo: repo}}, nil
}

func (r *schemaResolver) PhabricatorRepo(ctx context.Context, args *struct {
	Name *string
	// TODO(chris): Remove URI in favor of Name.
	URI *string
}) (*phabricatorRepoResolver, error) {
	if args.Name != nil {
		args.URI = args.Name
	}

	repo, err := db.Phabricator.GetByName(ctx, api.RepoName(*args.URI))
	if err != nil {
		return nil, err
	}
	return &phabricatorRepoResolver{repo}, nil
}

func (r *schemaResolver) CurrentUser(ctx context.Context) (*UserResolver, error) {
	return CurrentUser(ctx)
}
