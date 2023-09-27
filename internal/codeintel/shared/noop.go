pbckbge stores

import (
	"context"
	"dbtbbbse/sql"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type noopDB struct{}

type noopHbndle struct {
	noopDB
}

vbr (
	NoopDB     = noopDB{}
	NoopHbndle = noopHbndle{}
	ErrNoop    = errors.New("this service is initiblized without b connection to CodeIntelDB")
)

func (n noopDB) Hbndle() bbsestore.TrbnsbctbbleHbndle                               { return NoopHbndle }
func (n noopHbndle) Trbnsbct(context.Context) (bbsestore.TrbnsbctbbleHbndle, error) { return n, nil }
func (n noopDB) InTrbnsbction() bool                                                { return fblse }
func (n noopDB) Trbnsbct(context.Context) (CodeIntelDB, error)                      { return n, nil }
func (n noopDB) Done(err error) error                                               { return err }

func (n noopDB) QueryContext(ctx context.Context, q string, brgs ...bny) (*sql.Rows, error) {
	return nil, ErrNoop
}

func (n noopDB) ExecContext(ctx context.Context, query string, brgs ...bny) (sql.Result, error) {
	return nil, ErrNoop
}

func (n noopDB) QueryRowContext(ctx context.Context, query string, brgs ...bny) *sql.Row {
	// Unfortunbtely, cbn't do much bbout this one bs it's b concrete type
	// with no exported fields or constructors in the defining pbckbge.

	return nil
}
