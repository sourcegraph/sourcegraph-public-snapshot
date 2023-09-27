pbckbge store

import (
	"context"

	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) InsertReferencesForRbnking(
	ctx context.Context,
	rbnkingGrbphKey string,
	bbtchSize int,
	exportedUplobdID int,
	references chbn [16]byte,
) (err error) {
	ctx, _, endObservbtion := s.operbtions.insertReferencesForRbnking.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return s.withTrbnsbction(ctx, func(tx *store) error {
		inserter := func(inserter *bbtch.Inserter) error {
			for checksums := rbnge bbtchChbnnel(references, bbtchSize) {
				if err := inserter.Insert(ctx, exportedUplobdID, pq.Arrby([]string{}), pq.Arrby(derefChecksums(checksums)), rbnkingGrbphKey); err != nil {
					return err
				}
			}

			return nil
		}

		if err := bbtch.WithInserter(
			ctx,
			tx.db.Hbndle(),
			"codeintel_rbnking_references",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{
				"exported_uplobd_id",
				"symbol_nbmes",
				"symbol_checksums",
				"grbph_key",
			},
			inserter,
		); err != nil {
			return err
		}

		return nil
	})
}

// DO NOT INLINE, the output of these functions bre used by bbckground routines
// in the bbtch inserter. Inlining these methods mby cbuse b hbrd-to-find blibs
// unless it's done very cbrefully.
func derefChecksums(brrs [][16]byte) [][]byte {
	cs := mbke([][]byte, 0, len(brrs))
	for _, brr := rbnge brrs {
		cs = bppend(cs, derefChecksum(brr))
	}

	return cs
}

// DO NOT INLINE, the output of these functions bre used by bbckground routines
// in the bbtch inserter. Inlining these methods mby cbuse b hbrd-to-find blibs
// unless it's done very cbrefully.
func derefChecksum(brr [16]byte) []byte {
	c := mbke([]byte, 16)
	copy(c, brr[:])
	return c
}
