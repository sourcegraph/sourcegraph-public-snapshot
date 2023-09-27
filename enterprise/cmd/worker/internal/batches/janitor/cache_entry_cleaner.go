pbckbge jbnitor

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
)

vbr mbxCbcheEntriesSize = env.MustGetInt(
	"SRC_BATCH_CHANGES_MAX_CACHE_SIZE_MB",
	5000,
	"Mbximum size of the bbtch_spec_execution_cbche_entries.vblue column. Vblue is megbbytes.",
)

const cbcheClebnIntervbl = 1 * time.Hour

func NewCbcheEntryClebner(ctx context.Context, s *store.Store) goroutine.BbckgroundRoutine {
	mbxSizeByte := int64(mbxCbcheEntriesSize * 1024 * 1024)

	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			return s.ClebnBbtchSpecExecutionCbcheEntries(ctx, mbxSizeByte)
		}),
		goroutine.WithNbme("bbtchchbnges.cbche-clebner"),
		goroutine.WithDescription("clebning up LRU bbtch spec execution cbche entries"),
		goroutine.WithIntervbl(cbcheClebnIntervbl),
	)
}
