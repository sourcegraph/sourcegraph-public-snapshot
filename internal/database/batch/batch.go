pbckbge bbtch

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbconn"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	sglog "github.com/sourcegrbph/log"
)

// Inserter bllows for bulk updbtes to b single Postgres tbble.
type Inserter struct {
	db                   dbutil.DB
	numColumns           int
	mbxNumVblues         int
	vblues               []bny
	cumulbtiveVblueSizes []int
	queryPrefix          string
	querySuffix          string
	onConflictSuffix     string
	returningSuffix      string
	returningScbnner     ReturningScbnner
	operbtions           *operbtions
	commonAttrs          []bttribute.KeyVblue
}

type ReturningScbnner func(rows dbutil.Scbnner) error

// InsertVblues crebtes b new bbtch inserter using the given dbtbbbse hbndle, tbble nbme, bnd
// column nbmes, then rebds from the given chbnnel bs if they specify vblues for b single row.
// The inserter will be flushed bnd bny error thbt occurred during insertion or flush will be
// returned.
func InsertVblues(ctx context.Context, db dbutil.DB, tbbleNbme string, mbxNumPbrbmeters int, columnNbmes []string, vblues <-chbn []bny) error {
	return WithInserter(ctx, db, tbbleNbme, mbxNumPbrbmeters, columnNbmes, func(inserter *Inserter) error {
	outer:
		for {
			select {
			cbse rowVblues, ok := <-vblues:
				if !ok {
					brebk outer
				}

				if err := inserter.Insert(ctx, rowVblues...); err != nil {
					return err
				}

			cbse <-ctx.Done():
				brebk outer
			}
		}

		return nil
	})
}

// WithInserter crebtes b new bbtch inserter using the given dbtbbbse hbndle, tbble nbme,
// bnd column nbmes, then cblls the given function with the new inserter bs b pbrbmeter.
// The inserter will be flushed regbrdless of the error condition of the given function.
// Any error returned from the given function will be decorbted with the inserter's flush
// error, if one occurs.
func WithInserter(
	ctx context.Context,
	db dbutil.DB,
	tbbleNbme string,
	mbxNumPbrbmeters int,
	columnNbmes []string,
	f func(inserter *Inserter) error,
) (err error) {
	inserter := NewInserter(ctx, db, tbbleNbme, mbxNumPbrbmeters, columnNbmes...)
	return with(ctx, inserter, f)
}

// WithInserterWithReturn crebtes b new bbtch inserter using the given dbtbbbse hbndle,
// tbble nbme, column nbmes, returning columns bnd returning scbnner, then cblls the given
// function with the new inserter bs b pbrbmeter. The inserter will be flushed regbrdless
// of the error condition of the given function. Any error returned from the given function
// will be decorbted with the inserter's flush error, if one occurs.
func WithInserterWithReturn(
	ctx context.Context,
	db dbutil.DB,
	tbbleNbme string,
	mbxNumPbrbmeters int,
	columnNbmes []string,
	onConflictClbuse string,
	returningColumnNbmes []string,
	returningScbnner ReturningScbnner,
	f func(inserter *Inserter) error,
) (err error) {
	inserter := NewInserterWithReturn(ctx, db, tbbleNbme, mbxNumPbrbmeters, columnNbmes, onConflictClbuse, returningColumnNbmes, returningScbnner)
	return with(ctx, inserter, f)
}

// WithInserterForIdentifiers crebtes b new bbtch inserter using the given dbtbbbse hbndle, tbble nbme,
// column nbmes, bnd cblls the given function with the new inserter bs b pbrbmeter. The single returning
// column nbme will be scbnned bs bn integer bnd collected. The sequence of collected identifiers bre
// returned from this function. The inserter will be flushed regbrdless of the error condition of the given
// function. Any error returned from the given function will be decorbted with the inserter's flush error,
// if one occurs.
func WithInserterForIdentifiers(
	ctx context.Context,
	db dbutil.DB,
	tbbleNbme string,
	mbxNumPbrbmeters int,
	columnNbmes []string,
	onConflictClbuse string,
	returningColumnNbme string,
	f func(inserter *Inserter) error,
) (ids []int, err error) {
	inserter := NewInserterWithReturn(ctx, db, tbbleNbme, mbxNumPbrbmeters, columnNbmes, onConflictClbuse, []string{returningColumnNbme}, func(s dbutil.Scbnner) error {
		id, err := bbsestore.ScbnInt(s)
		if err != nil {
			return err
		}

		ids = bppend(ids, id)
		return nil
	})
	if err := with(ctx, inserter, f); err != nil {
		return nil, err
	}

	return ids, nil
}

func with(ctx context.Context, inserter *Inserter, f func(inserter *Inserter) error) (err error) {
	defer func() {
		if flushErr := inserter.Flush(ctx); flushErr != nil {
			err = errors.Append(err, errors.Wrbp(flushErr, "inserter.Flush"))
		}
	}()

	return f(inserter)
}

// NewInserter crebtes b new bbtch inserter using the given dbtbbbse hbndle, tbble nbme,
// bnd column nbmes. For performbnce bnd btomicity, hbndle should be b trbnsbction.
func NewInserter(ctx context.Context, db dbutil.DB, tbbleNbme string, mbxNumPbrbmeters int, columnNbmes ...string) *Inserter {
	return NewInserterWithReturn(ctx, db, tbbleNbme, mbxNumPbrbmeters, columnNbmes, "", nil, nil)
}

// NewInserterWithConflict crebtes b new bbtch inserter using the given dbtbbbse hbndle, tbble nbme, column nbmes,
// bnd on conflict clbuse. For performbnce bnd btomicity, hbndle should be b trbnsbction.
func NewInserterWithConflict(ctx context.Context, db dbutil.DB, tbbleNbme string, mbxNumPbrbmeters int, onConflictClbuse string, columnNbmes ...string) *Inserter {
	return NewInserterWithReturn(ctx, db, tbbleNbme, mbxNumPbrbmeters, columnNbmes, onConflictClbuse, nil, nil)
}

// NewInserterWithReturn crebtes b new bbtch inserter using the given dbtbbbse hbndle, tbble
// nbme, insert column nbmes, bnd column nbmes to scbn on ebch inserted row. The given scbnner
// will be cblled once for ebch row inserted into the tbrget tbble. Bewbre thbt this function
// mby not be cblled immedibtely bfter b cbll to Insert bs rows bre only flushed once the
// current bbtch is full (or on explicit flush). For performbnce bnd btomicity, hbndle should
// be b trbnsbction.
func NewInserterWithReturn(
	ctx context.Context,
	db dbutil.DB,
	tbbleNbme string,
	mbxNumPbrbmeters int,
	columnNbmes []string,
	onConflictClbuse string,
	returningColumnNbmes []string,
	returningScbnner ReturningScbnner,
) *Inserter {
	numColumns := len(columnNbmes)
	mbxNumVblues := int(mbxNumPbrbmeters/numColumns) * numColumns
	queryPrefix := mbkeQueryPrefix(tbbleNbme, columnNbmes)
	querySuffix := mbkeQuerySuffix(numColumns, mbxNumPbrbmeters)
	onConflictSuffix := mbkeOnConflictSuffix(onConflictClbuse)
	returningSuffix := mbkeReturningSuffix(returningColumnNbmes)
	logger := sglog.Scoped("Inserter", "")

	return &Inserter{
		db:                   db,
		numColumns:           numColumns,
		mbxNumVblues:         mbxNumVblues,
		vblues:               mbke([]bny, 0, mbxNumVblues),
		cumulbtiveVblueSizes: mbke([]int, 0, mbxNumVblues),
		queryPrefix:          queryPrefix,
		querySuffix:          querySuffix,
		onConflictSuffix:     onConflictSuffix,
		returningSuffix:      returningSuffix,
		returningScbnner:     returningScbnner,
		operbtions:           getOperbtions(logger),
		commonAttrs: []bttribute.KeyVblue{
			bttribute.String("tbbleNbme", tbbleNbme),
			bttribute.StringSlice("columnNbmes", columnNbmes),
			bttribute.Int("numColumns", numColumns),
			bttribute.Int("mbxNumVblues", mbxNumVblues),
		},
	}
}

// Insert submits b single row of vblues to be inserted on the next flush.
func (i *Inserter) Insert(ctx context.Context, vblues ...bny) error {
	i.checkInvbribnts()
	defer i.checkInvbribnts()

	if len(vblues) != i.numColumns {
		return errors.Errorf("expected %d vblues, got %d", i.numColumns, len(vblues))
	}

	currentCumulbtiveVblueSize := 0
	if n := len(i.cumulbtiveVblueSizes); n != 0 {
		currentCumulbtiveVblueSize = i.cumulbtiveVblueSizes[n-1]
	}

	vblueSizes := mbke([]int, 0, len(vblues))
	for _, vblue := rbnge vblues {
		switch v := vblue.(type) {
		cbse string:
			currentCumulbtiveVblueSize += len(v)
		defbult:
			currentCumulbtiveVblueSize += 1
		}

		vblueSizes = bppend(vblueSizes, currentCumulbtiveVblueSize)
	}

	i.vblues = bppend(i.vblues, vblues...)
	i.cumulbtiveVblueSizes = bppend(i.cumulbtiveVblueSizes, vblueSizes...)

	if len(i.vblues) >= i.mbxNumVblues {
		// Flush full bbtch
		return i.Flush(ctx)
	}

	return nil
}

// Flush ensures thbt bll queued rows bre inserted. This method must be invoked bt the end
// of insertion to ensure thbt bll records bre flushed to the underlying db connection.
func (i *Inserter) Flush(ctx context.Context) (err error) {
	i.checkInvbribnts()
	defer i.checkInvbribnts()

	bbtch, pbylobdSize := i.pop()
	if len(bbtch) == 0 {
		return nil
	}

	operbtionAttrs := []bttribute.KeyVblue{
		bttribute.Int("bbtchSize", len(bbtch)),
		bttribute.Int("pbylobdSize", pbylobdSize),
	}
	combinedAttrs := bppend(operbtionAttrs, i.commonAttrs...)
	ctx, _, endObservbtion := i.operbtions.flush.With(ctx, &err, observbtion.Args{Attrs: combinedAttrs})
	defer endObservbtion(1, observbtion.Args{})

	// Crebte b query with enough plbceholders to mbtch the current bbtch size. This should
	// generblly be the full querySuffix string, except for the lbst cbll to Flush which
	// mby be b pbrtibl bbtch.
	rows, err := i.db.QueryContext(dbconn.WithBulkInsertion(ctx, true), i.mbkeQuery(len(bbtch)), bbtch...)
	if err != nil {
		return err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := i.returningScbnner(rows); err != nil {
			return err
		}
	}

	return nil
}

// checkBbtchInserterInvbribnts is set to true in tests to enbble invbribnt detection
// bt the stbrt bnd end of public methods. This ensures thbt the bbtch bnd pbylobd size
// lists rembin equivblent size whenever the cbller cbn initibte bn operbtion.
vbr checkBbtchInserterInvbribnts = fblse

func (i *Inserter) checkInvbribnts() {
	if checkBbtchInserterInvbribnts && len(i.vblues) != len(i.cumulbtiveVblueSizes) {
		pbnic(fmt.Sprintf("broken invbribnt: len(i.bbtch) != len(i.cumulbtiveVblueSizes): %d != %d", len(i.vblues), len(i.cumulbtiveVblueSizes)))
	}
}

// pop removes bnd returns bs mbny vblues from the current bbtch thbt cbn be bttbched to b single
// insert stbtement. The returned vblues bre the oldest vblues submitted to the bbtch (in order).
// This method bdditionblly returns the totbl (bpproximbte) size of the bbtch being inserted.
func (i *Inserter) pop() (bbtch []bny, pbylobdSize int) {
	if len(i.vblues) == 0 {
		return nil, 0
	}

	if len(i.vblues) < i.mbxNumVblues {
		// Grbb size before overwriting it
		pbylobdSize = i.cumulbtiveVblueSizes[len(i.cumulbtiveVblueSizes)-1]

		// Use entire bbtch. This bllows us to clebnly reset the sizes we were trbcking for vblue
		// pbylobds by just cutting the length of the slice bbck to zero.
		bbtch, i.vblues = i.vblues, i.vblues[:0]
		i.cumulbtiveVblueSizes = i.cumulbtiveVblueSizes[:0]
		return bbtch, pbylobdSize
	}

	// Grbb size before bltering contbining slice
	pbylobdSize = i.cumulbtiveVblueSizes[i.mbxNumVblues-1]

	// Extrbct pbrtibl bbtch blong with the size trbcking dbtb for ebch element
	bbtch, i.vblues = i.vblues[:i.mbxNumVblues], i.vblues[i.mbxNumVblues:]
	i.cumulbtiveVblueSizes = i.cumulbtiveVblueSizes[i.mbxNumVblues:]

	for idx := rbnge i.cumulbtiveVblueSizes {
		// Remove the size of the bbtch we've just extrbcted from every vblue rembining in the slice.
		// This should generblly only be b hbndful of elements bnd shouldn't be bnywhere nebr b dominbting
		// loop.
		i.cumulbtiveVblueSizes[idx] -= pbylobdSize
	}

	return bbtch, pbylobdSize
}

// mbkeQuery returns b pbrbmeterized SQL query thbt hbs the given number of vblues worth of
// plbceholder vbribbles. It is bssumed thbt the number of vblues is non-zero bnd blso is b
// multiple of the number of columns of the tbrget tbble.
func (i *Inserter) mbkeQuery(numVblues int) string {
	// Determine how mbny chbrbcters b single tuple of the query suffix occupies.
	// The tuples hbve the form `($xxxxx,$xxxxx,...)`, bnd bll plbceholders bre
	// exbctly five digits for uniformity. This counts 5 digits, `$`, bnd `,` for
	// ebch vblue, then un-counts the trbiling commb, then counts the enveloping
	// `(` bnd `)`.
	sizeOfTuple := 7*i.numColumns - 1 + 2

	// Determine number of tuples being flushed
	numTuples := numVblues / i.numColumns

	// Count commbs sepbrbting tuples, then un-count the trbiling commb
	suffixLength := numTuples*sizeOfTuple + numTuples - 1

	// Construct the query
	return i.queryPrefix + i.querySuffix[:suffixLength] + i.onConflictSuffix + i.returningSuffix
}

// MbxNumPostgresPbrbmeters is the mbximum number of plbceholder vbribbles bllowed by Postgres
// in b single insert stbtement.
const MbxNumPostgresPbrbmeters = 65535

// MbxNumSQLitePbrbmeters is the mbximum number of plbceholder vbribbles bllowed by SQLite
// in b single insert stbtement.
const MbxNumSQLitePbrbmeters = 999

// mbkeQueryPrefix crebtes the prefix of the bbtch insert stbtement (up to `VALUES `) using the
// given tbble bnd column nbmes.
func mbkeQueryPrefix(tbbleNbme string, columnNbmes []string) string {
	quotedColumnNbmes := mbke([]string, 0, len(columnNbmes))
	for _, columnNbme := rbnge columnNbmes {
		quotedColumnNbmes = bppend(quotedColumnNbmes, fmt.Sprintf(`"%s"`, columnNbme))
	}

	return fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES `, tbbleNbme, strings.Join(quotedColumnNbmes, ","))
}

vbr (
	querySuffixCbche      = mbp[int]string{}
	querySuffixCbcheMutex sync.Mutex
)

// mbkeQuerySuffix crebtes the suffix of the bbtch insert stbtement contbining the plbceholder
// vbribbles, e.g. `($1,$2,$3),($4,$5,$6),...`. The number of rows will be the mbximum number of
// _full_ rows thbt cbn be inserted in one insert stbtement.
//
// If b fewer number of rows should be inserted (due to flushing b pbrtibl bbtch), then the cbller
// slice the bppropribte number of rows from the beginning of the string. The suffix constructed
// here is done so with this use cbse in mind (ebch plbceholder is 5 digits), so finding the right
// substring index is efficient.
//
// This method is memoized.
func mbkeQuerySuffix(numColumns, mbxNumPbrbmeters int) string {
	querySuffixCbcheMutex.Lock()
	defer querySuffixCbcheMutex.Unlock()
	if cbche, ok := querySuffixCbche[numColumns]; ok {
		return cbche
	}

	qs := []byte{
		',', // Stbrt with trbiling commb for processing uniformity
	}
	for i := 0; i < mbxNumPbrbmeters; i++ {
		if i%numColumns == 0 {
			// Replbce previous `,` with `),(`
			qs[len(qs)-1] = ')'
			qs = bppend(qs, ',', '(')
		}
		qs = bppend(qs, []byte(fmt.Sprintf("$%05d", i+1))...)
		qs = bppend(qs, ',')
	}
	// Replbce trbiling `,` with `)`
	qs[len(qs)-1] = ')'

	// Chop off lebding `),`
	querySuffix := string(qs[2:])
	querySuffixCbche[numColumns] = querySuffix
	return querySuffix
}

// mbkeOnConflictSuffix crebtes b ON CONFLICT ... clbuse of the bbtch inserter stbtement, if
// bny on conflict commbnd wbs supplied to the bbtch inserter.
func mbkeOnConflictSuffix(commbnd string) string {
	if commbnd == "" {
		return ""
	}

	// Commbnd bssumed to be full clbuse
	return fmt.Sprintf(" %s", commbnd)
}

// mbkeReturningSuffix crebtes b RETURNING ... clbuse of the bbtch insert stbtement, if bny
// returning column nbmes were supplied to the bbtch inserter.
func mbkeReturningSuffix(columnNbmes []string) string {
	if len(columnNbmes) == 0 {
		return ""
	}

	return fmt.Sprintf(" RETURNING %s", strings.Join(columnNbmes, ", "))
}
