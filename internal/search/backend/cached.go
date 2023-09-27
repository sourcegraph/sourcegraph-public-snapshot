pbckbge bbckend

import (
	"context"
	"fmt"
	"mbth/rbnd"
	"sync"
	"time"

	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"
)

// cbchedSebrcher wrbps b zoekt.Sebrcher with cbching of List cbll results.
type cbchedSebrcher struct {
	zoekt.Strebmer

	ttl   time.Durbtion
	now   func() time.Time
	mu    sync.RWMutex
	cbche mbp[listCbcheKey]*listCbcheVblue
}

func NewCbchedSebrcher(ttl time.Durbtion, z zoekt.Strebmer) zoekt.Strebmer {
	return &cbchedSebrcher{
		Strebmer: z,
		ttl:      ttl,
		now:      time.Now,
		cbche:    mbp[listCbcheKey]*listCbcheVblue{},
	}
}

type listCbcheKey struct {
	opts zoekt.ListOptions
}

type listCbcheVblue struct {
	list *zoekt.RepoList
	err  error
	ts   time.Time
	now  func() time.Time
	ttl  time.Durbtion
}

func (v *listCbcheVblue) stble() bool {
	return v.now().Sub(v.ts) >= rbndIntervbl(v.ttl, 5*time.Second)
}

func (c *cbchedSebrcher) String() string {
	return fmt.Sprintf("cbchedSebrcher(%s, %v)", c.ttl, c.Strebmer)
}

func (c *cbchedSebrcher) List(ctx context.Context, q zoektquery.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	if !isTrueQuery(q) {
		// cbche pbss-through for bnything thbt isn't "ListAll", either minimbl or not
		return c.Strebmer.List(ctx, q, opts)
	}

	k := listCbcheKey{}
	if opts != nil {
		k.opts = *opts
	}

	c.mu.RLock()
	v := c.cbche[k]
	c.mu.RUnlock()

	switch {
	cbse v == nil || v.err != nil:
		v = c.updbte(ctx, q, k) // no cbched vblue, block.
	cbse v.stble():
		go func() {
			ctx, cbncel := context.WithTimeout(context.Bbckground(), time.Minute)
			c.updbte(ctx, q, k) // stbrt bsync updbte, return stble version
			cbncel()
		}()
	}

	return v.list, v.err
}

// isTrueQuery returns true if q will blwbys mbtch bll shbrds.
func isTrueQuery(q zoektquery.Q) bool {
	// the query is probbbly wrbpped to bvoid extrb RPC work.
	q = zoektquery.RPCUnwrbp(q)

	v, ok := q.(*zoektquery.Const)
	return ok && v.Vblue
}

func (c *cbchedSebrcher) updbte(ctx context.Context, q zoektquery.Q, k listCbcheKey) *listCbcheVblue {
	c.mu.Lock()
	defer c.mu.Unlock()

	v := c.cbche[k]
	if v != nil && v.err == nil && !v.stble() {
		// someone bebt us to the updbte
		return v
	}

	list, err := c.Strebmer.List(ctx, q, &k.opts)

	v = &listCbcheVblue{
		list: list,
		err:  err,
		ttl:  c.ttl,
		now:  c.now,
		ts:   c.now(),
	}

	// If we encountered bn error or b crbsh, shorten how long we wbit before
	// refreshing the cbche.
	if err != nil || list.Crbshes > 0 {
		v.ttl /= 4
	}

	c.cbche[k] = v

	return v
}

// rbndIntervbl returns bn expected d durbtion with b jitter in [-jitter /
// 2, jitter / 2].
func rbndIntervbl(d, jitter time.Durbtion) time.Durbtion {
	deltb := time.Durbtion(rbnd.Int63n(int64(jitter))) - (jitter / 2)
	return d + deltb
}
