package sourcegraph

import (
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/grpccache"
)

// A Client communicates with the Sourcegraph API. All communication
// is done using gRPC over HTTP/2.
type Client struct {
	// Services used to communicate with different parts of the Sourcegraph API.
	Accounts            AccountsClient
	Auth                AuthClient
	Builds              BuildsClient
	Defs                DefsClient
	Deltas              DeltasClient
	Discussions         DiscussionsClient
	GraphUplink         GraphUplinkClient
	Markdown            MarkdownClient
	Meta                MetaClient
	MirrorRepos         MirrorReposClient
	MirroredRepoSSHKeys MirroredRepoSSHKeysClient
	Notify              NotifyClient
	Orgs                OrgsClient
	People              PeopleClient
	RegisteredClients   RegisteredClientsClient
	RepoBadges          RepoBadgesClient
	RepoStatuses        RepoStatusesClient
	RepoTree            RepoTreeClient
	Repos               ReposClient
	Storage             StorageClient
	Changesets          ChangesetsClient
	Search              SearchClient
	Units               UnitsClient
	Users               UsersClient
	UserKeys            UserKeysClient

	// gRPC client connection used to communicate with the Sourcegraph
	// API.
	Conn *grpc.ClientConn
}

// Cache is the gRPC cache used to cache API responses.
var Cache *grpccache.Cache

// NewClient returns a Sourcegraph API client.
func NewClient(conn *grpc.ClientConn) *Client {
	c := new(Client)

	// gRPC (HTTP/2)
	c.Conn = conn
	c.Accounts = &CachedAccountsClient{NewAccountsClient(conn), Cache}
	c.Auth = &CachedAuthClient{NewAuthClient(conn), Cache}
	c.Builds = &CachedBuildsClient{NewBuildsClient(conn), Cache}
	c.Defs = &CachedDefsClient{NewDefsClient(conn), Cache}
	c.Deltas = &CachedDeltasClient{NewDeltasClient(conn), Cache}
	c.Discussions = &CachedDiscussionsClient{NewDiscussionsClient(conn), Cache}
	c.GraphUplink = &CachedGraphUplinkClient{NewGraphUplinkClient(conn), Cache}
	c.Markdown = &CachedMarkdownClient{NewMarkdownClient(conn), Cache}
	c.Meta = &CachedMetaClient{NewMetaClient(conn), Cache}
	c.MirrorRepos = &CachedMirrorReposClient{NewMirrorReposClient(conn), Cache}
	c.MirroredRepoSSHKeys = &CachedMirroredRepoSSHKeysClient{NewMirroredRepoSSHKeysClient(conn), Cache}
	c.Notify = &CachedNotifyClient{NewNotifyClient(conn), Cache}
	c.Orgs = &CachedOrgsClient{NewOrgsClient(conn), Cache}
	c.People = &CachedPeopleClient{NewPeopleClient(conn), Cache}
	c.RegisteredClients = &CachedRegisteredClientsClient{NewRegisteredClientsClient(conn), Cache}
	c.RepoBadges = &CachedRepoBadgesClient{NewRepoBadgesClient(conn), Cache}
	c.RepoStatuses = &CachedRepoStatusesClient{NewRepoStatusesClient(conn), Cache}
	c.RepoTree = &CachedRepoTreeClient{NewRepoTreeClient(conn), Cache}
	c.Repos = &CachedReposClient{NewReposClient(conn), Cache}
	c.Storage = &CachedStorageClient{NewStorageClient(conn), Cache}
	c.Changesets = &CachedChangesetsClient{NewChangesetsClient(conn), Cache}
	c.Search = &CachedSearchClient{NewSearchClient(conn), Cache}
	c.Units = &CachedUnitsClient{NewUnitsClient(conn), Cache}
	c.Users = &CachedUsersClient{NewUsersClient(conn), Cache}
	c.UserKeys = &CachedUserKeysClient{NewUserKeysClient(conn), Cache}

	return c
}
