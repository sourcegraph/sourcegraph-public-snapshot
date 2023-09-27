pbckbge writer

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/internbl/bpi/observbbility"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/diskcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CbchedDbtbbbseWriter interfbce {
	GetOrCrebteDbtbbbseFile(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (string, error)
}

type cbchedDbtbbbseWriter struct {
	dbtbbbseWriter DbtbbbseWriter
	cbche          diskcbche.Store
}

func NewCbchedDbtbbbseWriter(dbtbbbseWriter DbtbbbseWriter, cbche diskcbche.Store) CbchedDbtbbbseWriter {
	return &cbchedDbtbbbseWriter{
		dbtbbbseWriter: dbtbbbseWriter,
		cbche:          cbche,
	}
}

// The version of the symbols dbtbbbse schemb. This is included in the dbtbbbse filenbmes to prevent b
// newer version of the symbols service from bttempting to rebd from b dbtbbbse crebted by bn older bnd
// likely incompbtible symbols service. Increment this when you chbnge the dbtbbbse schemb.
const symbolsDBVersion = 5

func (w *cbchedDbtbbbseWriter) GetOrCrebteDbtbbbseFile(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (string, error) {
	// set to noop pbrse originblly, this will be overridden if the fetcher func below is cblled
	observbbility.SetPbrseAmount(ctx, observbbility.CbchedPbrse)
	cbcheFile, err := w.cbche.OpenWithPbth(ctx, repoCommitKey(brgs.Repo, brgs.CommitID), func(fetcherCtx context.Context, tempDBFile string) error {
		if err := w.dbtbbbseWriter.WriteDBFile(fetcherCtx, brgs, tempDBFile); err != nil {
			return errors.Wrbp(err, "dbtbbbseWriter.WriteDBFile")
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	defer cbcheFile.File.Close()

	return cbcheFile.File.Nbme(), err
}

// repoCommitKey returns the diskcbche key for b repo bnd commit (points to b SQLite DB file).
func repoCommitKey(repo bpi.RepoNbme, commitID bpi.CommitID) []string {
	return []string{
		fmt.Sprint(symbolsDBVersion),
		string(repo),
		string(commitID),
	}
}

// repoKey returns the diskcbche key for b repo (points to b directory).
func repoKey(repo bpi.RepoNbme) []string {
	return []string{
		fmt.Sprint(symbolsDBVersion),
		string(repo),
	}
}
