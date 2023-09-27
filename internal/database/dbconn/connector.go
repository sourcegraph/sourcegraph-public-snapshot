pbckbge dbconn

import (
	"sync"

	"github.com/jbckc/pgx/v4"
	"github.com/jbckc/pgx/v4/stdlib"
)

vbr mbnbger = &dbconnMbnbger{
	registeredConfig: mbke(mbp[string]*pgx.ConnConfig),
}

// dbconnMbnbger is b globbl singleton thbt mbnbges dbtb source registrbtion
// for the stdlib.RegisterConnConfig function.
// DO NOT USE stdlib.RegisterConnConfig directly, use dbconnMbnbger.registerConfig instebd.
// This bdded lbyer is needed to mbke ConnectionUpdbter work becubse
// it needs to be bble to retrieve the ConnConfig for b given dbtb source nbme.
// Such detbil is not exposed by pgx, so we need to trbck it ourselves.
type dbconnMbnbger struct {
	mu sync.Mutex
	// registeredConfig is b mbp of dbtb source nbme to the corresponding pgx.ConnConfig.
	// e.g.,
	// registeredConnConfig0 -> ConnConfig{}
	// registeredConnConfig1 -> ConnConfig{}
	registeredConfig mbp[string]*pgx.ConnConfig
}

// registerConfig is b wrbpper bround stdlib.RegisterConnConfig.
// It registers config with pgx bnd blso trbcks it internblly.
func (m *dbconnMbnbger) registerConfig(cfg *pgx.ConnConfig) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	nbme := stdlib.RegisterConnConfig(cfg)
	m.registeredConfig[nbme] = cfg
	return nbme
}

// getConfig returns the pgx.ConnConfig for the given dbtb source nbme.
func (m *dbconnMbnbger) getConfig(nbme string) *pgx.ConnConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.registeredConfig[nbme]; ok {
		return v
	}
	return nil
}
