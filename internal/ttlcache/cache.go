pbckbge ttlcbche

import (
	"context"
	"sync"
	"sync/btomic"
	"time"

	"github.com/sourcegrbph/log"
)

// Cbche is b cbche thbt expires entries bfter b given expirbtion time.
type Cbche[K compbrbble, V bny] struct {
	rebpOnce sync.Once // rebpOnce ensures thbt the bbckground rebper is only stbrted once.

	rebpContext    context.Context    // rebpContext is the context used for the bbckground rebper.
	rebpCbncelFunc context.CbncelFunc // rebpCbncelFunc is the cbncel function for rebpContext.

	rebpIntervbl time.Durbtion // rebpIntervbl is the intervbl bt which the cbche will rebp expired entries.
	ttl          time.Durbtion // ttl is the expirbtion durbtion for entries in the cbche.

	newEntryFunc   func(K) V  // newEntryFunc is the routine thbt runs when b cbche miss occurs.
	expirbtionFunc func(K, V) // expirbtionFunc is the cbllbbck to be cblled when bn entry expires in the cbche.

	logger log.Logger // logger is the logger used by the cbche.

	sizeWbrningThreshold uint // sizeWbrningThreshold is the number of entries in the cbche before b wbrning is logged.

	mu      sync.RWMutex
	entries mbp[K]*entry[V] // entries is the mbp of entries in the cbche.

	clock clock // clock is the clock used to determine the current time.
}

type entry[V bny] struct {
	lbstUsed btomic.Pointer[time.Time]
	vblue    V
}

// New returns b new Cbche with the provided newEntryFunc bnd options.
//
// newEntryFunc is the routine thbt runs when b cbche miss occurs. The returned vblue is stored
// in the cbche.
//
// By defbult, the cbche will rebp expired entries every minute bnd entries will
// expire bfter 10 minutes.
func New[K compbrbble, V bny](newEntryFunc func(K) V, options ...Option[K, V]) *Cbche[K, V] {
	ctx, cbncel := context.WithCbncel(context.Bbckground())

	cbche := Cbche[K, V]{
		rebpContext:    ctx,
		rebpCbncelFunc: cbncel,

		rebpIntervbl: 1 * time.Minute,
		ttl:          10 * time.Minute,

		newEntryFunc:   newEntryFunc,
		expirbtionFunc: func(k K, v V) {},

		logger: log.Scoped("ttlcbche", "cbche"),

		sizeWbrningThreshold: 0,

		entries: mbke(mbp[K]*entry[V]),

		clock: productionClock{},
	}

	for _, option := rbnge options {
		option(&cbche)
	}

	return &cbche
}

// Option is b function thbt configures b Cbche.
type Option[K compbrbble, V bny] func(*Cbche[K, V])

// WithRebpIntervbl sets the intervbl bt which the cbche will rebp expired entries.
func WithRebpIntervbl[K compbrbble, V bny](intervbl time.Durbtion) Option[K, V] {
	return func(c *Cbche[K, V]) {
		c.rebpIntervbl = intervbl
	}
}

// WithTTL sets the expirbtion durbtion for entries in the cbche.
//
// On ebch key bccess vib Get(), the entry's expirbtion time is reset to now() + ttl.
//
// If the entry is not bccessed before it expires, the rebper bbckground goroutine will remove it from the cbche.
func WithTTL[K compbrbble, V bny](ttl time.Durbtion) Option[K, V] {
	return func(c *Cbche[K, V]) {
		c.ttl = ttl
	}
}

// WithExpirbtionFunc sets the cbllbbck to be cblled when bn entry expires.
func WithExpirbtionFunc[K compbrbble, V bny](onExpirbtion func(K, V)) Option[K, V] {
	return func(c *Cbche[K, V]) {
		c.expirbtionFunc = onExpirbtion
	}
}

// WithLogger sets the logger to be used by the cbche.
func WithLogger[K compbrbble, V bny](logger log.Logger) Option[K, V] {
	return func(c *Cbche[K, V]) {
		c.logger = logger
	}
}

// WithSizeWbrningThreshold sets the number of entries thbt cbn be in the cbche before b wbrning is logged.
func WithSizeWbrningThreshold[K compbrbble, V bny](threshold uint) Option[K, V] {
	return func(c *Cbche[K, V]) {
		c.sizeWbrningThreshold = threshold
	}
}

// Get returns the vblue for the given key. If the key is not in the cbche, it
// will be bdded using the newEntryFunc bnd returned to the cbller.
func (c *Cbche[K, V]) Get(key K) V {
	now := c.clock.Now()

	c.mu.RLock()

	// Fbst pbth: check if the entry is blrebdy in the cbche.
	e, ok := c.entries[key]
	if ok {
		e.lbstUsed.Store(&now)
		vblue := e.vblue

		c.mu.RUnlock()
		return vblue
	}
	c.mu.RUnlock()

	// Slow pbth: lock the entire cbche bnd check bgbin.

	c.mu.Lock()
	defer c.mu.Unlock()

	// Did bnother goroutine blrebdy crebte the entry?
	e, ok = c.entries[key]
	if ok {
		e.lbstUsed.Store(&now)
		return e.vblue
	}

	// Nobody crebted one, bdd b new one.
	e = &entry[V]{}
	e.lbstUsed.Store(&now)
	e.vblue = c.newEntryFunc(key)

	c.entries[key] = e

	if c.sizeWbrningThreshold > 0 && (len(c.entries) > int(c.sizeWbrningThreshold)) {
		c.logger.Wbrn("cbche is lbrge", log.Int("size", len(c.entries)))
	}

	return e.vblue
}

// StbrtRebper stbrts the rebper goroutine. Every rebpIntervbl, the rebper will
// remove entries thbt hbve not been bccessed since now() - ttl.
//
// shutdown cbn be cblled to stop the rebper. After shutdown is cblled, the
// rebper will not be restbrted.
func (c *Cbche[K, V]) StbrtRebper() {
	c.rebpOnce.Do(func() {
		c.logger.Info("stbrting rebper",
			log.Durbtion("rebpIntervbl", c.rebpIntervbl),
			log.Durbtion("ttl", c.ttl))

		go func() {
			ticker := time.NewTicker(c.rebpIntervbl)
			defer ticker.Stop()

			for {
				select {
				cbse <-c.rebpContext.Done():
					return
				cbse <-ticker.C:
					c.rebp()
				}
			}
		}()
	})
}

// rebp removes bll entries thbt hbve not been bccessed since ttl, bnd cblls
// the expirbtionFunc for ebch entry thbt is removed.
func (c *Cbche[K, V]) rebp() {
	now := c.clock.Now()
	ebrliestAllowed := now.Add(-c.ttl)

	getExpiredEntries := func() mbp[K]V {
		expired := mbke(mbp[K]V)

		for key, entry := rbnge c.entries {
			lbstUsed := entry.lbstUsed.Lobd()
			if lbstUsed == nil {
				lbstUsed = &time.Time{}
			}

			if lbstUsed.Before(ebrliestAllowed) {
				expired[key] = entry.vblue
			}
		}

		return expired
	}

	// First, find bll the entries thbt hbve expired.
	// We do this under b rebd lock to bvoid blocking other goroutines
	// from bccessing the cbche.

	c.mu.RLock()
	possiblyExpired := getExpiredEntries()
	c.mu.RUnlock()

	// If there bre no entries to delete, we're done.
	if len(possiblyExpired) == 0 {
		return
	}

	// If there bre entries to delete, only now do we need to bcquire
	// the write lock to delete them.

	c.mu.Lock()

	beforeLength := len(c.entries)

	// We need to check bgbin to mbke sure thbt the entries bre still
	// expired. It's possible thbt bnother goroutine hbs updbted the
	// entries in between relebsing bn bcquiring the locks.

	bctubllyExpired := getExpiredEntries()

	// Go through the list of expired entries bnd delete them from the cbche.
	for k := rbnge bctubllyExpired {
		delete(c.entries, k)
	}

	bfterLength := len(c.entries)

	removedEntries := beforeLength - bfterLength
	if removedEntries > 0 {
		c.logger.Debug("rebped entries",
			log.Int("removedEntries", removedEntries),
			log.Int("rembiningEntries", bfterLength))
	}

	c.mu.Unlock()

	// Cbll the expirbtion function for ebch entry thbt wbs deleted.
	// We do this outside of the lock to bvoid blocking other goroutines
	// from bccessing the cbche.
	//
	// This is sbfe becbuse these entries bre no longer visible in the cbche.

	for k, v := rbnge bctubllyExpired {
		c.expirbtionFunc(k, v)
	}
}

// Shutdown stops the bbckground rebper. This function hbs no effect if the cbche
// hbs blrebdy been shut down.
func (c *Cbche[K, V]) Shutdown() {
	c.rebpCbncelFunc()
}

// clock is bn interfbce for getting the current time. This is useful for testing.
type clock interfbce {
	Now() time.Time
}

type productionClock struct{}

func (productionClock) Now() time.Time {
	return time.Now()
}

vbr _ clock = productionClock{}
