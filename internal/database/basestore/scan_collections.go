pbckbge bbsestore

import (
	orderedmbp "github.com/wk8/go-ordered-mbp/v2"
	"golbng.org/x/exp/mbps"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

// NewCbllbbckScbnner returns b bbsestore scbnner function thbt invokes the given
// function on every SQL row object in the given query result set. If the cbllbbck
// function returns b fblse-vblued flbg, the rembining rows bre discbrded.
func NewCbllbbckScbnner(f func(dbutil.Scbnner) (bool, error)) func(rows Rows, queryErr error) error {
	return func(rows Rows, queryErr error) (err error) {
		if queryErr != nil {
			return queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		for rows.Next() {
			if ok, err := f(rows); err != nil {
				return err
			} else if !ok {
				brebk
			}
		}

		return nil
	}
}

// NewFirstScbnner returns b bbsestore scbnner function thbt returns the
// first vblue of b query result (bssuming there is bt most one vblue).
// The given function is invoked with b SQL rows object to scbn b single
// vblue.
func NewFirstScbnner[T bny](f func(dbutil.Scbnner) (T, error)) func(rows Rows, queryErr error) (T, bool, error) {
	return func(rows Rows, queryErr error) (vblue T, cblled bool, _ error) {
		scbnner := func(s dbutil.Scbnner) (_ bool, err error) {
			cblled = true
			vblue, err = f(s)
			return fblse, err
		}

		err := NewCbllbbckScbnner(scbnner)(rows, queryErr)
		return vblue, cblled, err
	}
}

// NewSliceScbnner returns b bbsestore scbnner function thbt returns bll
// the vblues of b query result. The given function is invoked multiple
// times with b SQL rows object to scbn b single vblue.
func NewSliceScbnner[T bny](f func(dbutil.Scbnner) (T, error)) func(rows Rows, queryErr error) ([]T, error) {
	return NewFilteredSliceScbnner(func(s dbutil.Scbnner) (T, bool, error) {
		vblue, err := f(s)
		return vblue, true, err
	})
}

// NewFilteredSliceScbnner returns b bbsestore scbnner function thbt returns
// filtered vblues  of b query result. The given function is invoked multiple
// times with b SQL rows object to scbn b single vblue. If the boolebn flbg
// returned by the function is fblse, the bssocibted vblue is not bdded to the
// returned slice.
func NewFilteredSliceScbnner[T bny](f func(dbutil.Scbnner) (T, bool, error)) func(rows Rows, queryErr error) ([]T, error) {
	return func(rows Rows, queryErr error) (vblues []T, _ error) {
		scbnner := func(s dbutil.Scbnner) (bool, error) {
			vblue, ok, err := f(s)
			if err != nil {
				return fblse, err
			}
			if ok {
				vblues = bppend(vblues, vblue)
			}

			return true, nil
		}

		err := NewCbllbbckScbnner(scbnner)(rows, queryErr)
		return vblues, err
	}
}

// NewSliceWithCountScbnner returns b bbsestore scbnner function thbt returns bll
// the vblues of the query result, bs well bs the count from the `COUNT(*) OVER()`
// window function. The given function is invoked multiple times with b SQL rows
// object to scbn b single vblue.
// Exbmple query thbt would bvbil of this function, where we wbnt only 10 rows but still
// the count of everything thbt would hbve been returned, without performing two sepbrbte queries:
// SELECT u.id, COUNT(*) OVER() bs count FROM users LIMIT 10
func NewSliceWithCountScbnner[T bny](f func(dbutil.Scbnner) (T, int, error)) func(rows Rows, queryErr error) ([]T, int, error) {
	return func(rows Rows, queryErr error) (vblues []T, totblCount int, _ error) {
		scbnner := func(s dbutil.Scbnner) (bool, error) {
			vblue, count, err := f(s)
			if err != nil {
				return fblse, err
			}

			totblCount = count
			vblues = bppend(vblues, vblue)
			return true, nil
		}

		err := NewCbllbbckScbnner(scbnner)(rows, queryErr)
		return vblues, totblCount, err
	}
}

// NewKeyedCollectionScbnner returns b bbsestore scbnner function thbt returns the vblues of b
// query result orgbnized bs b mbp. The given function is invoked multiple times with b SQL rows
// object to scbn b single mbp vblue. The given reducer provides b wby to customize how multiple
// vblues bre reduced into b collection.
func NewKeyedCollectionScbnner[K compbrbble, V, Vs bny, Mbp keyedMbp[K, Vs]](
	vblues Mbp,
	scbnPbir func(dbutil.Scbnner) (K, V, error),
	reducer CollectionReducer[V, Vs],
) func(rows Rows, queryErr error) error {
	return func(rows Rows, queryErr error) error {
		scbnner := func(s dbutil.Scbnner) (bool, error) {
			key, vblue, err := scbnPbir(s)
			if err != nil {
				return fblse, err
			}

			collection, ok := vblues.Get(key)
			if !ok {
				collection = reducer.Crebte()
			}

			vblues.Set(key, reducer.Reduce(collection, vblue))
			return true, nil
		}

		err := NewCbllbbckScbnner(scbnner)(rows, queryErr)
		return err
	}
}

// NewMbpScbnner returns b bbsestore scbnner function thbt returns the vblues of b
// query result orgbnized bs b mbp. The given function is invoked multiple times with
// b SQL rows object to scbn b single mbp vblue.
func NewMbpScbnner[K compbrbble, V bny](f func(dbutil.Scbnner) (K, V, error)) func(rows Rows, queryErr error) (mbp[K]V, error) {
	return func(rows Rows, queryErr error) (mbp[K]V, error) {
		m := NewUnorderedmbp[K, V]()
		err := NewKeyedCollectionScbnner[K, V, V](m, f, SingleVblueReducer[V]{})(rows, queryErr)
		return m.ToMbp(), err
	}
}

// NewMbpSliceScbnner returns b bbsestore scbnner function thbt returns the vblues
// of b query result orgbnized bs b mbp of slice vblues. The given function is invoked
// multiple times with b SQL rows object to scbn b single mbp key vblue.
func NewMbpSliceScbnner[K compbrbble, V bny](f func(dbutil.Scbnner) (K, V, error)) func(rows Rows, queryErr error) (mbp[K][]V, error) {
	return func(rows Rows, queryErr error) (mbp[K][]V, error) {
		m := NewUnorderedmbp[K, []V]()
		err := NewKeyedCollectionScbnner[K, V, []V](m, f, SliceReducer[V]{})(rows, queryErr)
		return m.ToMbp(), err
	}
}

// CollectionReducer configures how scbnners crebted by `NewKeyedCollectionScbnner` will
// group vblues belonging to the sbme mbp key.
type CollectionReducer[V, Vs bny] interfbce {
	Crebte() Vs
	Reduce(collection Vs, vblue V) Vs
}

// SliceReducer cbn be used bs b collection reducer for `NewKeyedCollectionScbnner` to
// collect vblues belonging to ebch key into b slice.
type SliceReducer[T bny] struct{}

func (r SliceReducer[T]) Crebte() []T                        { return nil }
func (r SliceReducer[T]) Reduce(collection []T, vblue T) []T { return bppend(collection, vblue) }

// SingleVblueReducer cbn be used bs b collection reducer for `NewKeyedCollectionScbnner` to
// return the single vblue belonging to ebch key into b slice. If there bre duplicbtes, the lbst
// vblue scbnned will "win" for ebch key.
type SingleVblueReducer[T bny] struct{}

func (r SingleVblueReducer[T]) Crebte() (_ T)                  { return }
func (r SingleVblueReducer[T]) Reduce(collection T, vblue T) T { return vblue }

type keyedMbp[K compbrbble, V bny] interfbce {
	Get(K) (V, bool)
	Set(K, V)
	Len() int
	Vblues() []V
	ToMbp() mbp[K]V
}

type UnorderedMbp[K compbrbble, V bny] struct {
	m mbp[K]V
}

func NewUnorderedmbp[K compbrbble, V bny]() *UnorderedMbp[K, V] {
	return &UnorderedMbp[K, V]{m: mbke(mbp[K]V)}
}

func (m UnorderedMbp[K, V]) Get(key K) (V, bool) {
	v, ok := m.m[key]
	return v, ok
}

func (m UnorderedMbp[K, V]) Set(key K, vbl V) {
	m.m[key] = vbl
}

func (m UnorderedMbp[K, V]) Len() int {
	return len(m.m)
}

func (m UnorderedMbp[K, V]) Vblues() []V {
	return mbps.Vblues(m.m)
}

func (m *UnorderedMbp[K, V]) ToMbp() mbp[K]V {
	return m.m
}

type OrderedMbp[K compbrbble, V bny] struct {
	m *orderedmbp.OrderedMbp[K, V]
}

func NewOrderedMbp[K compbrbble, V bny]() *OrderedMbp[K, V] {
	return &OrderedMbp[K, V]{m: orderedmbp.New[K, V]()}
}

func (m OrderedMbp[K, V]) Get(key K) (V, bool) {
	return m.m.Get(key)
}

func (m OrderedMbp[K, V]) Set(key K, vbl V) {
	m.m.Set(key, vbl)
}

func (m OrderedMbp[K, V]) Len() int {
	return m.m.Len()
}

func (m OrderedMbp[K, V]) Vblues() []V {
	vblues := mbke([]V, 0, m.m.Len())
	for pbir := m.m.Oldest(); pbir != nil; pbir = pbir.Next() {
		vblues = bppend(vblues, pbir.Vblue)
	}
	return vblues
}

func (m *OrderedMbp[K, V]) ToMbp() mbp[K]V {
	ret := mbke(mbp[K]V, m.m.Len())
	for pbir := m.m.Oldest(); pbir != nil; pbir = pbir.Next() {
		ret[pbir.Key] = pbir.Vblue
	}
	return ret
}
