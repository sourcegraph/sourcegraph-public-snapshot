package connection

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/grpc/concurrencylimiter"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/limiter"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	addrForRepoInvoked = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_addr_for_repo_invoked",
		Help: "Number of times gitserver.AddrForRepo was invoked",
	})
)

// NewGitserverAddresses fetches the current set of gitserver addresses
// and pinned repos for gitserver.
func NewGitserverAddresses(cfg conftypes.UnifiedQuerier) GitserverAddresses {
	addrs := GitserverAddresses{
		Addresses: cfg.ServiceConnections().GitServers,
	}
	s := cfg.SiteConfig()
	if s.ExperimentalFeatures != nil {
		addrs.PinnedServers = s.ExperimentalFeatures.GitServerPinnedRepos
	}
	return addrs
}

type GitserverAddresses struct {
	// The current list of gitserver addresses
	Addresses []string

	// A list of overrides to pin a repo to a specific gitserver instance. This
	// ensures that, even if the number of gitservers changes, these repos will
	// not be moved.
	PinnedServers map[string]string
}

// AddrForRepo returns the gitserver address to use for the given repo name.
func (g *GitserverAddresses) AddrForRepo(ctx context.Context, repoName api.RepoName) string {
	addrForRepoInvoked.Inc()

	// We undelete the repo name for the addr function so that we can still reach the
	// right gitserver after a repo has been deleted (and the name changed by that).
	// Ideally we wouldn't need this, but as long as we use RepoName as the identifier
	// in gitserver, we have to do this.
	name := string(api.UndeletedRepoName(repoName))
	if pinnedAddr, ok := g.PinnedServers[name]; ok {
		return pinnedAddr
	}

	// We use the normalize function here, because that's what we did previously.
	// Ideally, this would not be required, but it would reshuffle GitHub.com repos
	// with uppercase characters in the name. So until we have a better migration
	// strategy, we keep this old behavior in.
	return addrForKey(string(protocol.NormalizeRepo(api.RepoName(name))), g.Addresses)
}

// addrForKey returns the gitserver address to use for the given string key,
// which is hashed for sharding purposes.
func addrForKey(key string, addrs []string) string {
	sum := md5.Sum([]byte(key))
	serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(addrs))
	return addrs[serverIndex]
}

type GitserverConns struct {
	GitserverAddresses

	// invariant: there is one conn for every gitserver address
	grpcConns map[string]connAndErr
}

func (g *GitserverConns) ConnForRepo(ctx context.Context, repo api.RepoName) (*grpc.ClientConn, error) {
	addr := g.AddrForRepo(ctx, repo)
	ce, ok := g.grpcConns[addr]
	if !ok {
		return nil, errors.Newf("no gRPC connection found for address %q", addr)
	}
	return ce.conn, ce.err
}

// AddressWithConn is a gitserver address with a conn.
type AddressWithConn interface {
	// returns the address of the endpoint that this GRPC client is targeting
	Address() string
	// returns the gRPC connection to use to contact the given address
	GRPCConn() (*grpc.ClientConn, error)
}

type connAndErr struct {
	address string
	conn    *grpc.ClientConn
	err     error
}

func (c *connAndErr) Address() string {
	return c.address
}

func (c *connAndErr) GRPCConn() (*grpc.ClientConn, error) {
	return c.conn, c.err
}

type atomicGitServerConns struct {
	conns     atomic.Pointer[GitserverConns]
	watchOnce sync.Once
}

func (a *atomicGitServerConns) AddrForRepo(ctx context.Context, repo api.RepoName) string {
	return a.get().AddrForRepo(ctx, repo)
}

func (a *atomicGitServerConns) ConnForRepo(ctx context.Context, repo api.RepoName) (*grpc.ClientConn, error) {
	return a.get().ConnForRepo(ctx, repo)
}

func (a *atomicGitServerConns) Addresses() []AddressWithConn {
	conns := a.get()
	addrs := make([]AddressWithConn, 0, len(conns.Addresses))
	for _, addr := range conns.Addresses {
		addrs = append(addrs, &connAndErr{
			address: addr,
			conn:    conns.grpcConns[addr].conn,
			err:     conns.grpcConns[addr].err,
		})
	}
	return addrs
}

func (a *atomicGitServerConns) GetAddressWithConn(addr string) AddressWithConn {
	conns := a.get()
	addrConn, ok := conns.grpcConns[addr]
	if ok {
		return &connAndErr{
			address: addr,
			conn:    addrConn.conn,
			err:     addrConn.err,
		}
	}
	return nil
}

func (a *atomicGitServerConns) get() *GitserverConns {
	a.initOnce()
	return a.conns.Load()
}

func (a *atomicGitServerConns) initOnce() {
	// Initialize lazily because conf.Watch cannot be used during init time.
	a.watchOnce.Do(func() {
		conf.Watch(func() {
			a.update(conf.Get())
		})
	})
}

func (a *atomicGitServerConns) update(cfg *conf.Unified) {
	after := GitserverConns{
		GitserverAddresses: NewGitserverAddresses(cfg),
		grpcConns:          nil, // to be filled in
	}

	before := a.conns.Load()
	if before == nil {
		before = &GitserverConns{}
	}

	if slices.Equal(before.Addresses, after.Addresses) {
		// No change in addresses. Reuse the old connections.
		// We still update newAddrs in case the pinned repos have changed.
		after.grpcConns = before.grpcConns
		a.conns.Store(&after)
		return
	}
	log.Scoped("").Info(
		"new gitserver addresses",
		log.Strings("before", before.Addresses),
		log.Strings("after", after.Addresses),
	)

	// Open connections for each address
	clientLogger := log.Scoped("gitserver.client")

	after.grpcConns = make(map[string]connAndErr, len(after.Addresses))
	for _, addr := range after.Addresses {
		conn, err := defaults.Dial(
			addr,
			clientLogger,
			grpc.WithChainUnaryInterceptor(concurrencylimiter.UnaryClientInterceptor(lim)),
			grpc.WithChainStreamInterceptor(concurrencylimiter.StreamClientInterceptor(lim)),
		)
		after.grpcConns[addr] = connAndErr{conn: conn, err: err}
	}

	a.conns.Store(&after)

	// After making the new conns visible, close the old conns
	for _, ce := range before.grpcConns {
		if ce.err == nil {
			ce.conn.Close()
		}
	}
}

var (
	concurrencyLimit = env.MustGetInt("SRC_GITSERVER_CLIENT_CONCURRENCY_LIMIT", 500, "maximum number of concurrent gitserver RPC calls")
	lim              = &observedLimiter{Limiter: limiter.New(concurrencyLimit)}
)

var concurrentRequestsGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "src_gitserver_client_concurrent_requests",
	Help: "Current number of concurrent requests running against gitserver client.",
})

type observedLimiter struct {
	limiter.Limiter
}

func (l *observedLimiter) Acquire() {
	l.Limiter.Acquire()
	concurrentRequestsGauge.Inc()
}

func (l *observedLimiter) Release() {
	l.Limiter.Release()
	concurrentRequestsGauge.Dec()
}

var _ AddressWithConn = &connAndErr{}
