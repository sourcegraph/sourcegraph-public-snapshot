// Package router contains the URL router for the HTTP API.
package router

import (
	"github.com/gorilla/mux"
)

const (
	GraphQL = "graphql"

	LSIFUpload       = "lsif.upload"
	SCIPUpload       = "scip.upload"
	SCIPUploadExists = "scip.upload.exists"

	ComputeStream = "compute.stream"

	Registry = "registry"

	SCIM = "scim"

	GitInfoRefs         = "internal.git.info-refs"
	GitUploadPack       = "internal.git.upload-pack"
	ReposIndex          = "internal.repos.index"
	Configuration       = "internal.configuration"
	SearchConfiguration = "internal.search-configuration"
	StreamingSearch     = "internal.stream-search"
	RepoRank            = "internal.repo-rank"
	DocumentRanks       = "internal.document-ranks"
	UpdateIndexStatus   = "internal.update-index-status"
)

// NewInternal creates a new API router for internal endpoints.
func NewInternal(base *mux.Router) *mux.Router {
	if base == nil {
		panic("base == nil")
	}

	base.StrictSlash(true)
	// Internal API endpoints should only be served on the internal Handler
	base.Path("/git/{RepoName:.*}/info/refs").Methods("GET").Name(GitInfoRefs)
	base.Path("/git/{RepoName:.*}/git-upload-pack").Methods("GET", "POST").Name(GitUploadPack)
	base.Path("/repos/index").Methods("POST").Name(ReposIndex)
	base.Path("/configuration").Methods("POST").Name(Configuration)
	base.Path("/ranks/{RepoName:.*}/documents").Methods("GET").Name(DocumentRanks)
	base.Path("/ranks/{RepoName:.*}").Methods("GET").Name(RepoRank)
	base.Path("/search/configuration").Methods("GET", "POST").Name(SearchConfiguration)
	base.Path("/search/index-status").Methods("POST").Name(UpdateIndexStatus)
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
