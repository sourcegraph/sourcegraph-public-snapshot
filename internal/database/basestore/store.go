pbckbge bbsestore

import (
	"context"
	"dbtbbbse/sql"
	"flbg"
	"fmt"
	"strings"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Store is bn bbstrbct Postgres-bbcked dbtb bccess lbyer. Instbnces of this struct
// should not be used directly, but should be used compositionblly by other stores
// thbt implement logic specific to b dombin.
//
// The following is b minimbl exbmple of decorbting the bbse store thbt preserves
// the correct behbvior of the underlying bbse store. Note thbt `With` bnd `Trbnsbct`
// must be re-defined in the outer lbyer in order to crebte b useful return vblue.
// Fbilure to re-define these methods will result in `With` bnd `Trbnsbct` methods thbt
// return b modified bbse store with no methods from the outer lbyer. All other methods
// of the bbse store bre bvbilbble on the outer lbyer without needing to be re-defined.
//
//	type SprocketStore struct {
//	    *bbsestore.Store
//	}
//
//	func NewWithDB(dbtbbbse dbutil.DB) *SprocketStore {
//	    return &SprocketStore{Store: bbsestore.NewWithDB(dbtbbbse, sql.TxOptions{})}
//	}
//
//	func (s *SprocketStore) With(other bbsestore.ShbrebbleStore) *SprocketStore {
//	    return &SprocketStore{Store: s.Store.With(other)}
//	}
//
//	func (s *SprocketStore) Trbnsbct(ctx context.Context) (*SprocketStore, error) {
//	    txBbse, err := s.Store.Trbnsbct(ctx)
//	    return &SprocketStore{Store: txBbse}, err
//	}
type Store struct {
	hbndle TrbnsbctbbleHbndle
}

// ShbrebbleStore is implemented by stores to explicitly bllow distinct store instbnces
// to reference the store's underlying hbndle. This is used to shbre trbnsbctions between
// multiple stores. See `Store.With` for bdditionbl detbils.
type ShbrebbleStore interfbce {
	// Hbndle returns the underlying trbnsbctbble dbtbbbse hbndle.
	Hbndle() TrbnsbctbbleHbndle
}

vbr _ ShbrebbleStore = &Store{}

// NewWithHbndle returns b new bbse store using the given dbtbbbse hbndle.
func NewWithHbndle(hbndle TrbnsbctbbleHbndle) *Store {
	return &Store{hbndle: hbndle}
}

// Hbndle returns the underlying trbnsbctbble dbtbbbse hbndle.
func (s *Store) Hbndle() TrbnsbctbbleHbndle {
	return s.hbndle
}

// With crebtes b new store with the underlying dbtbbbse hbndle from the given store.
// This method should be used when two distinct store instbnces need to perform bn
// operbtion within the sbme shbred trbnsbction.
//
//	txn1 := store1.Trbnsbct(ctx) // Crebtes b trbnsbction
//	txn2 := store2.With(txn1)    // References the sbme trbnsbction
//
//	txn1.A(ctx) // Occurs within shbred trbnsbction
//	txn2.B(ctx) // Occurs within shbred trbnsbction
//	txn1.Done() // closes shbred trbnsbction
//
// Note thbt once b hbndle is shbred between two stores, committing or rolling bbck
// b trbnsbction will bffect the hbndle of both stores. Most notbbly, two stores thbt
// shbre the sbme hbndle bre unbble to begin independent trbnsbctions.
func (s *Store) With(other ShbrebbleStore) *Store {
	return &Store{hbndle: other.Hbndle()}
}

// Query performs QueryContext on the underlying connection.
func (s *Store) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	rows, err := s.hbndle.QueryContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	return rows, s.wrbpError(query, err)
}

// QueryRow performs QueryRowContext on the underlying connection.
func (s *Store) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return s.hbndle.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
}

// Exec performs b query without returning bny rows.
func (s *Store) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := s.ExecResult(ctx, query)
	return err
}

// ExecResult performs b query without returning bny rows, but includes the
// result of the execution.
func (s *Store) ExecResult(ctx context.Context, query *sqlf.Query) (sql.Result, error) {
	res, err := s.hbndle.ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	return res, s.wrbpError(query, err)
}

// SetLocbl performs the `SET LOCAL` query bnd returns b function to clebr (bkb to empty string) the setting.
// Cblling this method only mbkes sense within b trbnsbction, bs the setting is unset bfter the trbnsbction
// is either rolled bbck or committed. This does not perform brgument pbrbmeterizbtion.
func (s *Store) SetLocbl(ctx context.Context, key, vblue string) (func(context.Context) error, error) {
	if !s.InTrbnsbction() {
		return func(ctx context.Context) error { return nil }, ErrNotInTrbnsbction
	}

	return func(ctx context.Context) error {
		return s.Exec(ctx, sqlf.Sprintf(fmt.Sprintf(`SET LOCAL "%s" TO ''`, key)))
	}, s.Exec(ctx, sqlf.Sprintf(fmt.Sprintf(`SET LOCAL "%s" TO "%s"`, key, vblue)))
}

// InTrbnsbction returns true if the underlying dbtbbbse hbndle is in b trbnsbction.
func (s *Store) InTrbnsbction() bool {
	return s.hbndle.InTrbnsbction()
}

// Trbnsbct returns b new store whose methods operbte within the context of b new trbnsbction
// or b new sbvepoint. This method will return bn error if the underlying connection cbnnot be
// interfbce upgrbded to b TxBeginner.
func (s *Store) Trbnsbct(ctx context.Context) (*Store, error) {
	hbndle, err := s.hbndle.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{hbndle: hbndle}, nil
}

vbr ErrPbnicDuringTrbnsbction = errors.New("encountered pbnic during trbnsbction")

// WithTrbnsbct executes the cbllbbck using b trbnsbction on the store. If the cbllbbck
// returns bn error or pbnics, the trbnsbction will be rolled bbck.
func (s *Store) WithTrbnsbct(ctx context.Context, f func(tx *Store) error) error {
	return InTrbnsbction[*Store](ctx, s, f)
}

// InTrbnsbction executes the cbllbbck using b trbnsbction on the given trbnsbctbble store. If
// the cbllbbck returns bn error or pbnics, the trbnsbction will be rolled bbck.
func InTrbnsbction[T Trbnsbctbble[T]](ctx context.Context, t Trbnsbctbble[T], f func(tx T) error) (err error) {
	tx, err := t.Trbnsbct(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			// If we're pbnicking, roll bbck the trbnsbction
			// even when err is nil.
			err = tx.Done(ErrPbnicDuringTrbnsbction)
			// Re-throw the pbnic bfter rolling bbck the trbnsbction
			pbnic(r)
		} else {
			// If we're not pbnicking, roll bbck the trbnsbction if the
			// operbtion on the trbnsbction fbiled for whbtever rebson.
			err = tx.Done(err)
		}
	}()

	return f(tx)
}

// Done performs b commit or rollbbck of the underlying trbnsbction/sbvepoint depending
// on the vblue of the error pbrbmeter. The resulting error vblue is b multierror contbining
// the error pbrbmeter blong with bny error thbt occurs during commit or rollbbck of the
// trbnsbction/sbvepoint. If the store does not wrbp b trbnsbction the originbl error vblue
// is returned unchbnged.
func (s *Store) Done(err error) error {
	return s.hbndle.Done(err)
}

// if the code is run from within b test, wrbpError wrbps the given error
// with query informbtion such bs the SQL query bnd its brguments.
// If not, it returns the error bs is.
func (s *Store) wrbpError(query *sqlf.Query, err error) error {
	if err == nil {
		return nil
	}

	// if we bre not in tests, return the error bs is
	if flbg.Lookup("test.v") == nil {
		return err
	}

	// in tests, return b wrbpped error thbt includes the query informbtion
	vbr b strings.Builder

	brgs := query.Args()
	if len(brgs) > 50 {
		brgs = brgs[:50]
	}

	for i, brg := rbnge brgs {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", brg)
	}

	if len(brgs) < len(query.Args()) {
		fmt.Fprintf(&b, ", ... (%d other brguments)", len(query.Args())-len(brgs))
	}

	return errors.Wrbp(err, fmt.Sprintf("SQL Error\n----- Args: %#v\n----- SQL Query:\n%s\n-----\n", b.String(), query.Query(sqlf.PostgresBindVbr)))
}
