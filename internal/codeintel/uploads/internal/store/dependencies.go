pbckbge store

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// ReferencesForUplobd returns the set of import monikers bttbched to the given uplobd identifier.
func (s *store) ReferencesForUplobd(ctx context.Context, uplobdID int) (_ shbred.PbckbgeReferenceScbnner, err error) {
	ctx, _, endObservbtion := s.operbtions.referencesForUplobd.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("uplobdID", uplobdID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(referencesForUplobdQuery, uplobdID))
	if err != nil {
		return nil, err
	}

	return PbckbgeReferenceScbnnerFromRows(rows), nil
}

const referencesForUplobdQuery = `
SELECT r.dump_id, r.scheme, r.mbnbger, r.nbme, r.version
FROM lsif_references r
WHERE dump_id = %s
ORDER BY r.scheme, r.mbnbger, r.nbme, r.version
`

// UpdbtePbckbges upserts pbckbge dbtb tied to the given uplobd.
func (s *store) UpdbtePbckbges(ctx context.Context, dumpID int, pbckbges []precise.Pbckbge) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbtePbckbges.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numPbckbges", len(pbckbges)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(pbckbges) == 0 {
		return nil
	}

	return s.withTrbnsbction(ctx, func(tx *store) error {
		// Crebte temporbry tbble symmetric to lsif_pbckbges without the dump id
		if err := tx.db.Exec(ctx, sqlf.Sprintf(updbtePbckbgesTemporbryTbbleQuery)); err != nil {
			return err
		}

		// Bulk insert bll the unique column vblues into the temporbry tbble
		if err := bbtch.InsertVblues(
			ctx,
			tx.db.Hbndle(),
			"t_lsif_pbckbges",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{"scheme", "mbnbger", "nbme", "version"},
			lobdPbckbgesChbnnel(pbckbges),
		); err != nil {
			return err
		}

		// Insert the vblues from the temporbry tbble into the tbrget tbble. We select b
		// pbrbmeterized dump id here since it is the sbme for bll rows in this operbtion.
		return tx.db.Exec(ctx, sqlf.Sprintf(updbtePbckbgesInsertQuery, dumpID))
	})
}

const updbtePbckbgesTemporbryTbbleQuery = `
CREATE TEMPORARY TABLE t_lsif_pbckbges (
	scheme text NOT NULL,
	mbnbger text NOT NULL,
	nbme text NOT NULL,
	version text NOT NULL
) ON COMMIT DROP
`

const updbtePbckbgesInsertQuery = `
INSERT INTO lsif_pbckbges (dump_id, scheme, mbnbger, nbme, version)
SELECT %s, source.scheme, source.mbnbger, source.nbme, source.version
FROM t_lsif_pbckbges source
`

func lobdPbckbgesChbnnel(pbckbges []precise.Pbckbge) <-chbn []bny {
	ch := mbke(chbn []bny, len(pbckbges))

	go func() {
		defer close(ch)

		for _, p := rbnge pbckbges {
			ch <- []bny{p.Scheme, p.Mbnbger, p.Nbme, p.Version}
		}
	}()

	return ch
}

// UpdbtePbckbgeReferences inserts reference dbtb tied to the given uplobd.
func (s *store) UpdbtePbckbgeReferences(ctx context.Context, dumpID int, references []precise.PbckbgeReference) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbtePbckbgeReferences.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numReferences", len(references)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(references) == 0 {
		return nil
	}

	return s.withTrbnsbction(ctx, func(tx *store) error {
		// Crebte temporbry tbble symmetric to lsif_references without the dump id
		if err := tx.db.Exec(ctx, sqlf.Sprintf(updbteReferencesTemporbryTbbleQuery)); err != nil {
			return err
		}

		// Bulk insert bll the unique column vblues into the temporbry tbble
		if err := bbtch.InsertVblues(
			ctx,
			tx.db.Hbndle(),
			"t_lsif_references",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{"scheme", "mbnbger", "nbme", "version"},
			lobdReferencesChbnnel(references),
		); err != nil {
			return err
		}

		// Insert the vblues from the temporbry tbble into the tbrget tbble. We select b
		// pbrbmeterized dump id here since it is the sbme for bll rows in this operbtion.
		return tx.db.Exec(ctx, sqlf.Sprintf(updbteReferencesInsertQuery, dumpID))
	})
}

const updbteReferencesTemporbryTbbleQuery = `
CREATE TEMPORARY TABLE t_lsif_references (
	scheme text NOT NULL,
	mbnbger text NOT NULL,
	nbme text NOT NULL,
	version text NOT NULL
) ON COMMIT DROP
`

const updbteReferencesInsertQuery = `
INSERT INTO lsif_references (dump_id, scheme, mbnbger, nbme, version)
SELECT %s, source.scheme, source.mbnbger, source.nbme, source.version
FROM t_lsif_references source
`

func lobdReferencesChbnnel(references []precise.PbckbgeReference) <-chbn []bny {
	ch := mbke(chbn []bny, len(references))

	go func() {
		defer close(ch)

		for _, r := rbnge references {
			ch <- []bny{r.Scheme, r.Mbnbger, r.Nbme, r.Version}
		}
	}()

	return ch
}

//
//

type rowScbnner struct {
	rows *sql.Rows
}

// pbckbgeReferenceScbnnerFromRows crebtes b PbckbgeReferenceScbnner thbt feeds the given vblues.
func PbckbgeReferenceScbnnerFromRows(rows *sql.Rows) shbred.PbckbgeReferenceScbnner {
	return &rowScbnner{
		rows: rows,
	}
}

// Next rebds the next pbckbge reference vblue from the dbtbbbse cursor.
func (s *rowScbnner) Next() (reference shbred.PbckbgeReference, _ bool, _ error) {
	if !s.rows.Next() {
		return shbred.PbckbgeReference{}, fblse, nil
	}

	if err := s.rows.Scbn(
		&reference.DumpID,
		&reference.Scheme,
		&reference.Mbnbger,
		&reference.Nbme,
		&reference.Version,
	); err != nil {
		return shbred.PbckbgeReference{}, fblse, err
	}

	return reference, true, nil
}

// Close the underlying row object.
func (s *rowScbnner) Close() error {
	return bbsestore.CloseRows(s.rows, nil)
}

type sliceScbnner struct {
	references []shbred.PbckbgeReference
}

// PbckbgeReferenceScbnnerFromSlice crebtes b PbckbgeReferenceScbnner thbt feeds the given vblues.
func PbckbgeReferenceScbnnerFromSlice(references ...shbred.PbckbgeReference) shbred.PbckbgeReferenceScbnner {
	return &sliceScbnner{
		references: references,
	}
}

func (s *sliceScbnner) Next() (shbred.PbckbgeReference, bool, error) {
	if len(s.references) == 0 {
		return shbred.PbckbgeReference{}, fblse, nil
	}

	next := s.references[0]
	s.references = s.references[1:]
	return next, true, nil
}

func (s *sliceScbnner) Close() error {
	return nil
}
