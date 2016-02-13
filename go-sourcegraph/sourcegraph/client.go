package sourcegraph

import "google.golang.org/grpc"

// A Client communicates with the Sourcegraph API. All communication
// is done using gRPC over HTTP/2.
type Client struct {
	// Services used to communicate with different parts of the Sourcegraph API.
	Accounts            AccountsClient
	Auth                AuthClient
	Builds              BuildsClient
	Defs                DefsClient
	Deltas              DeltasClient
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

// NewClient returns a Sourcegraph API client.
func NewClient(conn *grpc.ClientConn) *Client {
	c := new(Client)

	// gRPC (HTTP/2)
	c.Conn = conn
	c.Accounts = NewAccountsClient(conn)
	c.Auth = NewAuthClient(conn)
	c.Builds = NewBuildsClient(conn)
	c.Defs = NewDefsClient(conn)
	c.Deltas = NewDeltasClient(conn)
	c.GraphUplink = NewGraphUplinkClient(conn)
	c.Markdown = NewMarkdownClient(conn)
	c.Meta = NewMetaClient(conn)
	c.MirrorRepos = NewMirrorReposClient(conn)
	c.MirroredRepoSSHKeys = NewMirroredRepoSSHKeysClient(conn)
	c.Notify = NewNotifyClient(conn)
	c.Orgs = NewOrgsClient(conn)
	c.People = NewPeopleClient(conn)
	c.RegisteredClients = NewRegisteredClientsClient(conn)
	c.RepoBadges = NewRepoBadgesClient(conn)
	c.RepoStatuses = NewRepoStatusesClient(conn)
	c.RepoTree = NewRepoTreeClient(conn)
	c.Repos = NewReposClient(conn)
	c.Storage = NewStorageClient(conn)
	c.Changesets = NewChangesetsClient(conn)
	c.Search = NewSearchClient(conn)
	c.Units = NewUnitsClient(conn)
	c.Users = NewUsersClient(conn)
	c.UserKeys = NewUserKeysClient(conn)

	return c
}
