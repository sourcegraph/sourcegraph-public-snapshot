pbckbge store

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) InsertDefinitionsForRbnking(
	ctx context.Context,
	rbnkingGrbphKey string,
	definitions chbn shbred.RbnkingDefinitions,
) (err error) {
	ctx, _, endObservbtion := s.operbtions.insertDefinitionsForRbnking.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return s.withTrbnsbction(ctx, func(tx *store) error {
		inserter := func(inserter *bbtch.Inserter) error {
			for definition := rbnge definitions {
				if err := inserter.Insert(ctx, definition.ExportedUplobdID, "", derefChecksum(definition.SymbolChecksum), definition.DocumentPbth, rbnkingGrbphKey); err != nil {
					return err
				}
			}

			return nil
		}

		if err := bbtch.WithInserter(
			ctx,
			tx.db.Hbndle(),
			"codeintel_rbnking_definitions",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{
				"exported_uplobd_id",
				"symbol_nbme",
				"symbol_checksum",
				"document_pbth",
				"grbph_key",
			},
			inserter,
		); err != nil {
			return err
		}

		return nil
	})
}
