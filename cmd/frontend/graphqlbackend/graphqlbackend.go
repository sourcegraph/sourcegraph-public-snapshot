package graphqlbackend

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-go/trace"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
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

func (r *nodeResolver) ToAccessToken() (*accessTokenResolver, bool) {
	n, ok := r.node.(*accessTokenResolver)
	return n, ok
}

func (r *nodeResolver) ToDiscussionComment() (*discussionCommentResolver, bool) {
	n, ok := r.node.(*discussionCommentResolver)
	return n, ok
}

func (r *nodeResolver) ToDiscussionThread() (*discussionThreadResolver, bool) {
	n, ok := r.node.(*discussionThreadResolver)
	return n, ok
}

func (r *nodeResolver) ToProductLicense() (ProductLicense, bool) {
	n, ok := r.node.(ProductLicense)
	return n, ok
}

func (r *nodeResolver) ToProductSubscription() (ProductSubscription, bool) {
	n, ok := r.node.(ProductSubscription)
	return n, ok
}

func (r *nodeResolver) ToExternalAccount() (*externalAccountResolver, bool) {
	n, ok := r.node.(*externalAccountResolver)
	return n, ok
}

func (r *nodeResolver) ToExternalService() (*externalServiceResolver, bool) {
	n, ok := r.node.(*externalServiceResolver)
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

func (r *nodeResolver) ToUser() (*UserResolver, bool) {
	n, ok := r.node.(*UserResolver)
	return n, ok
}

func (r *nodeResolver) ToOrg() (*OrgResolver, bool) {
	n, ok := r.node.(*OrgResolver)
	return n, ok
}

func (r *nodeResolver) ToOrganizationInvitation() (*organizationInvitationResolver, bool) {
	n, ok := r.node.(*organizationInvitationResolver)
	return n, ok
}

func (r *nodeResolver) ToGitCommit() (*gitCommitResolver, bool) {
	n, ok := r.node.(*gitCommitResolver)
	return n, ok
}

func (r *nodeResolver) ToRegistryExtension() (RegistryExtension, bool) {
	if NodeToRegistryExtension == nil {
		return nil, false
	}
	return NodeToRegistryExtension(r.node)
}

func (r *nodeResolver) ToSavedSearch() (*savedSearchResolver, bool) {
	n, ok := r.node.(*savedSearchResolver)
	return n, ok
}

func (r *nodeResolver) ToSite() (*siteResolver, bool) {
	n, ok := r.node.(*siteResolver)
	return n, ok
}

// stringLogger describes something that can log strings, list them and also
// clean up to make sure they don't use too much storage space.
type stringLogger interface {
	// Log stores the given string s.
	Log(ctx context.Context, s string) error

	// Top returns the top n most frequently occurring strings.
	// The returns are parallel slices for the unique strings and their associated counts.
	Top(ctx context.Context, n int32) ([]string, []int32, error)

	// Cleanup removes old entries such that there are no more than limit remaining.
	Cleanup(ctx context.Context, limit int) error
}

// schemaResolver handles all GraphQL queries for Sourcegraph.  To do this, it
// uses subresolvers, some of which are globals and some of which are fields on
// schemaResolver.
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
	case "AccessToken":
		return accessTokenByID(ctx, id)
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
		return orgByID(ctx, id)
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
	default:
		return nil, errors.New("invalid id")
	}
}

func (r *schemaResolver) Repository(ctx context.Context, args *struct {
	Name     *string
	CloneURL *string
	// TODO(chris): Remove URI in favor of Name.
	URI *string
}) (*repositoryResolver, error) {
	var name api.RepoName
	if args.URI != nil {
		// Deprecated query by "URI"
		name = api.RepoName(*args.URI)
	} else if args.Name != nil {
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
		return nil, errors.New("Neither name nor cloneURL given")
	}

	repo, err := backend.Repos.GetByName(ctx, name)
	if err != nil {
		if err, ok := err.(backend.ErrRepoSeeOther); ok {
			return &repositoryResolver{repo: &types.Repo{}, redirectURL: &err.RedirectURL}, nil
		}
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &repositoryResolver{repo: repo}, nil
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
