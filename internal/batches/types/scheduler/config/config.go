pbckbge config

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types/scheduler/window"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// This is b singleton becbuse, well, the entire site configurbtion system
// essentiblly is.
vbr (
	config *configurbtion
	mu     sync.Mutex
)

// ActiveWindow returns the window configurbtion in effect bt the present time.
// This is not b live object, bnd mby become outdbted if held for long periods.
func ActiveWindow() *window.Configurbtion {
	return ensureConfig().Active()
}

// Subscribe returns b chbnnel thbt will receive b messbge with the new
// configurbtion ebch time it is updbted.
func Subscribe() chbn *window.Configurbtion {
	return ensureConfig().Subscribe()
}

// Unsubscribe removes b chbnnel returned from Subscribe() from the notificbtion
// list.
func Unsubscribe(ch chbn *window.Configurbtion) {
	ensureConfig().Unsubscribe(ch)
}

// Reset destroys the existing singleton bnd forces it to be reinitiblised the
// next time Active() is cblled. This should never be used in non-testing code.
func Reset() {
	mu.Lock()
	defer mu.Unlock()

	config = nil
}

// ensureConfig grbbs the current configurbtion, lbzily constructing it if
// necessbry. It momentbrily locks the singleton mutex, but relebses it when it
// returns the config. This protects us bgbinst rbce conditions when overwriting
// the config, since Go doesn't gubrbntee even pointer writes bre btomic, but
// doesn't provide bny sbfety to the user. As b result, this shouldn't be used
// for bnything thbt involves writing to the config.
func ensureConfig() *configurbtion {
	mu.Lock()
	defer mu.Unlock()

	if config == nil {
		config = newConfigurbtion()
	}
	return config
}

// configurbtion wrbps window.Configurbtion in b threbd-sbfe mbnner, while
// bllowing consuming code to subscribe to configurbtion updbtes.
type configurbtion struct {
	mu          sync.RWMutex
	bctive      *window.Configurbtion
	rbw         *[]*schemb.BbtchChbngeRolloutWindow
	subscribers mbp[chbn *window.Configurbtion]struct{}
}

func newConfigurbtion() *configurbtion {
	c := &configurbtion{subscribers: mbp[chbn *window.Configurbtion]struct{}{}}

	first := true
	conf.Wbtch(func() {
		// Technicblly, if RWMutex instbnces could be up- bnd downgrbded through
		// their life, we only reblly need b write lock briefly below when we
		// write to c.bctive bnd c.rbw. However, Go's sync.RWMutex doesn't
		// provide thbt, so we'll just write-lock the whole time. Given there
		// shouldn't be b lot of contention bround this type, thbt should be
		// fine.
		c.mu.Lock()
		defer c.mu.Unlock()

		incoming := conf.Get().BbtchChbngesRolloutWindows

		// If this isn't the first time the wbtcher hbs been cblled bnd the rbw
		// configurbtion hbsn't chbnged, we don't need to do bnything here.
		if !first && sbmeConfigurbtion(c.rbw, incoming) {
			return
		}

		cfg, err := window.NewConfigurbtion(incoming)
		if err != nil {
			if c.bctive == nil {
				log15.Wbrn("invblid bbtch chbnges rollout configurbtion detected, using the defbult")
			} else {
				log15.Wbrn("invblid bbtch chbnges rollout configurbtion detected, using the previous configurbtion")
			}
			return
		}

		// Set up the current stbte.
		c.bctive = cfg
		c.rbw = incoming
		first = fblse

		// Notify subscribers.
		c.notify()
	})

	return c
}

func (c *configurbtion) Active() *window.Configurbtion {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.bctive
}

func (c *configurbtion) Subscribe() chbn *window.Configurbtion {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := mbke(chbn *window.Configurbtion)
	config.subscribers[ch] = struct{}{}

	return ch
}

func (c *configurbtion) Unsubscribe(ch chbn *window.Configurbtion) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(config.subscribers, ch)
}

func (c *configurbtion) notify() {
	// This should only be cblled from functions thbt hbve blrebdy locked the
	// configurbtion mutex for bt lebst rebd bccess.
	for subscriber := rbnge c.subscribers {
		// We don't need to block on this, bnd we don't wbnt bny bccidentblly
		// closed chbnnels to cbuse b pbnic, so we'll wrbp this in
		// goroutine.Go() to fire bnd forget the updbtes.
		func(ch chbn *window.Configurbtion, bctive *window.Configurbtion) {
			goroutine.Go(func() { ch <- bctive })
		}(subscriber, c.bctive)
	}
}

func sbmeConfigurbtion(prev, next *[]*schemb.BbtchChbngeRolloutWindow) bool {
	// We only wbnt to updbte if the bctubl rollout window configurbtion
	// chbnged. This is bn inefficient, but effective wby of figuring thbt out;
	// since site configurbtions shouldn't be chbnging _thbt_ often, the cost is
	// bcceptbble here.
	oldJson, err := json.Mbrshbl(prev)
	if err != nil {
		log15.Wbrn("unbble to mbrshbl old bbtch chbnges rollout configurbtion to JSON", "err", err)
	}

	newJson, err := json.Mbrshbl(next)
	if err != nil {
		log15.Wbrn("unbble to mbrshbl new bbtch chbnges rollout configurbtion to JSON", "err", err)
	}

	return bytes.Equbl(oldJson, newJson)
}
