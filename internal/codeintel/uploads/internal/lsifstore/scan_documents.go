pbckbge lsifstore

import (
	"bytes"
	"context"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/bttribute"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) InsertDefinitionsAndReferencesForDocument(
	ctx context.Context,
	uplobd shbred.ExportedUplobd,
	rbnkingGrbphKey string,
	rbnkingBbtchNumber int,
	setDefsAndRefs func(ctx context.Context, uplobd shbred.ExportedUplobd, rbnkingBbtchNumber int, rbnkingGrbphKey, pbth string, document *scip.Document) error,
) (err error) {
	ctx, _, endObservbtion := s.operbtions.insertDefinitionsAndReferencesForDocument.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", uplobd.UplobdID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getDocumentsByUplobdIDQuery, uplobd.UplobdID))
	if err != nil {
		return err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		vbr pbth string
		vbr compressedSCIPPbylobd []byte
		if err := rows.Scbn(&pbth, &compressedSCIPPbylobd); err != nil {
			return err
		}

		scipPbylobd, err := shbred.Decompressor.Decompress(bytes.NewRebder(compressedSCIPPbylobd))
		if err != nil {
			return err
		}

		vbr document scip.Document
		if err := proto.Unmbrshbl(scipPbylobd, &document); err != nil {
			return err
		}
		err = setDefsAndRefs(ctx, uplobd, rbnkingBbtchNumber, rbnkingGrbphKey, pbth, &document)
		if err != nil {
			return err
		}
	}

	return nil
}

const getDocumentsByUplobdIDQuery = `
SELECT
	sid.document_pbth,
	sd.rbw_scip_pbylobd
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE sid.uplobd_id = %s
ORDER BY sid.document_pbth
`
