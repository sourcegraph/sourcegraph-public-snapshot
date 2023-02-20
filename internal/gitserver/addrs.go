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

type GitServerAddresses struct {
	Addresses     []string
	PinnedServers map[string]string
}

func (g *GitServerAddresses) AddrForRepo(userAgent string, repo api.RepoName) string {
	repo = protocol.NormalizeRepo(repo) // in case the caller didn't already normalize it
	rs := string(repo)

	if pinnedAddr, ok := g.PinnedServers[rs]; ok {
		return pinnedAddr
	}

	return addrForKey(rs, g.Addresses)
}

func newConfPicker() addrPicker {
	cp := &atomicPicker{}
	conf.Watch(func() {
		cp.update(conf.Get())
	})
	return cp
}

func newTestPicker(addrs []string) addrPicker {
	cp := &atomicPicker{}

	errTest := errors.New("grpc connections not available in test picker")
	conns := make(map[string]connErr, len(addrs))
	for _, addr := range addrs {
		conns[addr] = connErr{err: errTest}
	}

	cp.cur.Store(&addrsAndConns{
		GitServerAddresses: GitServerAddresses{
			Addresses:     addrs,
			PinnedServers: nil,
		},
		conns: conns,
	})

	return cp
}

type addrPicker interface {
	addrs() []string
	addrForRepo(userAgent string, repo api.RepoName) string
	connForRepo(userAgent string, repo api.RepoName) (*grpc.ClientConn, error)
}

type atomicPicker struct {
	cur atomic.Pointer[addrsAndConns]
}

type addrsAndConns struct {
	GitServerAddresses

	// invariant: exactly one connection per addr
	conns map[string]connErr
}

type connErr struct {
	conn *grpc.ClientConn
	err  error
}

func (c *atomicPicker) addrs() []string {
	return c.cur.Load().Addresses
}

func (c *atomicPicker) addrForRepo(userAgent string, repo api.RepoName) string {
	set := c.cur.Load()
	return set.AddrForRepo(userAgent, repo)
}

func (c *atomicPicker) connForRepo(userAgent string, repo api.RepoName) (*grpc.ClientConn, error) {
	set := c.cur.Load()

	addr := set.AddrForRepo(userAgent, repo)
	ce := set.conns[addr]
	return ce.conn, ce.err
}

func (c *atomicPicker) update(cfg *conf.Unified) {
	newAddrs := cfg.ServiceConnections().GitServers
	newPinned := map[string]string(nil)
	if cfg.ExperimentalFeatures != nil {
		newPinned = cfg.ExperimentalFeatures.GitServerPinnedRepos
	}

	oldAddrSet := c.cur.Load()

	// Diff the old and new addresses
	added, removed, unchanged := diffStrings(oldAddrSet.Addresses, newAddrs)
	newConns := make(map[string]connErr, len(newAddrs))

	// For each address that hasn't changed, reuse the gRPC connection
	for _, addr := range unchanged {
		newConns[addr] = oldAddrSet.conns[addr]
	}

	// For each new address, open a new gRPC connection
	for _, addr := range added {
		conn, err := grpc.Dial(addr, defaults.DialOptions()...)
		newConns[addr] = connErr{conn: conn, err: err}
	}

	if len(newAddrs) != len(newConns) {
		panic("violated invariant: addrs is not the same size as conns")
	}

	// Store the updated set of addrs in the picker
	newSet := addrsAndConns{
		GitServerAddresses: GitServerAddresses{
			Addresses:     newAddrs,
			PinnedServers: newPinned,
		},
		conns: newConns,
	}
	c.cur.Store(&newSet)

	// After storing the updated value, close the old connections
	for _, addr := range removed {
		c := oldAddrSet.conns[addr]
		if c.err != nil {
			// Skip the connection if opening it failed
			continue
		}
		c.conn.Close()
	}
}

func diffStrings(before, after []string) (added, removed, unchanged []string) {
	beforeSet := make(map[string]struct{})
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

// addrForKey returns the gitserver address to use for the given string key,
// which is hashed for sharding purposes.
func addrForKey(key string, addrs []string) string {
	sum := md5.Sum([]byte(key))
	serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(addrs))
	return addrs[serverIndex]
}
