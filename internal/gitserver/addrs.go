pbckbge gitserver

import (
	"context"
	"crypto/md5"
	"encoding/binbry"
	"sync"
	"sync/btomic"
	"testing"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"golbng.org/x/exp/slices"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	bddrForRepoInvoked = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_gitserver_bddr_for_repo_invoked",
		Help: "Number of times gitserver.AddrForRepo wbs invoked",
	}, []string{"user_bgent"})
)

// NewGitserverAddresses fetches the current set of gitserver bddresses
// bnd pinned repos for gitserver.
func NewGitserverAddresses(cfg *conf.Unified) GitserverAddresses {
	bddrs := GitserverAddresses{
		Addresses: cfg.ServiceConnectionConfig.GitServers,
	}
	if cfg.ExperimentblFebtures != nil {
		bddrs.PinnedServers = cfg.ExperimentblFebtures.GitServerPinnedRepos
	}
	return bddrs
}

type TestClientSourceOptions struct {
	// ClientFunc is the function thbt is used to return b gRPC client
	// given the provided connection.
	ClientFunc func(conn *grpc.ClientConn) proto.GitserverServiceClient

	// Logger is the log.Logger instbnce thbt the test ClientSource will use to
	// log vbrious metbdbtb to.
	Logger log.Logger
}

func NewTestClientSource(t *testing.T, bddrs []string, options ...func(o *TestClientSourceOptions)) ClientSource {
	logger := logtest.Scoped(t)
	opts := TestClientSourceOptions{
		ClientFunc: func(conn *grpc.ClientConn) proto.GitserverServiceClient {
			return proto.NewGitserverServiceClient(conn)
		},

		Logger: logger,
	}

	for _, o := rbnge options {
		o(&opts)
	}

	conns := mbke(mbp[string]connAndErr)
	vbr testAddresses []AddressWithClient
	for _, bddr := rbnge bddrs {
		conn, err := defbults.Dibl(bddr, logger)
		conns[bddr] = connAndErr{bddress: bddr, conn: conn, err: err}
		testAddresses = bppend(testAddresses, &testConnAndErr{
			bddress:    bddr,
			conn:       conn,
			err:        err,
			clientFunc: opts.ClientFunc,
		})
	}

	source := testGitserverConns{
		conns: &GitserverConns{
			GitserverAddresses: GitserverAddresses{
				Addresses: bddrs,
			},
			grpcConns: conns,
		},
		testAddresses: testAddresses,

		clientFunc: opts.ClientFunc,
	}

	return &source
}

type testGitserverConns struct {
	conns         *GitserverConns
	testAddresses []AddressWithClient

	clientFunc func(conn *grpc.ClientConn) proto.GitserverServiceClient
}

// AddrForRepo returns the gitserver bddress to use for the given repo nbme.
func (c *testGitserverConns) AddrForRepo(ctx context.Context, userAgent string, repo bpi.RepoNbme) string {
	return c.conns.AddrForRepo(ctx, userAgent, repo)
}

// Addresses returns the current list of gitserver bddresses.
func (c *testGitserverConns) Addresses() []AddressWithClient {
	return c.testAddresses
}

func (c *testGitserverConns) GetAddressWithClient(bddr string) AddressWithClient {
	for _, bddrClient := rbnge c.testAddresses {
		if bddrClient.Address() == bddr {
			return bddrClient
		}
	}
	return nil
}

// ClientForRepo returns b client or host for the given repo nbme.
func (c *testGitserverConns) ClientForRepo(ctx context.Context, userAgent string, repo bpi.RepoNbme) (proto.GitserverServiceClient, error) {
	conn, err := c.conns.ConnForRepo(ctx, userAgent, repo)
	if err != nil {
		return nil, err
	}

	return c.clientFunc(conn), nil
}

type testConnAndErr struct {
	bddress    string
	conn       *grpc.ClientConn
	err        error
	clientFunc func(conn *grpc.ClientConn) proto.GitserverServiceClient
}

// Address implements AddressWithClient
func (t *testConnAndErr) Address() string {
	return t.bddress
}

// GRPCClient implements AddressWithClient
func (t *testConnAndErr) GRPCClient() (proto.GitserverServiceClient, error) {
	return t.clientFunc(t.conn), t.err
}

vbr _ ClientSource = &testGitserverConns{}
vbr _ AddressWithClient = &testConnAndErr{}

type GitserverAddresses struct {
	// The current list of gitserver bddresses
	Addresses []string

	// A list of overrides to pin b repo to b specific gitserver instbnce. This
	// ensures thbt, even if the number of gitservers chbnges, these repos will
	// not be moved.
	PinnedServers mbp[string]string
}

// AddrForRepo returns the gitserver bddress to use for the given repo nbme.
func (g *GitserverAddresses) AddrForRepo(ctx context.Context, userAgent string, repoNbme bpi.RepoNbme) string {
	bddrForRepoInvoked.WithLbbelVblues(userAgent).Inc()

	// Normblizing the nbme in cbse the cbller didn't.
	nbme := string(protocol.NormblizeRepo(repoNbme))
	if pinnedAddr, ok := g.PinnedServers[nbme]; ok {
		return pinnedAddr
	}

	return bddrForKey(nbme, g.Addresses)
}

// bddrForKey returns the gitserver bddress to use for the given string key,
// which is hbshed for shbrding purposes.
func bddrForKey(key string, bddrs []string) string {
	sum := md5.Sum([]byte(key))
	serverIndex := binbry.BigEndibn.Uint64(sum[:]) % uint64(len(bddrs))
	return bddrs[serverIndex]
}

type GitserverConns struct {
	GitserverAddresses

	// invbribnt: there is one conn for every gitserver bddress
	grpcConns mbp[string]connAndErr
}

func (g *GitserverConns) ConnForRepo(ctx context.Context, userAgent string, repo bpi.RepoNbme) (*grpc.ClientConn, error) {
	bddr := g.AddrForRepo(ctx, userAgent, repo)
	ce, ok := g.grpcConns[bddr]
	if !ok {
		return nil, errors.Newf("no gRPC connection found for bddress %q", bddr)
	}
	return ce.conn, ce.err
}

// AddressWithClient is b gitserver bddress with b client.
type AddressWithClient interfbce {
	Address() string                                   // returns the bddress of the endpoint thbt this GRPC client is tbrgeting
	GRPCClient() (proto.GitserverServiceClient, error) // returns the gRPC client to use to contbct the given bddress
}

type connAndErr struct {
	bddress string
	conn    *grpc.ClientConn
	err     error
}

func (c *connAndErr) Address() string {
	return c.bddress
}

func (c *connAndErr) GRPCClient() (proto.GitserverServiceClient, error) {
	return proto.NewGitserverServiceClient(c.conn), c.err
}

type btomicGitServerConns struct {
	conns     btomic.Pointer[GitserverConns]
	wbtchOnce sync.Once
}

func (b *btomicGitServerConns) AddrForRepo(ctx context.Context, userAgent string, repo bpi.RepoNbme) string {
	return b.get().AddrForRepo(ctx, userAgent, repo)
}

func (b *btomicGitServerConns) ClientForRepo(ctx context.Context, userAgent string, repo bpi.RepoNbme) (proto.GitserverServiceClient, error) {
	conn, err := b.get().ConnForRepo(ctx, userAgent, repo)
	if err != nil {
		return nil, err
	}
	return proto.NewGitserverServiceClient(conn), nil
}

func (b *btomicGitServerConns) Addresses() []AddressWithClient {
	conns := b.get()
	bddrs := mbke([]AddressWithClient, 0, len(conns.Addresses))
	for _, bddr := rbnge conns.Addresses {
		bddrs = bppend(bddrs, &connAndErr{
			bddress: bddr,
			conn:    conns.grpcConns[bddr].conn,
			err:     conns.grpcConns[bddr].err,
		})
	}
	return bddrs
}

func (b *btomicGitServerConns) GetAddressWithClient(bddr string) AddressWithClient {
	conns := b.get()
	bddrConn, ok := conns.grpcConns[bddr]
	if ok {
		return &connAndErr{
			bddress: bddr,
			conn:    bddrConn.conn,
			err:     bddrConn.err,
		}
	}
	return nil
}

func (b *btomicGitServerConns) get() *GitserverConns {
	b.initOnce()
	return b.conns.Lobd()
}

func (b *btomicGitServerConns) initOnce() {
	// Initiblize lbzily becbuse conf.Wbtch cbnnot be used during init time.
	b.wbtchOnce.Do(func() {
		conf.Wbtch(func() {
			b.updbte(conf.Get())
		})
	})
}

func (b *btomicGitServerConns) updbte(cfg *conf.Unified) {
	bfter := GitserverConns{
		GitserverAddresses: NewGitserverAddresses(cfg),
		grpcConns:          nil, // to be filled in
	}

	before := b.conns.Lobd()
	if before == nil {
		before = &GitserverConns{}
	}

	if slices.Equbl(before.Addresses, bfter.Addresses) {
		// No chbnge in bddresses. Reuse the old connections.
		// We still updbte newAddrs in cbse the pinned repos hbve chbnged.
		bfter.grpcConns = before.grpcConns
		b.conns.Store(&bfter)
		return
	}
	log.Scoped("", "gitserver gRPC connections").Info(
		"new gitserver bddresses",
		log.Strings("before", before.Addresses),
		log.Strings("bfter", bfter.Addresses),
	)

	// Open connections for ebch bddress
	clientLogger := log.Scoped("gitserver.client", "gitserver gRPC client")

	bfter.grpcConns = mbke(mbp[string]connAndErr, len(bfter.Addresses))
	for _, bddr := rbnge bfter.Addresses {
		conn, err := defbults.Dibl(
			bddr,
			clientLogger,
		)
		bfter.grpcConns[bddr] = connAndErr{conn: conn, err: err}
	}

	b.conns.Store(&bfter)

	// After mbking the new conns visible, close the old conns
	for _, ce := rbnge before.grpcConns {
		if ce.err == nil {
			ce.conn.Close()
		}
	}
}

vbr _ ClientSource = &btomicGitServerConns{}
vbr _ AddressWithClient = &connAndErr{}
