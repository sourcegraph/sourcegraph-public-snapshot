package graphqlbackend

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-go/trace"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cloneurls"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	graphqlFieldHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_graphql_field_seconds",
		Help:    "GraphQL field resolver latencies in seconds.",
		Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
	}, []string{"type", "field", "error", "source", "request_name"})

	codeIntelSearchHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_graphql_code_intel_search_seconds",
		Help:    "Code intel search latencies in seconds.",
		Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
	}, []string{"exact", "error"})
)

type prometheusTracer struct {
	tracer trace.OpenTracingTracer
}

func (t *prometheusTracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]any, varTypes map[string]*introspection.Type) (context.Context, trace.TraceQueryFinishFunc) {
	start := time.Now()
	var finish trace.TraceQueryFinishFunc
	if policy.ShouldTrace(ctx) {
		ctx, finish = t.tracer.TraceQuery(ctx, queryString, operationName, variables, varTypes)
	}

	ctx = context.WithValue(ctx, sgtrace.GraphQLQueryKey, queryString)

	_, disableLog := os.LookupEnv("NO_GRAPHQL_LOG")
	_, logAllRequests := os.LookupEnv("LOG_ALL_GRAPHQL_REQUESTS")

	// Note: We don't care about the error here, we just extract the username if
	// we get a non-nil user object.
	var currentUserID int32
	a := actor.FromContext(ctx)
	if a.IsAuthenticated() {
		currentUserID = a.UID
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
	requestSource := sgtrace.RequestSource(ctx)

	if !disableLog {
		lvl("serving GraphQL request", "name", requestName, "userID", currentUserID, "source", requestSource)
		if requestName == "unknown" || logAllRequests {
			reason := ""
			if requestName == "unknown" {
				reason = "for unnamed GraphQL request above "
			}
			log.Printf(`logging complete query %sname=%s userID=%d source=%s:
QUERY
-----
%s

VARIABLES
---------
%v

`, reason, requestName, currentUserID, requestSource, queryString, variables)
		}
	}

	return ctx, func(err []*gqlerrors.QueryError) {
		if finish != nil {
			finish(err)
		}
		d := time.Since(start)
		if v := conf.Get().ObservabilityLogSlowGraphQLRequests; v != 0 && d.Milliseconds() > int64(v) {
			encodedVariables, _ := json.Marshal(variables)
			log15.Warn("slow GraphQL request", "time", d, "name", requestName, "userID", currentUserID, "source", requestSource, "error", err, "variables", string(encodedVariables))
			if requestName == "unknown" {
				log.Printf(`logging complete query for slow GraphQL request above time=%v name=%s userID=%d source=%s error=%v:
QUERY
-----
%s

VARIABLES
---------
%s

`, d, requestName, currentUserID, requestSource, err, queryString, encodedVariables)
			}
		}
	}
}

func (prometheusTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]any) (context.Context, trace.TraceFieldFinishFunc) {
	// We don't call into t.OpenTracingTracer.TraceField since it generates too many spans which is really hard to read.
	start := time.Now()
	return ctx, func(err *gqlerrors.QueryError) {
		isErrStr := strconv.FormatBool(err != nil)
		graphqlFieldHistogram.WithLabelValues(
			prometheusTypeName(typeName),
			prometheusFieldName(typeName, fieldName),
			isErrStr,
			string(sgtrace.RequestSource(ctx)),
			prometheusGraphQLRequestName(sgtrace.GraphQLRequestName(ctx)),
		).Observe(time.Since(start).Seconds())

		origin := sgtrace.RequestOrigin(ctx)
		if origin != "unknown" && (fieldName == "search" || fieldName == "lsif") {
			isExact := strconv.FormatBool(fieldName == "lsif")
			codeIntelSearchHistogram.WithLabelValues(isExact, isErrStr).Observe(time.Since(start).Seconds())
		}
	}
}

func (t prometheusTracer) TraceValidation(ctx context.Context) trace.TraceValidationFinishFunc {
	var finish trace.TraceValidationFinishFunc
	if policy.ShouldTrace(ctx) {
		finish = t.tracer.TraceValidation(ctx)
	}
	return func(queryErrors []*gqlerrors.QueryError) {
		if finish != nil {
			finish(queryErrors)
		}
	}
}

var allowedPrometheusFieldNames = map[[2]string]struct{}{
	{"AccessTokenConnection", "nodes"}:          {},
	{"File", "isDirectory"}:                     {},
	{"File", "name"}:                            {},
	{"File", "path"}:                            {},
	{"File", "repository"}:                      {},
	{"File", "url"}:                             {},
	{"File2", "content"}:                        {},
	{"File2", "externalURLs"}:                   {},
	{"File2", "highlight"}:                      {},
	{"File2", "isDirectory"}:                    {},
	{"File2", "richHTML"}:                       {},
	{"File2", "url"}:                            {},
	{"FileDiff", "hunks"}:                       {},
	{"FileDiff", "internalID"}:                  {},
	{"FileDiff", "mostRelevantFile"}:            {},
	{"FileDiff", "newPath"}:                     {},
	{"FileDiff", "oldPath"}:                     {},
	{"FileDiff", "stat"}:                        {},
	{"FileDiffConnection", "diffStat"}:          {},
	{"FileDiffConnection", "nodes"}:             {},
	{"FileDiffConnection", "pageInfo"}:          {},
	{"FileDiffConnection", "totalCount"}:        {},
	{"FileDiffHunk", "body"}:                    {},
	{"FileDiffHunk", "newRange"}:                {},
	{"FileDiffHunk", "oldNoNewlineAt"}:          {},
	{"FileDiffHunk", "oldRange"}:                {},
	{"FileDiffHunk", "section"}:                 {},
	{"FileDiffHunkRange", "lines"}:              {},
	{"FileDiffHunkRange", "Line"}:               {},
	{"FileMatch", "file"}:                       {},
	{"FileMatch", "limitHit"}:                   {},
	{"FileMatch", "lineMatches"}:                {},
	{"FileMatch", "repository"}:                 {},
	{"FileMatch", "revSpec"}:                    {},
	{"FileMatch", "symbols"}:                    {},
	{"GitBlob", "blame"}:                        {},
	{"GitBlob", "commit"}:                       {},
	{"GitBlob", "content"}:                      {},
	{"GitBlob", "lsif"}:                         {},
	{"GitBlob", "path"}:                         {},
	{"GitBlob", "repository"}:                   {},
	{"GitBlob", "url"}:                          {},
	{"GitCommit", "abbreviatedOID"}:             {},
	{"GitCommit", "ancestors"}:                  {},
	{"GitCommit", "author"}:                     {},
	{"GitCommit", "blob"}:                       {},
	{"GitCommit", "body"}:                       {},
	{"GitCommit", "canonicalURL"}:               {},
	{"GitCommit", "committer"}:                  {},
	{"GitCommit", "externalURLs"}:               {},
	{"GitCommit", "file"}:                       {},
	{"GitCommit", "id"}:                         {},
	{"GitCommit", "message"}:                    {},
	{"GitCommit", "oid"}:                        {},
	{"GitCommit", "parents"}:                    {},
	{"GitCommit", "repository"}:                 {},
	{"GitCommit", "subject"}:                    {},
	{"GitCommit", "symbols"}:                    {},
	{"GitCommit", "tree"}:                       {},
	{"GitCommit", "url"}:                        {},
	{"GitCommitConnection", "nodes"}:            {},
	{"GitRefConnection", "nodes"}:               {},
	{"GitTree", "canonicalURL"}:                 {},
	{"GitTree", "entries"}:                      {},
	{"GitTree", "files"}:                        {},
	{"GitTree", "isRoot"}:                       {},
	{"GitTree", "url"}:                          {},
	{"Mutation", "configurationMutation"}:       {},
	{"Mutation", "createOrganization"}:          {},
	{"Mutation", "logEvent"}:                    {},
	{"Mutation", "logUserEvent"}:                {},
	{"Query", "clientConfiguration"}:            {},
	{"Query", "currentUser"}:                    {},
	{"Query", "dotcom"}:                         {},
	{"Query", "extensionRegistry"}:              {},
	{"Query", "highlightCode"}:                  {},
	{"Query", "node"}:                           {},
	{"Query", "organization"}:                   {},
	{"Query", "repositories"}:                   {},
	{"Query", "repository"}:                     {},
	{"Query", "repositoryRedirect"}:             {},
	{"Query", "search"}:                         {},
	{"Query", "settingsSubject"}:                {},
	{"Query", "site"}:                           {},
	{"Query", "user"}:                           {},
	{"Query", "viewerConfiguration"}:            {},
	{"Query", "viewerSettings"}:                 {},
	{"RegistryExtensionConnection", "nodes"}:    {},
	{"Repository", "cloneInProgress"}:           {},
	{"Repository", "commit"}:                    {},
	{"Repository", "comparison"}:                {},
	{"Repository", "gitRefs"}:                   {},
	{"RepositoryComparison", "commits"}:         {},
	{"RepositoryComparison", "fileDiffs"}:       {},
	{"RepositoryComparison", "range"}:           {},
	{"RepositoryConnection", "nodes"}:           {},
	{"Search", "results"}:                       {},
	{"Search", "suggestions"}:                   {},
	{"SearchAlert", "description"}:              {},
	{"SearchAlert", "proposedQueries"}:          {},
	{"SearchAlert", "title"}:                    {},
	{"SearchQueryDescription", "description"}:   {},
	{"SearchQueryDescription", "query"}:         {},
	{"SearchResultMatch", "body"}:               {},
	{"SearchResultMatch", "highlights"}:         {},
	{"SearchResultMatch", "url"}:                {},
	{"SearchResults", "alert"}:                  {},
	{"SearchResults", "approximateResultCount"}: {},
	{"SearchResults", "cloning"}:                {},
	{"SearchResults", "dynamicFilters"}:         {},
	{"SearchResults", "elapsedMilliseconds"}:    {},
	{"SearchResults", "indexUnavailable"}:       {},
	{"SearchResults", "limitHit"}:               {},
	{"SearchResults", "matchCount"}:             {},
	{"SearchResults", "missing"}:                {},
	{"SearchResults", "repositoriesCount"}:      {},
	{"SearchResults", "results"}:                {},
	{"SearchResults", "timedout"}:               {},
	{"SettingsCascade", "final"}:                {},
	{"SettingsMutation", "editConfiguration"}:   {},
	{"SettingsSubject", "latestSettings"}:       {},
	{"SettingsSubject", "settingsCascade"}:      {},
	{"Signature", "date"}:                       {},
	{"Signature", "person"}:                     {},
	{"Site", "alerts"}:                          {},
	{"SymbolConnection", "nodes"}:               {},
	{"TreeEntry", "isDirectory"}:                {},
	{"TreeEntry", "isSingleChild"}:              {},
	{"TreeEntry", "name"}:                       {},
	{"TreeEntry", "path"}:                       {},
	{"TreeEntry", "submodule"}:                  {},
	{"TreeEntry", "url"}:                        {},
	{"UserConnection", "nodes"}:                 {},
}

// prometheusFieldName reduces the cardinality of GraphQL field names to make it suitable
// for use in a Prometheus metric. We only track the ones most valuable to us.
//
// See https://github.com/sourcegraph/sourcegraph/issues/9895
func prometheusFieldName(typeName, fieldName string) string {
	if _, ok := allowedPrometheusFieldNames[[2]string{typeName, fieldName}]; ok {
		return fieldName
	}
	return "other"
}

var blocklistedPrometheusTypeNames = map[string]struct{}{
	"__Type":                                 {},
	"__Schema":                               {},
	"__InputValue":                           {},
	"__Field":                                {},
	"__EnumValue":                            {},
	"__Directive":                            {},
	"UserEmail":                              {},
	"UpdateSettingsPayload":                  {},
	"ExtensionRegistryCreateExtensionResult": {},
	"Range":                                  {},
	"LineMatch":                              {},
	"DiffStat":                               {},
	"DiffHunk":                               {},
	"DiffHunkRange":                          {},
	"FileDiffResolver":                       {},
}

// prometheusTypeName reduces the cardinality of GraphQL type names to make it
// suitable for use in a Prometheus metric. This is a blocklist of type names
// which involve non-complex calculations in the GraphQL backend and thus are
// not worth tracking. You can find a complete list of the ones Prometheus is
// currently tracking via:
//
//	sum by (type)(src_graphql_field_seconds_count)
func prometheusTypeName(typeName string) string {
	if _, ok := blocklistedPrometheusTypeNames[typeName]; ok {
		return "other"
	}
	return typeName
}

// prometheusGraphQLRequestName is a allowlist of GraphQL request names (e.g. /.api/graphql?Foobar)
// to include in a Prometheus metric. Be extremely careful
func prometheusGraphQLRequestName(requestName string) string {
	if requestName == "CodeIntelSearch" {
		return requestName
	}
	return "other"
}

func NewSchema(
	db database.DB,
	batchChanges BatchChangesResolver,
	codeIntel CodeIntelResolver,
	insights InsightsResolver,
	authz AuthzResolver,
	codeMonitors CodeMonitorsResolver,
	license LicenseResolver,
	dotcom DotcomRootResolver,
	searchContexts SearchContextsResolver,
	notebooks NotebooksResolver,
	compute ComputeResolver,
	insightsAggregation InsightsAggregationResolver,
) (*graphql.Schema, error) {
	resolver := newSchemaResolver(db)
	schemas := []string{mainSchema}

	if batchChanges != nil {
		EnterpriseResolvers.batchChangesResolver = batchChanges
		resolver.BatchChangesResolver = batchChanges
		schemas = append(schemas, batchesSchema)
		// Register NodeByID handlers.
		for kind, res := range batchChanges.NodeResolvers() {
			resolver.nodeByIDFns[kind] = res
		}
	}

	if codeIntel != nil {
		EnterpriseResolvers.codeIntelResolver = codeIntel
		resolver.CodeIntelResolver = codeIntel
		schemas = append(schemas, codeIntelSchema)
		// Register NodeByID handlers.
		for kind, res := range codeIntel.NodeResolvers() {
			resolver.nodeByIDFns[kind] = res
		}
	}

	if insights != nil {
		EnterpriseResolvers.insightsResolver = insights
		resolver.InsightsResolver = insights
		schemas = append(schemas, insightsSchema)
	}

	if authz != nil {
		EnterpriseResolvers.authzResolver = authz
		resolver.AuthzResolver = authz
		schemas = append(schemas, authzSchema)
	}

	if codeMonitors != nil {
		EnterpriseResolvers.codeMonitorsResolver = codeMonitors
		resolver.CodeMonitorsResolver = codeMonitors
		schemas = append(schemas, codeMonitorsSchema)
		// Register NodeByID handlers.
		for kind, res := range codeMonitors.NodeResolvers() {
			resolver.nodeByIDFns[kind] = res
		}
	}

	if license != nil {
		EnterpriseResolvers.licenseResolver = license
		resolver.LicenseResolver = license
		schemas = append(schemas, licenseSchema)
		// No NodeByID handlers currently.
	}

	if dotcom != nil {
		EnterpriseResolvers.dotcomResolver = dotcom
		resolver.DotcomRootResolver = dotcom
		schemas = append(schemas, dotcomSchema)
		// Register NodeByID handlers.
		for kind, res := range dotcom.NodeResolvers() {
			resolver.nodeByIDFns[kind] = res
		}
	}

	if searchContexts != nil {
		EnterpriseResolvers.searchContextsResolver = searchContexts
		resolver.SearchContextsResolver = searchContexts
		schemas = append(schemas, searchContextsSchema)
		// Register NodeByID handlers.
		for kind, res := range searchContexts.NodeResolvers() {
			resolver.nodeByIDFns[kind] = res
		}
	}

	if notebooks != nil {
		EnterpriseResolvers.notebooksResolver = notebooks
		resolver.NotebooksResolver = notebooks
		schemas = append(schemas, notebooksSchema)
		// Register NodeByID handlers.
		for kind, res := range notebooks.NodeResolvers() {
			resolver.nodeByIDFns[kind] = res
		}
	}

	if compute != nil {
		EnterpriseResolvers.computeResolver = compute
		resolver.ComputeResolver = compute
		schemas = append(schemas, computeSchema)
	}

	if insightsAggregation != nil {
		EnterpriseResolvers.InsightsAggregationResolver = insightsAggregation
		resolver.InsightsAggregationResolver = insightsAggregation
		schemas = append(schemas, insightsAggregationsSchema)
	}

	return graphql.ParseSchema(
		strings.Join(schemas, "\n"),
		resolver,
		graphql.Tracer(&prometheusTracer{}),
		graphql.UseStringDescriptions(),
	)
}

// schemaResolver handles all GraphQL queries for Sourcegraph. To do this, it
// uses subresolvers which are globals. Enterprise-only resolvers are assigned
// to a field of EnterpriseResolvers.
//
// schemaResolver must be instantiated using newSchemaResolver.
type schemaResolver struct {
	logger            sglog.Logger
	db                database.DB
	repoupdaterClient *repoupdater.Client
	nodeByIDFns       map[string]NodeByIDFunc

	// SubResolvers are assigned using the Schema constructor.

	BatchChangesResolver
	AuthzResolver
	CodeIntelResolver
	ComputeResolver
	InsightsResolver
	CodeMonitorsResolver
	LicenseResolver
	DotcomRootResolver
	SearchContextsResolver
	NotebooksResolver
	InsightsAggregationResolver
}

// newSchemaResolver will return a new, safely instantiated schemaResolver with some
// defaults. It does not implement any sub-resolvers.
func newSchemaResolver(db database.DB) *schemaResolver {
	r := &schemaResolver{
		logger:            sglog.Scoped("schemaResolver", "GraphQL schema resolver"),
		db:                db,
		repoupdaterClient: repoupdater.DefaultClient,
	}

	r.nodeByIDFns = map[string]NodeByIDFunc{
		"AccessToken": func(ctx context.Context, id graphql.ID) (Node, error) {
			return accessTokenByID(ctx, db, id)
		},
		"ExternalAccount": func(ctx context.Context, id graphql.ID) (Node, error) {
			return externalAccountByID(ctx, db, id)
		},
		externalServiceIDKind: func(ctx context.Context, id graphql.ID) (Node, error) {
			return externalServiceByID(ctx, db, id)
		},
		"GitRef": func(ctx context.Context, id graphql.ID) (Node, error) {
			return r.gitRefByID(ctx, id)
		},
		"Repository": func(ctx context.Context, id graphql.ID) (Node, error) {
			return r.repositoryByID(ctx, id)
		},
		"User": func(ctx context.Context, id graphql.ID) (Node, error) {
			return UserByID(ctx, db, id)
		},
		"Org": func(ctx context.Context, id graphql.ID) (Node, error) {
			return OrgByID(ctx, db, id)
		},
		"OrganizationInvitation": func(ctx context.Context, id graphql.ID) (Node, error) {
			return orgInvitationByID(ctx, db, id)
		},
		"GitCommit": func(ctx context.Context, id graphql.ID) (Node, error) {
			return r.gitCommitByID(ctx, id)
		},
		"RegistryExtension": func(ctx context.Context, id graphql.ID) (Node, error) {
			return RegistryExtensionByID(ctx, db, id)
		},
		"SavedSearch": func(ctx context.Context, id graphql.ID) (Node, error) {
			return r.savedSearchByID(ctx, id)
		},
		"Site": func(ctx context.Context, id graphql.ID) (Node, error) {
			return r.siteByGQLID(ctx, id)
		},
		"OutOfBandMigration": func(ctx context.Context, id graphql.ID) (Node, error) {
			return r.OutOfBandMigrationByID(ctx, id)
		},
		"WebhookLog": func(ctx context.Context, id graphql.ID) (Node, error) {
			return webhookLogByID(ctx, db, id)
		},
		"Executor": func(ctx context.Context, id graphql.ID) (Node, error) {
			return executorByID(ctx, db, id, r)
		},
		"ExternalServiceSyncJob": func(ctx context.Context, id graphql.ID) (Node, error) {
			return externalServiceSyncJobByID(ctx, db, id)
		},
	}
	return r
}

// EnterpriseResolvers holds the instances of resolvers which are enabled only
// in enterprise mode. These resolver instances are nil when running as OSS.
var EnterpriseResolvers = struct {
	codeIntelResolver           CodeIntelResolver
	computeResolver             ComputeResolver
	insightsResolver            InsightsResolver
	authzResolver               AuthzResolver
	batchChangesResolver        BatchChangesResolver
	codeMonitorsResolver        CodeMonitorsResolver
	licenseResolver             LicenseResolver
	dotcomResolver              DotcomRootResolver
	searchContextsResolver      SearchContextsResolver
	notebooksResolver           NotebooksResolver
	InsightsAggregationResolver InsightsAggregationResolver
}{}

// DEPRECATED
func (r *schemaResolver) Root() *schemaResolver {
	return newSchemaResolver(r.db)
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
	resolver, err := r.RepositoryRedirect(ctx, &repositoryRedirectArgs{args.Name, args.CloneURL, nil})
	if err != nil {
		return nil, err
	}
	if resolver == nil {
		return nil, nil
	}
	return resolver.repo, nil
}

func (r *schemaResolver) repositoryByID(ctx context.Context, id graphql.ID) (*RepositoryResolver, error) {
	var repoID api.RepoID
	if err := relay.UnmarshalSpec(id, &repoID); err != nil {
		return nil, err
	}
	repo, err := r.db.Repos().Get(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return NewRepositoryResolver(r.db, repo), nil
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

type repositoryRedirectArgs struct {
	Name       *string
	CloneURL   *string
	HashedName *string
}

func (r *repositoryRedirect) ToRepository() (*RepositoryResolver, bool) {
	return r.repo, r.repo != nil
}

func (r *repositoryRedirect) ToRedirect() (*RedirectResolver, bool) {
	return r.redirect, r.redirect != nil
}

func (r *schemaResolver) RepositoryRedirect(ctx context.Context, args *repositoryRedirectArgs) (*repositoryRedirect, error) {
	if args.HashedName != nil {
		// Query by repository hashed name
		repo, err := r.db.Repos().GetByHashedName(ctx, api.RepoHashedName(*args.HashedName))
		if err != nil {
			return nil, err
		}
		return &repositoryRedirect{repo: NewRepositoryResolver(r.db, repo)}, nil
	}
	var name api.RepoName
	if args.Name != nil {
		// Query by name
		name = api.RepoName(*args.Name)
	} else if args.CloneURL != nil {
		// Query by git clone URL
		var err error
		name, err = cloneurls.ReposourceCloneURLToRepoName(ctx, r.db, *args.CloneURL)
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

	repo, err := backend.NewRepos(r.logger, r.db).GetByName(ctx, name)
	if err != nil {
		var e backend.ErrRepoSeeOther
		if errors.As(err, &e) {
			return &repositoryRedirect{redirect: &RedirectResolver{url: e.RedirectURL}}, nil
		}
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &repositoryRedirect{repo: NewRepositoryResolver(r.db, repo)}, nil
}

func (r *schemaResolver) PhabricatorRepo(ctx context.Context, args *struct {
	Name *string
	// TODO(chris): Remove URI in favor of Name.
	URI *string
}) (*phabricatorRepoResolver, error) {
	if args.Name != nil {
		args.URI = args.Name
	}

	repo, err := r.db.Phabricator().GetByName(ctx, api.RepoName(*args.URI))
	if err != nil {
		return nil, err
	}
	return &phabricatorRepoResolver{repo}, nil
}

func (r *schemaResolver) CurrentUser(ctx context.Context) (*UserResolver, error) {
	return CurrentUser(ctx, r.db)
}

func (r *schemaResolver) AffiliatedRepositories(ctx context.Context, args *struct {
	Namespace graphql.ID
	CodeHost  *graphql.ID
	Query     *string
}) (*affiliatedRepositoriesConnection, error) {
	var userID, orgID int32
	err := UnmarshalNamespaceID(args.Namespace, &userID, &orgID)
	if err != nil {
		return nil, err
	}
	if userID > 0 {
		// 🚨 SECURITY: Make sure the user is the same user being requested
		if err := backend.CheckSameUser(ctx, userID); err != nil {
			return nil, err
		}
	}
	if orgID > 0 {
		// 🚨 SECURITY: Make sure the user can access the organization
		if err := backend.CheckOrgAccess(ctx, r.db, orgID); err != nil {
			return nil, err
		}
	}
	var codeHost int64
	if args.CodeHost != nil {
		codeHost, err = UnmarshalExternalServiceID(*args.CodeHost)
		if err != nil {
			return nil, err
		}
	}
	var query string
	if args.Query != nil {
		query = *args.Query
	}

	return &affiliatedRepositoriesConnection{
		db:       r.db,
		userID:   userID,
		orgID:    orgID,
		codeHost: codeHost,
		query:    query,
	}, nil
}

// CodeHostSyncDue returns true if any of the supplied code hosts are due to sync
// now or within "seconds" from now.
func (r *schemaResolver) CodeHostSyncDue(ctx context.Context, args *struct {
	IDs     []graphql.ID
	Seconds int32
}) (bool, error) {
	if len(args.IDs) == 0 {
		return false, errors.New("no ids supplied")
	}
	ids := make([]int64, len(args.IDs))
	for i, gqlID := range args.IDs {
		id, err := UnmarshalExternalServiceID(gqlID)
		if err != nil {
			return false, errors.New("unable to unmarshal id")
		}
		ids[i] = id
	}
	return r.db.ExternalServices().SyncDue(ctx, ids, time.Duration(args.Seconds)*time.Second)
}
