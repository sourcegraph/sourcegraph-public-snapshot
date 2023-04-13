// Package router contains the URL router for the HTTP API.
package router

import (
	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/routevar"
)

const (
	GraphQL = "graphql"

	LSIFUpload       = "lsif.upload"
	SCIPUpload       = "scip.upload"
	SCIPUploadExists = "scip.upload.exists"

	SearchStream      = "search.stream"
	ComputeStream     = "compute.stream"
	GitBlameStream    = "git.blame.stream"
	CompletionsStream = "completions.stream"
	CodeCompletions   = "completions.code"

	SrcCli             = "src-cli"
	SrcCliVersionCache = "src-cli.version-cache"

	Registry = "registry"

	RepoShield  = "repo.shield"
	RepoRefresh = "repo.refresh"
	Telemetry   = "telemetry"

	Webhooks                = "webhooks"
	GitHubWebhooks          = "github.webhooks"
	GitLabWebhooks          = "gitlab.webhooks"
	BitbucketServerWebhooks = "bitbucketServer.webhooks"
	BitbucketCloudWebhooks  = "bitbucketCloud.webhooks"

	SCIM = "scim"

	BatchesFileGet    = "batches.file.get"
	BatchesFileExists = "batches.file.exists"
	BatchesFileUpload = "batches.file.upload"

	CodeInsightsDataExport = "insights.data.export"

	ExternalURL            = "internal.app-url"
	SendEmail              = "internal.send-email"
	GitInfoRefs            = "internal.git.info-refs"
	GitUploadPack          = "internal.git.upload-pack"
	ReposIndex             = "internal.repos.index"
	Configuration          = "internal.configuration"
	SearchConfiguration    = "internal.search-configuration"
	ExternalServiceConfigs = "internal.external-services.configs"
	StreamingSearch        = "internal.stream-search"
	RepoRank               = "internal.repo-rank"
	DocumentRanks          = "internal.document-ranks"
	UpdateIndexStatus      = "internal.update-index-status"
)

// New creates a new API router with route URL pattern definitions but
// no handlers attached to the routes.
func New(base *mux.Router) *mux.Router {
	if base == nil {
		panic("base == nil")
	}

	base.StrictSlash(true)

	addRegistryRoute(base)
	addSCIMRoute(base)
	addGraphQLRoute(base)
	base.Path("/webhooks/{webhook_uuid}").Methods("POST").Name(Webhooks)
	base.Path("/github-webhooks").Methods("POST").Name(GitHubWebhooks)
	base.Path("/gitlab-webhooks").Methods("POST").Name(GitLabWebhooks)
	base.Path("/bitbucket-server-webhooks").Methods("POST").Name(BitbucketServerWebhooks)
	base.Path("/bitbucket-cloud-webhooks").Methods("POST").Name(BitbucketCloudWebhooks)
	base.Path("/files/batch-changes/{spec}/{file}").Methods("GET").Name(BatchesFileGet)
	base.Path("/files/batch-changes/{spec}/{file}").Methods("HEAD").Name(BatchesFileExists)
	base.Path("/files/batch-changes/{spec}").Methods("POST").Name(BatchesFileUpload)
	base.Path("/lsif/upload").Methods("POST").Name(LSIFUpload)
	base.Path("/scip/upload").Methods("POST").Name(SCIPUpload)
	base.Path("/scip/upload").Methods("HEAD").Name(SCIPUploadExists)
	base.Path("/search/stream").Methods("GET").Name(SearchStream)
	base.Path("/compute/stream").Methods("GET", "POST").Name(ComputeStream)
	base.Path("/blame/" + routevar.Repo + routevar.RepoRevSuffix + "/stream/{Path:.*}").Methods("GET").Name(GitBlameStream)
	base.Path("/src-cli/versions/{rest:.*}").Methods("GET", "POST").Name(SrcCliVersionCache)
	base.Path("/src-cli/{rest:.*}").Methods("GET").Name(SrcCli)
	base.Path("/insights/export/{id}").Methods("GET").Name(CodeInsightsDataExport)
	base.Path("/completions/stream").Methods("POST").Name(CompletionsStream)
	base.Path("/completions/code").Methods("POST").Name(CodeCompletions)

	// repo contains routes that are NOT specific to a revision. In these routes, the URL may not contain a revspec after the repo (that is, no "github.com/foo/bar@myrevspec").
	repoPath := `/repos/` + routevar.Repo

	// Additional paths added will be treated as a repo. To add a new path that should not be treated as a repo
	// add above repo paths.
	repo := base.PathPrefix(repoPath + "/" + routevar.RepoPathDelim + "/").Subrouter()
	repo.Path("/shield").Methods("GET").Name(RepoShield)
	repo.Path("/refresh").Methods("POST").Name(RepoRefresh)

	return base
}

// NewInternal creates a new API router for internal endpoints.
func NewInternal(base *mux.Router) *mux.Router {
	if base == nil {
		panic("base == nil")
	}

	base.StrictSlash(true)
	// Internal API endpoints should only be served on the internal Handler
	base.Path("/app-url").Methods("POST").Name(ExternalURL)
	base.Path("/send-email").Methods("POST").Name(SendEmail)
	base.Path("/git/{RepoName:.*}/info/refs").Methods("GET").Name(GitInfoRefs)
	base.Path("/git/{RepoName:.*}/git-upload-pack").Methods("GET", "POST").Name(GitUploadPack)
	base.Path("/external-services/configs").Methods("POST").Name(ExternalServiceConfigs)
	base.Path("/repos/index").Methods("POST").Name(ReposIndex)
	base.Path("/configuration").Methods("POST").Name(Configuration)
	base.Path("/ranks/{RepoName:.*}/documents").Methods("GET").Name(DocumentRanks)
	base.Path("/ranks/{RepoName:.*}").Methods("GET").Name(RepoRank)
	base.Path("/search/configuration").Methods("GET", "POST").Name(SearchConfiguration)
	base.Path("/search/index-status").Methods("POST").Name(UpdateIndexStatus)
	base.Path("/telemetry").Methods("POST").Name(Telemetry)
	base.Path("/lsif/upload").Methods("POST").Name(LSIFUpload)
	base.Path("/scip/upload").Methods("POST").Name(SCIPUpload)
	base.Path("/scip/upload").Methods("HEAD").Name(SCIPUploadExists)
	base.Path("/search/stream").Methods("GET").Name(StreamingSearch)
	base.Path("/compute/stream").Methods("GET", "POST").Name(ComputeStream)
	addRegistryRoute(base)
	addGraphQLRoute(base)

	return base
}

func addRegistryRoute(m *mux.Router) {
	m.PathPrefix("/registry").Methods("GET").Name(Registry)
}

func addSCIMRoute(m *mux.Router) {
	m.PathPrefix("/scim/v2").Methods("GET", "POST", "PUT", "PATCH", "DELETE").Name(SCIM)
}

func addGraphQLRoute(m *mux.Router) {
	m.Path("/graphql").Methods("POST").Name(GraphQL)
}
