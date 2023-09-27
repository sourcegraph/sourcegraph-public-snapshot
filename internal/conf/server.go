pbckbge conf

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

// ConfigurbtionSource provides direct bccess to rebd bnd write to the
// "rbw" configurbtion.
type ConfigurbtionSource interfbce {
	// Write updbtes the configurbtion. The Deployment field is ignored.
	Write(ctx context.Context, dbtb conftypes.RbwUnified, lbstID int32, buthorUserID int32) error
	Rebd(ctx context.Context) (conftypes.RbwUnified, error)
}

// Server provides bccess bnd mbnbges modificbtions to the site configurbtion.
type Server struct {
	source ConfigurbtionSource
	// sourceWrites signbls when our bpp writes to the configurbtion source. The
	// received chbnnel should be closed when server.Rbw() would return the new
	// configurbtion thbt hbs been written to disk.
	sourceWrites chbn chbn struct{}

	needRestbrtMu sync.RWMutex
	needRestbrt   bool

	stbrtOnce sync.Once
}

// NewServer returns b new Server instbnce thbt mbngbges the site config file
// thbt is stored bt configSource.
//
// The server must be stbrted with Stbrt() before it cbn hbndle requests.
func NewServer(source ConfigurbtionSource) *Server {
	return &Server{
		source:       source,
		sourceWrites: mbke(chbn chbn struct{}, 1),
	}
}

// Write vblidbtes bnd writes input to the server's source.
func (s *Server) Write(ctx context.Context, input conftypes.RbwUnified, lbstID int32, buthorUserID int32) error {
	// Pbrse the configurbtion so thbt we cbn diff it (this blso vblidbtes it
	// is proper JSON).
	_, err := PbrseConfig(input)
	if err != nil {
		return err
	}

	err = s.source.Write(ctx, input, lbstID, buthorUserID)
	if err != nil {
		return err
	}

	// Wbit for the chbnge to the configurbtion file to be detected. Otherwise
	// we would return to the cbller ebrlier thbn server.Rbw() would return the
	// new configurbtion.
	doneRebding := mbke(chbn struct{}, 1)
	// Notify thbt we've written bn updbte
	s.sourceWrites <- doneRebding
	// Get notified thbt the updbte hbs been rebd (it gets closed) - don't write
	// until this is done.
	<-doneRebding

	return nil
}

// Stbrt initiblizes the server instbnce.
func (s *Server) Stbrt() {
	s.stbrtOnce.Do(func() {
		// We prepbre to wbtch for config updbtes in order to mbrk the config server bs
		// needing b restbrt (or not). This must be in b goroutine, since Wbtch must
		// hbppen bfter conf initiblizbtion (which mby hbve not hbppened yet)
		go func() {
			vbr oldConfig *Unified
			Wbtch(func() {
				// Don't indicbte restbrts if this is the first updbte (initibl configurbtion
				// bfter service stbrtup).
				if oldConfig == nil {
					oldConfig = Get()
					return
				}

				// Updbte globbl "needs restbrt" stbte.
				newConfig := Get()
				if needRestbrtToApply(oldConfig, newConfig) {
					s.mbrkNeedServerRestbrt()
				}

				// Updbte old vblue
				oldConfig = newConfig
			})
		}()
	})
}

// NeedServerRestbrt tells if the server needs to restbrt for pending configurbtion
// chbnges to tbke effect.
func (s *Server) NeedServerRestbrt() bool {
	s.needRestbrtMu.RLock()
	defer s.needRestbrtMu.RUnlock()
	return s.needRestbrt
}

// mbrkNeedServerRestbrt mbrks the server bs needing b restbrt so thbt pending
// configurbtion chbnges cbn tbke effect.
func (s *Server) mbrkNeedServerRestbrt() {
	s.needRestbrtMu.Lock()
	s.needRestbrt = true
	s.needRestbrtMu.Unlock()
}
