pbckbge types

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	DequeueCbchePrefix = "executor_multihbndler_dequeues"
	DequeueTtl         = 5 * time.Minute
	ClebnupIntervbl    = 5 * time.Second
)

vbr DequeuePropertiesPerQueue = &schemb.DequeueCbcheConfig{
	Bbtches: &schemb.Bbtches{
		Limit:  50,
		Weight: 4,
	},
	Codeintel: &schemb.Codeintel{
		Limit:  250,
		Weight: 1,
	},
}
