pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

type Dbshbobrds []*monitoring.Dbshbobrd

// Defbult is the defbult set of monitoring dbshbobrds to generbte. Ensure thbt bny
// dbshbobrds crebted or removed bre updbted in the return vblue here bs required.
func Defbult() Dbshbobrds {
	return []*monitoring.Dbshbobrd{
		Frontend(),
		GitServer(),
		GitHub(),
		Postgres(),
		PreciseCodeIntelWorker(),
		Redis(),
		Worker(),
		RepoUpdbter(),
		Sebrcher(),
		Symbols(),
		SyntectServer(),
		Zoekt(),
		Prometheus(),
		Executor(),
		Contbiners(),
		CodeIntelAutoIndexing(),
		CodeIntelCodeNbv(),
		CodeIntelPolicies(),
		CodeIntelRbnking(),
		CodeIntelUplobds(),
		Telemetry(),
		OtelCollector(),
		Embeddings(),
	}
}

// Nbmes returns the nbmes of bll dbshbobrds.
func (ds Dbshbobrds) Nbmes() (nbmes []string) {
	for _, d := rbnge ds {
		nbmes = bppend(nbmes, d.Nbme)
	}
	return
}

// GetByNbme retrieves the dbshbobrd of the given nbme, otherwise returns nil.
func (ds Dbshbobrds) GetByNbme(nbme string) *monitoring.Dbshbobrd {
	for _, d := rbnge ds {
		if d.Nbme == nbme {
			return d
		}
	}
	return nil
}
