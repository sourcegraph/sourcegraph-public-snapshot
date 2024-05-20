package gitserver

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/connection"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
)

type TestClientSourceOptions struct {
	// ClientFunc is the function that is used to return a gRPC client
	// given the provided connection.
	ClientFunc func(conn *grpc.ClientConn) proto.GitserverServiceClient

	// Logger is the log.Logger instance that the test ClientSource will use to
	// log various metadata to.
	Logger log.Logger
}

func NewTestClientSource(t testing.TB, addrs []string, options ...func(o *TestClientSourceOptions)) ClientSource {
	logger := logtest.Scoped(t)
	opts := TestClientSourceOptions{
		ClientFunc: func(conn *grpc.ClientConn) proto.GitserverServiceClient {
			return proto.NewGitserverServiceClient(conn)
		},

		Logger: logger,
	}

	for _, o := range options {
		o(&opts)
	}

	conns := make(map[string]connAndErr)
	var testAddresses []AddressWithClient
	for _, addr := range addrs {
		conn, err := defaults.Dial(addr, logger)
		conns[addr] = connAndErr{address: addr, conn: conn, err: err}
		testAddresses = append(testAddresses, &testConnAndErr{
			address:    addr,
			conn:       conn,
			err:        err,
			clientFunc: opts.ClientFunc,
		})
	}

	source := testGitserverConns{
		conns: &connection.GitserverConns{
			GitserverAddresses: connection.NewGitserverAddresses(&conf.Unified{
				ServiceConnectionConfig: conftypes.ServiceConnections{
					GitServers: addrs,
				},
			}),
		},
		testAddresses: testAddresses,

		clientFunc: opts.ClientFunc,
	}

	return &source
}

type testGitserverConns struct {
	conns         *connection.GitserverConns
	testAddresses []AddressWithClient

	clientFunc func(conn *grpc.ClientConn) proto.GitserverServiceClient
}

// AddrForRepo returns the gitserver address to use for the given repo name.
func (c *testGitserverConns) AddrForRepo(ctx context.Context, repo api.RepoName) string {
	return c.conns.AddrForRepo(ctx, repo)
}

// Addresses returns the current list of gitserver addresses.
func (c *testGitserverConns) Addresses() []AddressWithClient {
	return c.testAddresses
}

func (c *testGitserverConns) GetAddressWithClient(addr string) AddressWithClient {
	for _, addrClient := range c.testAddresses {
		if addrClient.Address() == addr {
			return addrClient
		}
	}
	return nil
}

// ClientForRepo returns a client or host for the given repo name.
func (c *testGitserverConns) ClientForRepo(ctx context.Context, repo api.RepoName) (proto.GitserverServiceClient, error) {
	addr := c.conns.AddrForRepo(ctx, repo)

	conn, err := defaults.Dial(
		addr,
		log.NoOp(),
	)
	if err != nil {
		return nil, err
	}

	return &errorTranslatingClient{
		base: &automaticRetryClient{
			base: c.clientFunc(conn),
		},
	}, nil
}

type testConnAndErr struct {
	address    string
	conn       *grpc.ClientConn
	err        error
	clientFunc func(conn *grpc.ClientConn) proto.GitserverServiceClient
}

// Address implements AddressWithClient
func (t *testConnAndErr) Address() string {
	return t.address
}

// GRPCClient implements AddressWithClient
func (t *testConnAndErr) GRPCClient() (proto.GitserverServiceClient, error) {
	return &errorTranslatingClient{
		base: &automaticRetryClient{
			base: t.clientFunc(t.conn),
		},
	}, t.err
}

var _ ClientSource = &testGitserverConns{}
var _ AddressWithClient = &testConnAndErr{}

// AddressWithClient is a gitserver address with a client.
type AddressWithClient interface {
	Address() string                                   // returns the address of the endpoint that this GRPC client is targeting
	GRPCClient() (proto.GitserverServiceClient, error) // returns the gRPC client to use to contact the given address
}

type connAndErr struct {
	address string
	conn    *grpc.ClientConn
	err     error
}

func (c *connAndErr) Address() string {
	return c.address
}

func (c *connAndErr) GRPCClient() (proto.GitserverServiceClient, error) {
	return &errorTranslatingClient{
		base: &automaticRetryClient{
			base: proto.NewGitserverServiceClient(c.conn),
		},
	}, c.err
}

var _ AddressWithClient = &connAndErr{}
