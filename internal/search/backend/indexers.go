package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// EndpointMap is the subset of endpoint.Map (consistent hashmap) methods we
// use. Declared as an interface for testing.
type EndpointMap interface {
	// Endpoints returns a list of all addresses which can be searched. Do not modify the returned value.
	Endpoints() ([]string, error)
	// Get returns the endpoint for the key. (consistent hashing).
	Get(string) (string, error)
}

// Indexers provides methods over the set of indexed-search servers in a
// Sourcegraph cluster.
type Indexers struct {
	// Map is the desired mapping from repository names to endpoints.
	Map EndpointMap

	// Indexed returns a set of repository names currently indexed on
	// endpoint. If indexed fails, it is expected to return an empty set.
	Indexed func(ctx context.Context, endpoint string) zoekt.ReposMap
}

// ReposSubset returns the subset of repoNames that hostname should index.
//
// ReposSubset reuses the underlying array of repoNames.
//
// indexed is the set of repositories currently indexed by hostname.
//
// An error is returned if hostname is not part of the Indexers endpoints.
func (c *Indexers) ReposSubset(ctx context.Context, hostname string, indexed zoekt.ReposMap, repos []types.MinimalRepo) ([]types.MinimalRepo, error) {
	if !c.Enabled() {
		return repos, nil
	}

	eps, err := c.Map.Endpoints()
	if err != nil {
		return nil, err
	}

	endpoint, err := findEndpoint(eps, hostname)
	if err != nil {
		return nil, err
	}

	// Rebalancing: Other contains all repositories endpoint has indexed which
	// it should drop. We will only drop them if the assigned endpoint has
	// indexed it. This is to prevent dropping a computed index until
	// rebalancing is finished.
	other := map[string][]types.MinimalRepo{}

	subset := repos[:0]
	for _, r := range repos {
		assigned, err := c.Map.Get(string(r.Name))
		if err != nil {
			return nil, err
		}

		if assigned == endpoint {
			subset = append(subset, r)
		} else if _, ok := indexed[uint32(r.ID)]; ok {
			other[assigned] = append(other[assigned], r)
		}
	}

	// Only include repos from other if the assigned endpoint has not yet
	// indexed the repository.
	for assigned, otherRepos := range other {
		drop := c.Indexed(ctx, assigned)
		for _, r := range otherRepos {
			if _, ok := drop[uint32(r.ID)]; !ok {
				subset = append(subset, r)
			}
		}
	}

	return subset, nil
}

// Enabled returns true if this feature is enabled. At first horizontal
// sharding will be disabled, if so the functions here fallback to single
// shard behaviour.
func (c *Indexers) Enabled() bool {
	return c.Map != nil
}

// findEndpoint returns the endpoint in eps which matches hostname.
func findEndpoint(eps []string, hostname string) (string, error) {
	// The hostname can be a less qualified hostname. For example in k8s
	// $HOSTNAME will be "indexed-search-0", but to access the pod you will
	// need to specify the endpoint address
	// "indexed-search-0.indexed-search".
	//
	// Additionally an endpoint can also specify a port, which a hostname
	// won't.
	//
	// Given this looser matching, we ensure we don't match more than one
	// endpoint.
	endpoint := ""
	for _, ep := range eps {
		if !strings.HasPrefix(ep, hostname) {
			continue
		}
		if len(hostname) < len(ep) {
			if c := ep[len(hostname)]; c != '.' && c != ':' {
				continue
			}
		}

		if endpoint != "" {
			return "", errors.Errorf("hostname %q matches multiple in %s", hostname, endpointsString(eps))
		}
		endpoint = ep
	}
	if endpoint != "" {
		return endpoint, nil
	}

	return "", errors.Errorf("hostname %q not found in %s", hostname, endpointsString(eps))
}

// endpointsString creates a user readable String for an endpoint map.
func endpointsString(eps []string) string {
	var b strings.Builder
	b.WriteString("Endpoints{")
	for i, k := range eps {
		if i != 0 {
			b.WriteByte(' ')
		}
		_, _ = fmt.Fprintf(&b, "%q", k)
	}
	b.WriteByte('}')
	return b.String()
}
