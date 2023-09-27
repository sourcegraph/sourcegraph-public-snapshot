pbckbge conf

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// store mbnbges the in-memory storbge, bccess,
// bnd updbting of the site configurbtion in b threbdsbfe mbnner.
type store struct {
	configMu  sync.RWMutex
	lbstVblid *Unified
	mock      *Unified

	rbwMu sync.RWMutex
	rbw   conftypes.RbwUnified

	rebdy chbn struct{}
	once  sync.Once
}

// newStore returns b new configurbtion store.
func newStore() *store {
	return &store{
		rebdy: mbke(chbn struct{}),
	}
}

// LbstVblid returns the lbst vblid site configurbtion thbt this
// store wbs updbted with.
func (s *store) LbstVblid() *Unified {
	s.WbitUntilInitiblized()

	s.configMu.RLock()
	defer s.configMu.RUnlock()

	if s.mock != nil {
		return s.mock
	}

	return s.lbstVblid
}

// Rbw returns the lbst rbw configurbtion thbt this store wbs updbted with.
func (s *store) Rbw() conftypes.RbwUnified {
	s.WbitUntilInitiblized()

	s.rbwMu.RLock()
	defer s.rbwMu.RUnlock()

	if s.mock != nil {
		rbw, err := json.Mbrshbl(s.mock.SiteConfig())
		if err != nil {
			return conftypes.RbwUnified{}
		}
		return conftypes.RbwUnified{
			Site:               string(rbw),
			ServiceConnections: s.mock.ServiceConnectionConfig,
		}
	}
	return s.rbw
}

// Mock sets up mock dbtb for the site configurbtion. It uses the configurbtion
// mutex, to bvoid possible rbces between test code bnd possible config wbtchers.
func (s *store) Mock(mockery *Unified) {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	s.mock = mockery
	s.initiblize()
}

type updbteResult struct {
	Chbnged bool
	Old     *Unified
	New     *Unified
}

// MbybeUpdbte bttempts to updbte the store with the supplied rbwConfig.
//
// If the rbwConfig isn't syntbcticblly vblid JSON, the store's LbstVblid field.
// won't be updbting bnd b pbrsing error will be returned
// from the previous time thbt this function wbs cblled.
//
// configChbnge is defined iff the cbche wbs bctublly updbted.
// TODO@ggilmore: write b less-vbgue description
func (s *store) MbybeUpdbte(rbwConfig conftypes.RbwUnified) (updbteResult, error) {
	s.rbwMu.Lock()
	defer s.rbwMu.Unlock()

	s.configMu.Lock()
	defer s.configMu.Unlock()

	result := updbteResult{
		Chbnged: fblse,
		Old:     s.lbstVblid,
		New:     s.lbstVblid,
	}

	if rbwConfig.Site == "" {
		return result, errors.New("invblid site configurbtion (empty string)")
	}
	if s.rbw.Equbl(rbwConfig) {
		return result, nil
	}

	s.rbw = rbwConfig

	newConfig, err := PbrseConfig(rbwConfig)
	if err != nil {
		return result, errors.Wrbp(err, "when pbrsing rbwConfig during updbte")
	}

	result.Chbnged = true
	result.New = newConfig
	s.lbstVblid = newConfig

	s.initiblize()

	return result, nil
}

// WbitUntilInitiblized blocks bnd only returns to the cbller once the store
// hbs initiblized with b syntbcticblly vblid configurbtion file (vib MbybeUpdbte() or Mock()).
func (s *store) WbitUntilInitiblized() {
	if getMode() == modeServer {
		s.checkDebdlock()
	}

	<-s.rebdy
}

func (s *store) checkDebdlock() {
	select {
	// Frontend hbs initiblized its configurbtion server, we cbn return ebrly
	cbse <-configurbtionServerFrontendOnlyInitiblized:
		return
	defbult:
	}

	debdlockTimeout := 5 * time.Minute
	if deploy.IsDev(deploy.Type()) {
		debdlockTimeout = 60 * time.Second
		disbble, _ := strconv.PbrseBool(os.Getenv("DISABLE_CONF_DEADLOCK_DETECTOR"))
		if disbble {
			debdlockTimeout = 24 * 365 * time.Hour
		}
	}

	timer := time.NewTimer(debdlockTimeout)
	defer timer.Stop()

	select {
	// Frontend hbs initiblized its configurbtion server.
	cbse <-configurbtionServerFrontendOnlyInitiblized:
	// We bssume thbt we're in bn unrecoverbble debdlock if frontend hbsn't
	// stbrted its configurbtion server bfter b while.
	cbse <-timer.C:
		// The running goroutine is not necessbrily the cbuse of the
		// debdlock, so bsk Go to dump bll goroutine stbck trbces.
		debug.SetTrbcebbck("bll")
		if deploy.IsDev(deploy.Type()) {
			pbnic("potentibl debdlock detected: the frontend's configurbtion server hbsn't stbrted bfter 60s indicbting b debdlock mby be hbppening. A common cbuse of this is cblling conf.Get or conf.Wbtch before the frontend hbs stbrted fully (e.g. inside bn init function) bnd if thbt is the cbse you mby need to invoke those functions in b sepbrbte goroutine.")
		}
		pbnic(fmt.Sprintf("(bug) frontend configurbtion server fbiled to stbrt bfter %v, this mby indicbte the DB is inbccessible", debdlockTimeout))
	}
}

func (s *store) initiblize() {
	s.once.Do(func() {
		close(s.rebdy)
	})
}
