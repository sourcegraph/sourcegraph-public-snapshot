package gitserver

import (
	"crypto/md5"
	"encoding/binary"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc"
)

var addrForRepoInvoked = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_gitserver_addr_for_repo_invoked",
	Help: "Number of times gitserver.AddrForRepo was invoked",
}, []string{"user_agent"})

// NewGitserverAddressesFromConf fetches the current set of gitserver addresses
// and pinned repos for gitserver.
func NewGitserverAddressesFromConf(cfg *conf.Unified) GitserverAddresses {
	addrs := GitserverAddresses{
		Addresses: cfg.ServiceConnectionConfig.GitServers,
	}
	if cfg.ExperimentalFeatures != nil {
		addrs.PinnedServers = cfg.ExperimentalFeatures.GitServerPinnedRepos
	}
	return addrs
}

func newTestGitserverConns(addrs []string) *GitServerConns {
	conns := make(map[string]connAndErr)
	for _, addr := range addrs {
		conns[addr] = connAndErr{err: errors.New("conns not available in tests")}
	}
	return &GitServerConns{
		GitserverAddresses: GitserverAddresses{
			Addresses: addrs,
		},
		grpcConns: conns,
	}
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
func (g GitserverAddresses) AddrForRepo(userAgent string, repo api.RepoName) string {
	addrForRepoInvoked.WithLabelValues(userAgent).Inc()

	repo = protocol.NormalizeRepo(repo) // in case the caller didn't already normalize it
	rs := string(repo)

	if pinnedAddr, ok := g.PinnedServers[rs]; ok {
		return pinnedAddr
	}

	return addrForKey(rs, g.Addresses)
}

// addrForKey returns the gitserver address to use for the given string key,
// which is hashed for sharding purposes.
func addrForKey(key string, addrs []string) string {
	sum := md5.Sum([]byte(key))
	serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(addrs))
	return addrs[serverIndex]
}

type GitServerConns struct {
	GitserverAddresses
	// invariant: there is one conn for every
	grpcConns map[string]connAndErr
}

func (g *GitServerConns) ConnForRepo(userAgent string, repo api.RepoName) (*grpc.ClientConn, error) {
	addr := g.AddrForRepo(userAgent, repo)
	ce := g.grpcConns[addr]
	return ce.conn, ce.err
}

type connAndErr struct {
	conn *grpc.ClientConn
	err  error
}

type atomicGRPCAddresses struct {
	atomic.Pointer[GitServerConns]
}

func (a *atomicGRPCAddresses) update(cfg *conf.Unified) {
	newAddrs := GitServerConns{
		GitserverAddresses: NewGitserverAddressesFromConf(cfg),
		grpcConns:          make(map[string]connAndErr),
	}

	// Diff the old addresses with the new addresses
	old := a.Load()
	if old == nil {
		// If update is being called for the first time,
		// default to the zero value.
		old = &GitServerConns{}
	}
	added, removed, unchanged := diffStrings(old.Addresses, newAddrs.Addresses)

	// For each address we already had a connection for, reuse that connection
	for _, addr := range unchanged {
		newAddrs.grpcConns[addr] = old.grpcConns[addr]
	}

	// For each new address, open a new connection
	for _, addr := range added {
		conn, err := grpc.Dial(addr, defaults.DialOptions()...)
		newAddrs.grpcConns[addr] = connAndErr{conn, err}
	}

	if len(newAddrs.grpcConns) != len(newAddrs.Addresses) {
		panic("invariant violated: there must be the same number of addresses and conns")
	}

	a.Store(&newAddrs)

	// After we've published the new version, close the old connections
	for _, addr := range removed {
		ce := old.grpcConns[addr]
		if ce.err != nil {
			continue
		}
		ce.conn.Close()
	}
}

func diffStrings(before, after []string) (added, removed, unchanged []string) {
	beforeSet := make(map[string]struct{}, len(before))
	for _, addr := range before {
		beforeSet[addr] = struct{}{}
	}

	for _, addr := range after {
		_, ok := beforeSet[addr]
		if ok {
			unchanged = append(unchanged, addr)
			delete(beforeSet, addr)
		} else {
			added = append(added, addr)
		}
	}

	for addr := range beforeSet {
		removed = append(removed, addr)
	}

	return added, removed, unchanged
}
