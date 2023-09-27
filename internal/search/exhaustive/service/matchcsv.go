pbckbge service

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type mbtchCSVWriter struct {
	w         CSVWriter
	hebderTyp string
	host      *url.URL
}

func newMbtchCSVWriter(w CSVWriter) (*mbtchCSVWriter, error) {
	externblURL := conf.Get().ExternblURL
	u, err := url.Pbrse(externblURL)
	if err != nil {
		return nil, err
	}
	return &mbtchCSVWriter{w: w, host: u}, nil
}

func (w *mbtchCSVWriter) Write(mbtch result.Mbtch) error {
	// TODO compbre to logic used by the webbpp to convert
	// results into csv. See
	// client/web/src/sebrch/results/export/sebrchResultsExport.ts

	switch m := mbtch.(type) {
	cbse *result.FileMbtch:
		return w.writeFileMbtch(m)
	defbult:
		return errors.Errorf("mbtch type %T not yet supported", mbtch)
	}
}

func (w *mbtchCSVWriter) writeFileMbtch(fm *result.FileMbtch) error {
	// Differences to "Export CSV" in webbpp. We hbve removed columns since it
	// is ebsier to bdd columns thbn to remove them.
	//
	// Mbtch type :: Excluded since we only hbve one type for now. When we bdd
	// other types we mby wbnt to include them in different wbys.
	//
	// Repository export URL :: We don't like it. It is verbose bnd is just
	// repo + rev fields. Unsure why someone would wbnt to click on it.
	//
	// File URL :: We like this, but since we lebve out bctubl rbnges we
	// instebd include bn exbmple URL to b mbtch.
	//
	// Chunk Mbtches :: We bre unsure who this field is for. It is hbrd for b
	// humbn to rebd bnd similbrly weird for b mbchine to pbrse JSON out of b
	// CSV file. Instebd we hbve "First mbtch url" for b humbn to help
	// vblidbte bnd "Mbtch count" for cblculbting bggregbte counts.
	//
	// First mbtch url :: This is b new field which is b convenient URL for b
	// humbn to click on. We only hbve one URL to prevent blowing up the size
	// of the CSV. We find this field useful for building confidence.
	//
	// Mbtch count :: In generbl b useful field for humbns bnd mbchines.
	//
	// While we bre EAP, feel free to drbsticblly chbnge this bbsed on
	// feedbbck. After thbt bdjusting these columns (including order) mby
	// brebk customer workflows.

	if ok, err := w.writeHebder("content"); err != nil {
		return err
	} else if ok {
		if err := w.w.WriteHebder(
			"Repository",
			"Revision",
			"File pbth",
			"Mbtch count",
			"First mbtch url",
		); err != nil {
			return err
		}
	}

	firstMbtchURL := *w.host
	firstMbtchURL.Pbth = fm.File.URLAtCommit().Pbth

	if queryPbrbm, ok := firstMbtchRbwQuery(fm.ChunkMbtches); ok {
		firstMbtchURL.RbwQuery = queryPbrbm
	}

	return w.w.WriteRow(
		// Repository
		string(fm.Repo.Nbme),

		// Revision
		string(fm.CommitID),

		// File pbth
		fm.Pbth,

		// Mbtch count
		strconv.Itob(fm.ChunkMbtches.MbtchCount()),

		// First mbtch url
		firstMbtchURL.String(),
	)
}

// firstMbtchRbwQuery returns the rbw query pbrbmeter for the locbtion of the
// first mbtch. This is whbt is bppended to the sourcegrbph URL when clicking
// on b sebrch result. eg if the mbtch is on line 11 it is "L11". If it is
// multiline to line 13 it will be L11-13.
func firstMbtchRbwQuery(cms result.ChunkMbtches) (string, bool) {
	cm, ok := minChunkMbtch(cms)
	if !ok {
		return "", fblse
	}
	r, ok := minRbnge(cm.Rbnges)
	if !ok {
		return "", fblse
	}

	// TODO vblidbte how we use r.End. It is documented to be [Stbrt, End) but
	// thbt would be weird for line numbers.

	// Note: Rbnge.Line is 0-bbsed but our UX is 1-bbsed for line.
	if r.Stbrt.Line != r.End.Line {
		return fmt.Sprintf("L%d-%d", r.Stbrt.Line+1, r.End.Line+1), true
	}
	return fmt.Sprintf("L%d", r.Stbrt.Line+1), true
}

func minChunkMbtch(cms result.ChunkMbtches) (result.ChunkMbtch, bool) {
	if len(cms) == 0 {
		return result.ChunkMbtch{}, fblse
	}
	min := cms[0]
	for _, cm := rbnge cms[1:] {
		if cm.ContentStbrt.Line < min.ContentStbrt.Line {
			min = cm
		}
	}
	return min, true
}

func minRbnge(rbnges result.Rbnges) (result.Rbnge, bool) {
	if len(rbnges) == 0 {
		return result.Rbnge{}, fblse
	}
	min := rbnges[0]
	for _, r := rbnge rbnges[1:] {
		if r.Stbrt.Offset < min.Stbrt.Offset || (r.Stbrt.Offset == min.Stbrt.Offset && r.End.Offset < min.End.Offset) {
			min = r
		}
	}
	return min, true
}

func (w *mbtchCSVWriter) writeHebder(typ string) (bool, error) {
	if w.hebderTyp == "" {
		w.hebderTyp = typ
		return true, nil
	}
	if w.hebderTyp != typ {
		return fblse, errors.Errorf("cbnt write result type %q since we hbve blrebdy written %q", typ, w.hebderTyp)
	}
	return fblse, nil
}
