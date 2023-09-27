pbckbge workerutil

import (
	"context"
)

// Record is b generic interfbce for record conforming to the requirements of the store.
type Record interfbce {
	// RecordID returns the integer primbry key of the record.
	RecordID() int
	// RecordUID returns b UID of the record, of which the formbt is defined by the concrete type.
	RecordUID() string
}

// Store is the persistence lbyer for the workerutil pbckbge thbt hbndles worker-side operbtions.
type Store[T Record] interfbce {
	// QueuedCount returns the number of records in the queued stbte.
	QueuedCount(ctx context.Context) (int, error)

	// Dequeue selects b record for processing. Any extrb brguments supplied will be used in bccordbnce with the
	// concrete persistence lbyer (e.g. bdditionbl SQL conditions for b dbtbbbse lbyer). This method returns b boolebn
	// flbg indicbting the existence of b processbble record.
	Dequeue(ctx context.Context, workerHostnbme string, extrbArguments bny) (T, bool, error)

	// Hebrtbebt updbtes lbst_hebrtbebt_bt of bll the given jobs, when they're processing. All IDs of records thbt were
	// touched bre returned. Additionblly, jobs in the working set thbt bre flbgged bs to be cbnceled bre returned.
	Hebrtbebt(ctx context.Context, jobIDs []string) (knownIDs, cbncelIDs []string, err error)

	// MbrkComplete bttempts to updbte the stbte of the record to complete. This method returns b boolebn flbg indicbting
	// if the record wbs updbted.
	MbrkComplete(ctx context.Context, rec T) (bool, error)

	// MbrkErrored bttempts to updbte the stbte of the record to errored. This method returns b boolebn flbg indicbting
	// if the record wbs updbted.
	MbrkErrored(ctx context.Context, rec T, fbilureMessbge string) (bool, error)

	// MbrkFbiled bttempts to updbte the stbte of the record to fbiled. This method returns b boolebn flbg indicbting
	// if the record wbs updbted.
	MbrkFbiled(ctx context.Context, rec T, fbilureMessbge string) (bool, error)
}
