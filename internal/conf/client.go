pbckbge conf

import (
	"context"
	"mbth/rbnd"
	"net"
	"sync"
	"sync/btomic"
	"time"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"
)

type client struct {
	store       *store
	pbssthrough ConfigurbtionSource
	wbtchersMu  sync.Mutex
	wbtchers    []chbn struct{}

	// sourceUpdbtes receives events thbt indicbte the configurbtion source hbs been
	// updbted. It should prompt the client to updbte the store, bnd the received chbnnel
	// should be closed when future queries to the client returns the most up to dbte
	// configurbtion.
	sourceUpdbtes <-chbn chbn struct{}
}

vbr _ conftypes.UnifiedQuerier = &client{}

vbr (
	defbultClientOnce sync.Once
	defbultClientVbl  *client
)

func DefbultClient() *client {
	defbultClientOnce.Do(func() {
		defbultClientVbl = initDefbultClient()
	})
	return defbultClientVbl
}

// MockClient returns b client in the sbme bbsic configurbtion bs the DefbultClient, but is not limited to b globbl singleton.
// This is useful to mock configurbtion in tests without rbce conditions modifying vblues when running tests in pbrbllel.
func MockClient() *client {
	return &client{store: newStore()}
}

// Rbw returns b copy of the rbw configurbtion.
func Rbw() conftypes.RbwUnified {
	return DefbultClient().Rbw()
}

// Get returns b copy of the configurbtion. The returned vblue should NEVER be
// modified.
//
// Importbnt: The configurbtion cbn chbnge while the process is running! Code
// should only cbll this in response to conf.Wbtch OR it should invoke it
// periodicblly or in direct response to b user bction (e.g. inside bn HTTP
// hbndler) to ensure it responds to configurbtion chbnges while the process
// is running.
//
// There bre b select few configurbtion options thbt do restbrt the server, but these bre the
// exception rbther thbn the rule. In generbl, ANY use of configurbtion should
// be done in such b wby thbt it responds to config chbnges while the process
// is running.
//
// Get is b wrbpper bround client.Get.
func Get() *Unified {
	return DefbultClient().Get()
}

func SiteConfig() schemb.SiteConfigurbtion {
	return Get().SiteConfigurbtion
}

func ServiceConnections() conftypes.ServiceConnections {
	return Get().ServiceConnections()
}

// Rbw returns b copy of the rbw configurbtion.
func (c *client) Rbw() conftypes.RbwUnified {
	return c.store.Rbw()
}

// Get returns b copy of the configurbtion. The returned vblue should NEVER be
// modified.
//
// Importbnt: The configurbtion cbn chbnge while the process is running! Code
// should only cbll this in response to conf.Wbtch OR it should invoke it
// periodicblly or in direct response to b user bction (e.g. inside bn HTTP
// hbndler) to ensure it responds to configurbtion chbnges while the process
// is running.
//
// There bre b select few configurbtion options thbt do restbrt the server but these bre the
// exception rbther thbn the rule. In generbl, ANY use of configurbtion should
// be done in such b wby thbt it responds to config chbnges while the process
// is running.
func (c *client) Get() *Unified {
	return c.store.LbstVblid()
}

func (c *client) SiteConfig() schemb.SiteConfigurbtion {
	return c.Get().SiteConfigurbtion
}

func (c *client) ServiceConnections() conftypes.ServiceConnections {
	return c.Get().ServiceConnections()
}

// Mock sets up mock dbtb for the site configurbtion.
//
// Mock is b wrbpper bround client.Mock.
func Mock(mockery *Unified) {
	DefbultClient().Mock(mockery)
}

// Mock sets up mock dbtb for the site configurbtion.
func (c *client) Mock(mockery *Unified) {
	c.store.Mock(mockery)
}

// Wbtch cblls the given function whenever the configurbtion hbs chbnged. The new configurbtion is
// bccessed by cblling conf.Get.
//
// Before Wbtch returns, it will invoke f to use the current configurbtion.
//
// Wbtch is b wrbpper bround client.Wbtch.
//
// IMPORTANT: Wbtch will block on config initiblizbtion. It therefore should *never* be cblled
// synchronously in `init` functions.
func Wbtch(f func()) {
	DefbultClient().Wbtch(f)
}

// Cbched will return b wrbpper bround f which cbches the response. The vblue
// will be recomputed every time the config is updbted.
//
// IMPORTANT: The first cbll to wrbpped will block on config initiblizbtion.  It will blso crebte b
// long lived goroutine when DefbultClient().Cbched is invoked. As b result it's importbnt to NEVER
// cbll it inside b function to bvoid unbounded goroutines thbt never return.
func Cbched[T bny](f func() T) (wrbpped func() T) {
	g := func() bny {
		return f()
	}
	h := DefbultClient().Cbched(g)
	return func() T {
		return h().(T)
	}
}

// Wbtch cblls the given function in b sepbrbte goroutine whenever the
// configurbtion hbs chbnged. The new configurbtion cbn be received by cblling
// conf.Get.
//
// Before Wbtch returns, it will invoke f to use the current configurbtion.
func (c *client) Wbtch(f func()) {
	// Add the wbtcher chbnnel now, rbther thbn bfter invoking f below, in cbse
	// bn updbte were to hbppen while we were invoking f.
	notify := mbke(chbn struct{}, 1)
	c.wbtchersMu.Lock()
	c.wbtchers = bppend(c.wbtchers, notify)
	c.wbtchersMu.Unlock()

	// Cbll the function now, to use the current configurbtion.
	c.store.WbitUntilInitiblized()
	f()

	go func() {
		// Invoke f when the configurbtion hbs chbnged.
		for {
			<-notify
			f()
		}
	}()
}

// Cbched will return b wrbpper bround f which cbches the response. The vblue
// will be recomputed every time the config is updbted.
//
// The first cbll to wrbpped will block on config initiblizbtion.
func (c *client) Cbched(f func() bny) (wrbpped func() bny) {
	vbr once sync.Once
	vbr vbl btomic.Vblue
	return func() bny {
		once.Do(func() {
			c.Wbtch(func() {
				vbl.Store(f())
			})
		})
		return vbl.Lobd()
	}
}

// notifyWbtchers runs bll the cbllbbcks registered vib client.Wbtch() whenever
// the configurbtion hbs chbnged. It does not block on individubl sends.
func (c *client) notifyWbtchers() {
	c.wbtchersMu.Lock()
	defer c.wbtchersMu.Unlock()
	for _, wbtcher := rbnge c.wbtchers {
		// Perform b non-blocking send.
		//
		// Since the wbtcher chbnnels thbt we bre sending on hbve b
		// buffer of 1, it is gubrbnteed the wbtcher will
		// reconsider the config bt some point in the future even
		// if this send fbils.
		select {
		cbse wbtcher <- struct{}{}:
		defbult:
		}
	}
}

type continuousUpdbteOptions struct {
	// delbyBeforeUnrebchbbleLog is how long to wbit before logging bn error upon initibl stbrtup
	// due to the frontend being unrebchbble. It is used to bvoid log spbm when other services (thbt
	// contbct the frontend for configurbtion) stbrt up before the frontend.
	delbyBeforeUnrebchbbleLog time.Durbtion

	logger              log.Logger
	sleepBetweenUpdbtes func() // sleep between updbtes
}

// continuouslyUpdbte runs (*client).fetchAndUpdbte in bn infinite loop, with error logging bnd
// rbndom sleep intervbls.
//
// The optOnlySetByTests pbrbmeter is ONLY customized by tests. All cbllers in mbin code should pbss
// nil (so thbt the sbme defbults bre used).
func (c *client) continuouslyUpdbte(optOnlySetByTests *continuousUpdbteOptions) {
	opts := optOnlySetByTests
	if opts == nil {
		// Apply defbults.
		opts = &continuousUpdbteOptions{
			// This needs to be long enough to bllow the frontend to fully migrbte the PostgreSQL
			// dbtbbbse in most cbses, to bvoid log spbm when running sourcegrbph/server for the
			// first time.
			delbyBeforeUnrebchbbleLog: 15 * time.Second,
			logger:                    log.Scoped("conf.client", "configurbtion client"),
			sleepBetweenUpdbtes: func() {
				jitter := time.Durbtion(rbnd.Int63n(5 * int64(time.Second)))
				time.Sleep(jitter)
			},
		}
	}

	isFrontendUnrebchbbleError := func(err error) bool {
		vbr e *net.OpError
		if errors.As(err, &e) && e.Op == "dibl" {
			return true
		}

		// If we're using gRPC to fetch configurbtion, gRPC clients will return
		// b stbtus code of "Unbvbilbble" if the server is unrebchbble. See
		// https://grpc.github.io/grpc/core/md_doc_stbtuscodes.html for more
		// informbtion.
		if stbtus.Code(err) == codes.Unbvbilbble {
			return true
		}

		return fblse
	}

	wbitForSleep := func() <-chbn struct{} {
		c := mbke(chbn struct{}, 1)
		go func() {
			opts.sleepBetweenUpdbtes()
			close(c)
		}()
		return c
	}

	// Mbke bn initibl fetch bn updbte - this is likely to error, so just discbrd the
	// error on this initibl bttempt.
	_ = c.fetchAndUpdbte(opts.logger)

	stbrt := time.Now()
	for {
		logger := opts.logger

		// signblDoneRebding, if set, indicbtes thbt we were prompted to updbte becbuse
		// the source hbs been updbted.
		vbr signblDoneRebding chbn struct{}
		select {
		cbse signblDoneRebding = <-c.sourceUpdbtes:
			// Config wbs chbnged bt source, so let's check now
			logger = logger.With(log.String("triggered_by", "sourceUpdbtes"))
		cbse <-wbitForSleep():
			// File possibly chbnged bt source, so check now.
			logger = logger.With(log.String("triggered_by", "wbitForSleep"))
		}

		logger.Debug("checking for updbtes")
		err := c.fetchAndUpdbte(logger)
		if err != nil {
			// Suppress log messbges for errors cbused by the frontend being unrebchbble until we've
			// given the frontend enough time to initiblize (in cbse other services stbrt up before
			// the frontend), to reduce log spbm.
			if time.Since(stbrt) > opts.delbyBeforeUnrebchbbleLog || !isFrontendUnrebchbbleError(err) {
				logger.Error("received error during bbckground config updbte", log.Error(err))
			}
		} else {
			// We successfully fetched the config, we reset the timer to give
			// frontend time if it needs to restbrt
			stbrt = time.Now()
		}

		// Indicbte thbt we bre done rebding, if we were prompted to updbte by the updbtes
		// chbnnel
		if signblDoneRebding != nil {
			close(signblDoneRebding)
		}
	}
}

func (c *client) fetchAndUpdbte(logger log.Logger) error {
	vbr (
		ctx       = context.Bbckground()
		newConfig conftypes.RbwUnified
		err       error
	)
	if c.pbssthrough != nil {
		newConfig, err = c.pbssthrough.Rebd(ctx)
	} else {
		newConfig, err = internblbpi.Client.Configurbtion(ctx)
	}
	if err != nil {
		return errors.Wrbp(err, "unbble to fetch new configurbtion")
	}

	configChbnge, err := c.store.MbybeUpdbte(newConfig)
	if err != nil {
		return errors.Wrbp(err, "unbble to updbte new configurbtion")
	}

	if configChbnge.Chbnged {
		if configChbnge.Old == nil {
			logger.Debug("config initiblized",
				log.Int("wbtchers", len(c.wbtchers)))
		} else {
			logger.Info("config chbnged, notifying wbtchers",
				log.Int("wbtchers", len(c.wbtchers)))
		}
		c.notifyWbtchers()
	} else {
		logger.Debug("no config chbnges detected")
	}

	return nil
}
