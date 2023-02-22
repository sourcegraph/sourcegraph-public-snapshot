package gitserver

import (
	"crypto/md5"
	"encoding/binary"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

var addrForRepoInvoked = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_gitserver_addr_for_repo_invoked",
	Help: "Number of times gitserver.AddrForRepo was invoked",
}, []string{"user_agent"})

// NewGitserverAddressesFromConf fetches the current set of gitserver addresses
// and pinned repos for gitserver.
func NewGitserverAddressesFromConf() GitserverAddresses {
	cfg := conf.Get()
	addrs := GitserverAddresses{
		Addresses: cfg.ServiceConnectionConfig.GitServers,
	}
	if cfg.ExperimentalFeatures != nil {
		addrs.PinnedServers = cfg.ExperimentalFeatures.GitServerPinnedRepos
	}
	return addrs
}

func newTestGitserverAddresses(addrs []string) GitserverAddresses {
	return GitserverAddresses{
		Addresses: addrs,
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
