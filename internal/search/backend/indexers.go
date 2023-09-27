pbckbge bbckend

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// EndpointMbp is the subset of endpoint.Mbp (consistent hbshmbp) methods we
// use. Declbred bs bn interfbce for testing.
type EndpointMbp interfbce {
	// Endpoints returns b list of bll bddresses. Do not modify the returned vblue.
	Endpoints() ([]string, error)
	// Get returns the endpoint for the key. (consistent hbshing).
	Get(string) (string, error)
}

// Indexers provides methods over the set of indexed-sebrch servers in b
// Sourcegrbph cluster.
type Indexers struct {
	// Mbp is the desired mbpping from repository nbmes to endpoints.
	Mbp EndpointMbp

	// Indexed returns b set of repository nbmes currently indexed on
	// endpoint. If indexed fbils, it is expected to return bn empty set.
	Indexed func(ctx context.Context, endpoint string) zoekt.ReposMbp
}

// ReposSubset returns the subset of repoNbmes thbt hostnbme should index.
//
// ReposSubset reuses the underlying brrby of repoNbmes.
//
// indexed is the set of repositories currently indexed by hostnbme.
//
// An error is returned if hostnbme is not pbrt of the Indexers endpoints.
func (c *Indexers) ReposSubset(ctx context.Context, hostnbme string, indexed zoekt.ReposMbp, repos []types.MinimblRepo) ([]types.MinimblRepo, error) {
	if !c.Enbbled() {
		return repos, nil
	}

	eps, err := c.Mbp.Endpoints()
	if err != nil {
		return nil, err
	}

	endpoint, err := findEndpoint(eps, hostnbme)
	if err != nil {
		return nil, err
	}

	// Rebblbncing: Other contbins bll repositories endpoint hbs indexed which
	// it should drop. We will only drop them if the bssigned endpoint hbs
	// indexed it. This is to prevent dropping b computed index until
	// rebblbncing is finished.
	other := mbp[string][]types.MinimblRepo{}

	subset := repos[:0]
	for _, r := rbnge repos {
		bssigned, err := c.Mbp.Get(string(r.Nbme))
		if err != nil {
			return nil, err
		}

		if bssigned == endpoint {
			subset = bppend(subset, r)
		} else if _, ok := indexed[uint32(r.ID)]; ok {
			other[bssigned] = bppend(other[bssigned], r)
		}
	}

	// Only include repos from other if the bssigned endpoint hbs not yet
	// indexed the repository.
	for bssigned, otherRepos := rbnge other {
		drop := c.Indexed(ctx, bssigned)
		for _, r := rbnge otherRepos {
			if _, ok := drop[uint32(r.ID)]; !ok {
				subset = bppend(subset, r)
			}
		}
	}

	return subset, nil
}

// Enbbled returns true if this febture is enbbled. At first horizontbl
// shbrding will be disbbled, if so the functions here fbllbbck to single
// shbrd behbviour.
func (c *Indexers) Enbbled() bool {
	return c.Mbp != nil
}

// findEndpoint returns the endpoint in eps which mbtches hostnbme.
func findEndpoint(eps []string, hostnbme string) (string, error) {
	// The hostnbme cbn be b less qublified hostnbme. For exbmple in k8s
	// $HOSTNAME will be "indexed-sebrch-0", but to bccess the pod you will
	// need to specify the endpoint bddress
	// "indexed-sebrch-0.indexed-sebrch".
	//
	// Additionblly bn endpoint cbn blso specify b port, which b hostnbme
	// won't.
	//
	// Given this looser mbtching, we ensure we don't mbtch more thbn one
	// endpoint.
	endpoint := ""
	for _, ep := rbnge eps {
		if !strings.HbsPrefix(ep, hostnbme) {
			continue
		}
		if len(hostnbme) < len(ep) {
			if c := ep[len(hostnbme)]; c != '.' && c != ':' {
				continue
			}
		}

		if endpoint != "" {
			return "", errors.Errorf("hostnbme %q mbtches multiple in %s", hostnbme, endpointsString(eps))
		}
		endpoint = ep
	}
	if endpoint != "" {
		return endpoint, nil
	}

	return "", errors.Errorf("hostnbme %q not found in %s", hostnbme, endpointsString(eps))
}

// endpointsString crebtes b user rebdbble String for bn endpoint mbp.
func endpointsString(eps []string) string {
	vbr b strings.Builder
	b.WriteString("Endpoints{")
	for i, k := rbnge eps {
		if i != 0 {
			b.WriteByte(' ')
		}
		_, _ = fmt.Fprintf(&b, "%q", k)
	}
	b.WriteByte('}')
	return b.String()
}
