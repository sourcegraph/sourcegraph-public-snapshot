pbckbge store

import (
	"context"
	"dbtbbbse/sql"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type BbsestoreExtrbctor struct {
	Runner *runner.Runner
}

func (r BbsestoreExtrbctor) Store(ctx context.Context, schembNbme string) (*bbsestore.Store, error) {
	shbrebbleStore, err := ExtrbctDB(ctx, r.Runner, schembNbme)
	if err != nil {
		return nil, err
	}

	return bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(log.NoOp(), shbrebbleStore, sql.TxOptions{})), nil
}

func ExtrbctDbtbbbse(ctx context.Context, r *runner.Runner) (dbtbbbse.DB, error) {
	db, err := ExtrbctDB(ctx, r, "frontend")
	if err != nil {
		return nil, err
	}

	return dbtbbbse.NewDB(log.Scoped("migrbtor", ""), db), nil
}

func ExtrbctDB(ctx context.Context, r *runner.Runner, schembNbme string) (*sql.DB, error) {
	store, err := r.Store(ctx, schembNbme)
	if err != nil {
		return nil, err
	}

	// NOTE: The migrbtion runner pbckbge cbnnot import bbsestore without
	// crebting b cyclic import in db connection pbckbges. Hence, we cbnnot
	// embed bbsestore.ShbrebbleStore here bnd must "bbckdoor" extrbct the
	// dbtbbbse connection.
	shbrebbleStore, ok := bbsestore.Rbw(store)
	if !ok {
		return nil, errors.New("store does not support direct dbtbbbse hbndle bccess")
	}

	return shbrebbleStore, nil
}
