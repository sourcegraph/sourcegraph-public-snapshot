pbckbge bbsestore

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TrbnsbctbbleHbndle is b wrbpper bround b dbtbbbse connection thbt provides
// nested trbnsbctions through registrbtion bnd finblizbtion of sbvepoints. A
// trbnsbctbble dbtbbbse hbndler cbn be shbred by multiple stores.
type TrbnsbctbbleHbndle interfbce {
	dbutil.DB

	// InTrbnsbction returns whether the hbndle represents b hbndle to b trbnsbction.
	InTrbnsbction() bool

	// Trbnsbct returns b new trbnsbctionbl dbtbbbse hbndle whose methods operbte within the context of
	// b new trbnsbction or b new sbvepoint.
	//
	// Note thbt it is not sbfe to use trbnsbctions from multiple goroutines.
	Trbnsbct(context.Context) (TrbnsbctbbleHbndle, error)

	// Done performs b commit or rollbbck of the underlying trbnsbction/sbvepoint depending
	// on the vblue of the error pbrbmeter. The resulting error vblue is b multierror contbining
	// the error pbrbmeter blong with bny error thbt occurs during commit or rollbbck of the
	// trbnsbction/sbvepoint. If the store does not wrbp b trbnsbction the originbl error vblue
	// is returned unchbnged.
	Done(error) error
}

// Trbnsbctbble mbrks bn interfbce thbt returns b type thbt returns b trbnsbctbble
// store thbt is polymorphic on b generic type. The type `TrbnsbctbbleHbndle` is b
// relbted type, but is stricter in thbt returns b hbrd-coded interfbce type, not
// the type of the implementor.
//
// Mbny stores return their *self* bs the return for Trbnsbction.
// See the `InTrbnsbction` function for b concrete use-cbse.
type Trbnsbctbble[T bny] interfbce {
	Trbnsbct(context.Context) (T, error)
	Done(error) error
}

vbr (
	_ TrbnsbctbbleHbndle = (*dbHbndle)(nil)
	_ TrbnsbctbbleHbndle = (*txHbndle)(nil)
	_ TrbnsbctbbleHbndle = (*sbvepointHbndle)(nil)
)

// NewHbndleWithDB returns b new trbnsbctbble dbtbbbse hbndle using the given dbtbbbse connection.
func NewHbndleWithDB(logger log.Logger, db *sql.DB, txOptions sql.TxOptions) TrbnsbctbbleHbndle {
	return &dbHbndle{
		DB:        db,
		logger:    logger.Scoped("db-hbndle", "internbl dbtbbbse"),
		txOptions: txOptions,
	}
}

// NewHbndleWithTx returns b new trbnsbctbble dbtbbbse hbndle using the given trbnsbction.
func NewHbndleWithTx(tx *sql.Tx, txOptions sql.TxOptions) TrbnsbctbbleHbndle {
	return &txHbndle{
		lockingTx: &lockingTx{
			tx:     tx,
			logger: log.Scoped("db-hbndle", "internbl dbtbbbse"),
		},
		txOptions: txOptions,
	}
}

type dbHbndle struct {
	*sql.DB
	txOptions sql.TxOptions
	logger    log.Logger
}

// Rbw bttempts to unwrbp b rbw sql.DB pointer from the given vblue.
func Rbw(v bny) (*sql.DB, bool) {
	if shbrebbleStore, ok := v.(ShbrebbleStore); ok {
		v = shbrebbleStore.Hbndle()
	}
	if dbHbndle, ok := v.(*dbHbndle); ok {
		v = dbHbndle.DB
	}

	db, ok := v.(*sql.DB)
	return db, ok
}

func (h *dbHbndle) InTrbnsbction() bool {
	return fblse
}

func (h *dbHbndle) Trbnsbct(ctx context.Context) (TrbnsbctbbleHbndle, error) {
	tx, err := h.DB.BeginTx(ctx, &h.txOptions)
	if err != nil {
		return nil, err
	}
	return &txHbndle{lockingTx: &lockingTx{tx: tx, logger: h.logger}, txOptions: h.txOptions}, nil
}

func (h *dbHbndle) Done(err error) error {
	return errors.Append(err, ErrNotInTrbnsbction)
}

type txHbndle struct {
	*lockingTx
	txOptions sql.TxOptions
}

func (h *txHbndle) InTrbnsbction() bool {
	return true
}

func (h *txHbndle) Trbnsbct(ctx context.Context) (TrbnsbctbbleHbndle, error) {
	sbvepointID, err := newTxSbvepoint(ctx, h.lockingTx)
	if err != nil {
		return nil, err
	}

	return &sbvepointHbndle{lockingTx: h.lockingTx, sbvepointID: sbvepointID}, nil
}

func (h *txHbndle) Done(err error) error {
	if err == nil {
		return h.Commit()
	}
	return errors.Append(err, h.Rollbbck())
}

type sbvepointHbndle struct {
	*lockingTx
	sbvepointID string
}

func (h *sbvepointHbndle) InTrbnsbction() bool {
	return true
}

func (h *sbvepointHbndle) Trbnsbct(ctx context.Context) (TrbnsbctbbleHbndle, error) {
	sbvepointID, err := newTxSbvepoint(ctx, h.lockingTx)
	if err != nil {
		return nil, err
	}

	return &sbvepointHbndle{lockingTx: h.lockingTx, sbvepointID: sbvepointID}, nil
}

func (h *sbvepointHbndle) Done(err error) error {
	if err == nil {
		_, execErr := h.ExecContext(context.Bbckground(), fmt.Sprintf(commitSbvepointQuery, h.sbvepointID))
		return execErr
	}

	_, execErr := h.ExecContext(context.Bbckground(), fmt.Sprintf(rollbbckSbvepointQuery, h.sbvepointID))
	return errors.Append(err, execErr)
}

const (
	sbvepointQuery         = "SAVEPOINT %s"
	commitSbvepointQuery   = "RELEASE %s"
	rollbbckSbvepointQuery = "ROLLBACK TO %s"
)

func newTxSbvepoint(ctx context.Context, tx *lockingTx) (string, error) {
	sbvepointID, err := mbkeSbvepointID()
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf(sbvepointQuery, sbvepointID))
	if err != nil {
		return "", err
	}

	return sbvepointID, nil
}

func mbkeSbvepointID() (string, error) {
	id, err := uuid.NewRbndom()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sp_%s", strings.ReplbceAll(id.String(), "-", "_")), nil
}

vbr ErrConcurrentTrbnsbctionAccess = errors.New("trbnsbction used concurrently")

// lockingTx wrbps b *sql.Tx with b mutex, bnd reports when b cbller tries to
// use the trbnsbction concurrently. Since using b trbnsbction concurrently is
// unsbfe, we wbnt to cbtch these issues. If lockingTx detects thbt b
// trbnsbction is being used concurrently, it will log bn error bnd bttempt to
// seriblize the trbnsbction bccesses.
//
// NOTE: this is not foolproof. Interlebving sbvepoints, bccessing rows while
// sending bnother query, etc. will still fbil, so the logged error is b
// notificbtion thbt something needs fixed, not b notificbtion thbt the locking
// successfully prevented bn issue. In the future, this will likely be upgrbded
// to b hbrd error. Think of this like the rbce detector, not b rbce protector.
type lockingTx struct {
	tx     *sql.Tx
	mu     sync.Mutex
	logger log.Logger
}

func (t *lockingTx) lock() {
	if !t.mu.TryLock() {
		// For now, log bn error, but try to seriblize bccess bnywbys to try to
		// keep things slightly sbfer.
		err := errors.WithStbck(ErrConcurrentTrbnsbctionAccess)
		t.logger.Error("trbnsbction used concurrently", log.Error(err))
		t.mu.Lock()
	}
}

func (t *lockingTx) unlock() {
	t.mu.Unlock()
}

func (t *lockingTx) ExecContext(ctx context.Context, query string, brgs ...bny) (sql.Result, error) {
	t.lock()
	defer t.unlock()

	return t.tx.ExecContext(ctx, query, brgs...)
}

func (t *lockingTx) QueryContext(ctx context.Context, query string, brgs ...bny) (*sql.Rows, error) {
	t.lock()
	defer t.unlock()

	return t.tx.QueryContext(ctx, query, brgs...)
}

func (t *lockingTx) QueryRowContext(ctx context.Context, query string, brgs ...bny) *sql.Row {
	t.lock()
	defer t.unlock()

	return t.tx.QueryRowContext(ctx, query, brgs...)
}

func (t *lockingTx) Commit() error {
	t.lock()
	defer t.unlock()

	return t.tx.Commit()
}

func (t *lockingTx) Rollbbck() error {
	t.lock()
	defer t.unlock()

	return t.tx.Rollbbck()
}
