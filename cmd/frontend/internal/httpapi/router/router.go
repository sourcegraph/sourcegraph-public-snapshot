// Package router contains the URL router for the HTTP API.
package router

import (
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	GraphQL = "graphql"
	XLang   = "xlang"

	Registry = "registry"

	RepoShield  = "repo.shield"
	RepoRefresh = "repo.refresh"
	Telemetry   = "telemetry"

	SavedQueriesListAll    = "internal.saved-queries.list-all"
	SavedQueriesGetInfo    = "internal.saved-queries.get-info"
	SavedQueriesSetInfo    = "internal.saved-queries.set-info"
	SavedQueriesDeleteInfo = "internal.saved-queries.delete-info"
	SettingsGetForSubject  = "internal.settings.get-for-subject"
	OrgsListUsers          = "internal.orgs.list-users"
	OrgsGetByName          = "internal.orgs.get-by-name"
	UsersGetByUsername     = "internal.users.get-by-username"
	UserEmailsGetEmail     = "internal.user-emails.get-email"
	AppURL                 = "internal.app-url"
	CanSendEmail           = "internal.can-send-email"
	SendEmail              = "internal.send-email"
	Extension              = "internal.extension"
	DefsRefreshIndex       = "internal.defs.refresh-index"
	PkgsRefreshIndex       = "internal.pkgs.refresh-index"
	GitInfoRefs            = "internal.git.info-refs"
	GitResolveRevision     = "internal.git.resolve-revision"
	GitTar                 = "internal.git.tar"
	GitUploadPack          = "internal.git.upload-pack"
	PhabricatorRepoCreate  = "internal.phabricator.repo.create"
	ReposCreateIfNotExists = "internal.repos.create-if-not-exists"
	ReposGetByURI          = "internal.repos.get-by-uri"
	ReposInventoryUncached = "internal.repos.inventory-uncached"
	ReposInventory         = "internal.repos.inventory"
	ReposList              = "internal.repos.list"
	ReposListEnabled       = "internal.repos.list-enabled"
	ReposUpdateIndex       = "internal.repos.update-index"
	ReposUpdateMetadata    = "internal.repos.update-metadata"
)

// New creates a new API router with route URL pattern definitions but
// no handlers attached to the routes.
func New(base *mux.Router) *mux.Router {
	if base == nil {
		panic("base == nil")
	}

	base.StrictSlash(true)

	base.Path("/xlang/{LSPMethod:.*}").Methods("POST").Name(XLang)
	addRegistryRoute(base)
	addGraphQLRoute(base)
	addTelemetryRoute(base)

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
	base.Path("/saved-queries/list-all").Methods("POST").Name(SavedQueriesListAll)
	base.Path("/saved-queries/get-info").Methods("POST").Name(SavedQueriesGetInfo)
	base.Path("/saved-queries/set-info").Methods("POST").Name(SavedQueriesSetInfo)
	base.Path("/saved-queries/delete-info").Methods("POST").Name(SavedQueriesDeleteInfo)
	base.Path("/settings/get-for-subject").Methods("POST").Name(SettingsGetForSubject)
	base.Path("/orgs/list-users").Methods("POST").Name(OrgsListUsers)
	base.Path("/orgs/get-by-name").Methods("POST").Name(OrgsGetByName)
	base.Path("/users/get-by-username").Methods("POST").Name(UsersGetByUsername)
	base.Path("/user-emails/get-email").Methods("POST").Name(UserEmailsGetEmail)
	base.Path("/app-url").Methods("POST").Name(AppURL)
	base.Path("/can-send-email").Methods("POST").Name(CanSendEmail)
	base.Path("/send-email").Methods("POST").Name(SendEmail)
	base.Path("/extension").Methods("POST").Name(Extension)
	base.Path("/defs/refresh-index").Methods("POST").Name(DefsRefreshIndex)
	base.Path("/pkgs/refresh-index").Methods("POST").Name(PkgsRefreshIndex)
	base.Path("/git/{RepoURI:.*}/info/refs").Methods("GET").Name(GitInfoRefs)
	base.Path("/git/{RepoURI:.*}/resolve-revision/{Spec}").Methods("GET").Name(GitResolveRevision)
	base.Path("/git/{RepoURI:.*}/tar/{Commit}").Methods("GET").Name(GitTar)
	base.Path("/git/{RepoURI:.*}/git-upload-pack").Methods("POST").Name(GitUploadPack)
	base.Path("/phabricator/repo-create").Methods("POST").Name(PhabricatorRepoCreate)
	base.Path("/repos/create-if-not-exists").Methods("POST").Name(ReposCreateIfNotExists)
	base.Path("/repos/inventory-uncached").Methods("POST").Name(ReposInventoryUncached)
	base.Path("/repos/inventory").Methods("POST").Name(ReposInventory)
	base.Path("/repos/list").Methods("POST").Name(ReposList)
	base.Path("/repos/list-enabled").Methods("POST").Name(ReposListEnabled)
	base.Path("/repos/update-index").Methods("POST").Name(ReposUpdateIndex)
	base.Path("/repos/update-metadata").Methods("POST").Name(ReposUpdateMetadata)
	base.Path("/repos/{RepoURI:.*}").Methods("POST").Name(ReposGetByURI)
	addRegistryRoute(base)
	addGraphQLRoute(base)
	addTelemetryRoute(base)

	return base
}

func addRegistryRoute(m *mux.Router) {
	m.PathPrefix("/registry").Methods("GET").Name(Registry)
}

func addTelemetryRoute(m *mux.Router) {
	m.Path("/telemetry/{TelemetryPath:.*}").Methods("POST").Name(Telemetry)
}

func addGraphQLRoute(m *mux.Router) {
	m.Path("/graphql").Methods("POST").Name(GraphQL)
}
