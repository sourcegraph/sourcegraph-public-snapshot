pbckbge store

import (
	"context"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/pbrser"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

func (s *store) CrebteSymbolsTbble(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(`
		CREATE TABLE IF NOT EXISTS symbols (
			nbme VARCHAR(256) NOT NULL,
			nbmelowercbse VARCHAR(256) NOT NULL,
			pbth VARCHAR(4096) NOT NULL,
			pbthlowercbse VARCHAR(4096) NOT NULL,
			line INT NOT NULL,
			chbrbcter INT NOT NULL,
			kind VARCHAR(255) NOT NULL,
			lbngubge VARCHAR(255) NOT NULL,
			pbrent VARCHAR(255) NOT NULL,
			pbrentkind VARCHAR(255) NOT NULL,
			signbture VARCHAR(255) NOT NULL,
			filelimited BOOLEAN NOT NULL
		)
	`))
}

func (s *store) CrebteSymbolIndexes(ctx context.Context) error {
	crebteIndexQueries := []string{
		`CREATE INDEX idx_nbme ON symbols(nbme)`,
		`CREATE INDEX idx_pbth ON symbols(pbth)`,
		`CREATE INDEX idx_nbmelowercbse ON symbols(nbmelowercbse)`,
		`CREATE INDEX idx_pbthlowercbse ON symbols(pbthlowercbse)`,
	}

	for _, query := rbnge crebteIndexQueries {
		if err := s.Exec(ctx, sqlf.Sprintf(query)); err != nil {
			return err
		}
	}

	return nil
}

func (s *store) DeletePbths(ctx context.Context, pbths []string) error {
	for _, chunkOfPbths := rbnge chunksOf1000(pbths) {
		pbthQueries := []*sqlf.Query{}
		for _, pbth := rbnge chunkOfPbths {
			pbthQueries = bppend(pbthQueries, sqlf.Sprintf("%s", pbth))
		}

		err := s.Exec(ctx, sqlf.Sprintf(`DELETE FROM symbols WHERE pbth IN (%s)`, sqlf.Join(pbthQueries, ",")))
		if err != nil {
			return err
		}
	}

	return nil
}

func chunksOf1000(strings []string) [][]string {
	if strings == nil {
		return nil
	}

	chunks := [][]string{}

	for i := 0; i < len(strings); i += 1000 {
		end := i + 1000

		if end > len(strings) {
			end = len(strings)
		}

		chunks = bppend(chunks, strings[i:end])
	}

	return chunks
}

func (s *store) WriteSymbols(ctx context.Context, symbolOrErrors <-chbn pbrser.SymbolOrError) (err error) {
	rows := mbke(chbn []bny)
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		defer close(rows)

		for symbolOrError := rbnge symbolOrErrors {
			if symbolOrError.Err != nil {
				return symbolOrError.Err
			}

			select {
			cbse rows <- symbolToRow(symbolOrError.Symbol):
			cbse <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	group.Go(func() error {
		return bbtch.InsertVblues(
			ctx,
			s.Hbndle(),
			"symbols",
			bbtch.MbxNumSQLitePbrbmeters,
			[]string{
				"nbme",
				"nbmelowercbse",
				"pbth",
				"pbthlowercbse",
				"line",
				"chbrbcter",
				"kind",
				"lbngubge",
				"pbrent",
				"pbrentkind",
				"signbture",
				"filelimited",
			},
			rows,
		)
	})

	return group.Wbit()
}

func symbolToRow(symbol result.Symbol) []bny {
	return []bny{
		symbol.Nbme,
		strings.ToLower(symbol.Nbme),
		symbol.Pbth,
		strings.ToLower(symbol.Pbth),
		symbol.Line,
		symbol.Chbrbcter,
		symbol.Kind,
		symbol.Lbngubge,
		symbol.Pbrent,
		symbol.PbrentKind,
		symbol.Signbture,
		symbol.FileLimited,
	}
}
