package sourcegraph

import "google.golang.org/grpc"

// A Client communicates with the Sourcegraph API. All communication
// is done using gRPC over HTTP/2.
type Client struct {
	// Services used to communicate with different parts of the Sourcegraph API.
	Annotations  AnnotationsClient
	Async        AsyncClient
	Defs         DefsClient
	Meta         MetaClient
	MirrorRepos  MirrorReposClient
	RepoStatuses RepoStatusesClient
	RepoTree     RepoTreeClient
	Repos        ReposClient
	Search       SearchClient

	// gRPC client connection used to communicate with the Sourcegraph
	// API.
	Conn *grpc.ClientConn
}

// NewClient returns a Sourcegraph API client.
func NewClient(conn *grpc.ClientConn) *Client {
	c := new(Client)

	// gRPC (HTTP/2)
	c.Conn = conn
	c.Annotations = NewAnnotationsClient(conn)
	c.Async = NewAsyncClient(conn)
	c.Defs = NewDefsClient(conn)
	c.Meta = NewMetaClient(conn)
	c.MirrorRepos = NewMirrorReposClient(conn)
	c.RepoStatuses = NewRepoStatusesClient(conn)
	c.RepoTree = NewRepoTreeClient(conn)
	c.Repos = NewReposClient(conn)
	c.Search = NewSearchClient(conn)

	return c
}
