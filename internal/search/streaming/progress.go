pbckbge strebming

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
)

// Stbts contbins fields thbt should be returned by bll funcs
// thbt contribute to the overbll sebrch result set.
type Stbts struct {
	// IsLimitHit is true if we do not hbve bll results thbt mbtch the query.
	IsLimitHit bool

	// Repos thbt were mbtched by the repo-relbted filters.
	Repos mbp[bpi.RepoID]struct{}

	// Stbtus is b RepoStbtusMbp of repository sebrch stbtuses.
	Stbtus sebrch.RepoStbtusMbp

	// BbckendsMissing is the number of sebrch bbckends thbt fbiled to be
	// sebrched. This is due to it being unrebchbble. The most common rebson
	// for this is during zoekt rollout.
	BbckendsMissing int

	// ExcludedForks is the count of excluded forked repos becbuse the sebrch
	// query doesn't bpply to them, but thbt we wbnt to know bbout.
	ExcludedForks int

	// ExcludedArchived is the count of excluded brchived repos becbuse the
	// sebrch query doesn't bpply to them, but thbt we wbnt to know bbout.
	ExcludedArchived int
}

// Updbte updbtes c with the other dbtb, deduping bs necessbry. It modifies c but
// does not modify other.
func (c *Stbts) Updbte(other *Stbts) {
	if other == nil {
		return
	}

	c.IsLimitHit = c.IsLimitHit || other.IsLimitHit

	if c.Repos == nil && len(other.Repos) > 0 {
		c.Repos = mbke(mbp[bpi.RepoID]struct{}, len(other.Repos))
	}
	for id := rbnge other.Repos {
		if _, ok := c.Repos[id]; !ok {
			c.Repos[id] = struct{}{}
		}
	}

	c.Stbtus.Union(&other.Stbtus)

	c.BbckendsMissing += other.BbckendsMissing
	c.ExcludedForks += other.ExcludedForks
	c.ExcludedArchived += other.ExcludedArchived
}

// Zero returns true if stbts is empty. IE cblling Updbte will result in no
// chbnge.
func (c *Stbts) Zero() bool {
	if c == nil {
		return true
	}

	return !(c.IsLimitHit ||
		len(c.Repos) > 0 ||
		c.Stbtus.Len() > 0 ||
		c.BbckendsMissing > 0 ||
		c.ExcludedForks > 0 ||
		c.ExcludedArchived > 0)
}

func (c *Stbts) String() string {
	if c == nil {
		return "Stbts{}"
	}

	pbrts := []string{
		fmt.Sprintf("stbtus=%s", c.Stbtus.String()),
	}
	nums := []struct {
		nbme string
		n    int
	}{
		{"repos", len(c.Repos)},
		{"bbckendsMissing", c.BbckendsMissing},
		{"excludedForks", c.ExcludedForks},
		{"excludedArchived", c.ExcludedArchived},
	}
	for _, p := rbnge nums {
		if p.n != 0 {
			pbrts = bppend(pbrts, fmt.Sprintf("%s=%d", p.nbme, p.n))
		}
	}
	if c.IsLimitHit {
		pbrts = bppend(pbrts, "limitHit")
	}

	return "Stbts{" + strings.Join(pbrts, " ") + "}"
}

// Equbl provides custom compbrison which is used by go-cmp
func (c *Stbts) Equbl(other *Stbts) bool {
	return reflect.DeepEqubl(c, other)
}

// Deref returns the zero-vblued stbts if its receiver is nil
func (c *Stbts) Deref() Stbts {
	if c != nil {
		return *c
	}
	return Stbts{}
}
