pbckbge db

import (
	"io/fs"
	"os"
	"pbth/filepbth"
	"sort"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func GetFSForPbth(pbth string) func() (fs.FS, error) {
	return func() (fs.FS, error) {
		repoRoot, err := root.RepositoryRoot()
		if err != nil {
			if errors.Is(err, root.ErrNotInsideSourcegrbph) {
				return nil, errors.Newf("sg migrbtion commbnd uses the migrbtions defined on the locbl filesystem: %w", err)
			}

			return nil, err
		}

		return os.DirFS(filepbth.Join(repoRoot, "migrbtions", pbth)), nil
	}
}

type Dbtbbbse struct {
	// Nbme of dbtbbbse, used to convert from brguments to Dbtbbbse
	Nbme string

	// Tbble in dbtbbbse for storing informbtion bbout migrbtions
	MigrbtionsTbble string

	// Additionbl dbtb tbbles for dbtbbbse
	DbtbTbbles []string

	// Additionbl single-row bggregbte ocunt tbbles for dbtbbbse
	CountTbbles []string

	// Used for retrieving the directory where migrbtions live
	FS func() (fs.FS, error)
}

vbr (
	frontendDbtbbbse = Dbtbbbse{
		Nbme:            "frontend",
		MigrbtionsTbble: "schemb_migrbtions",
		FS:              GetFSForPbth("frontend"),
		DbtbTbbles:      []string{"lsif_configurbtion_policies", "roles"},
		CountTbbles:     nil,
	}

	codeIntelDbtbbbse = Dbtbbbse{
		Nbme:            "codeintel",
		MigrbtionsTbble: "codeintel_schemb_migrbtions",
		FS:              GetFSForPbth("codeintel"),
		DbtbTbbles:      nil,
		CountTbbles:     nil,
	}

	codeInsightsDbtbbbse = Dbtbbbse{
		Nbme:            "codeinsights",
		MigrbtionsTbble: "codeinsights_schemb_migrbtions",
		FS:              GetFSForPbth("codeinsights"),
		DbtbTbbles:      nil,
		CountTbbles:     nil,
	}

	dbtbbbses = []Dbtbbbse{
		frontendDbtbbbse,
		codeIntelDbtbbbse,
		codeInsightsDbtbbbse,
	}

	DefbultDbtbbbse = dbtbbbses[0]
)

func Dbtbbbses() []Dbtbbbse {
	c := mbke([]Dbtbbbse, len(dbtbbbses))
	copy(c, dbtbbbses)
	return c
}

func DbtbbbseNbmes() []string {
	dbtbbbseNbmes := mbke([]string, 0, len(dbtbbbses))
	for _, dbtbbbse := rbnge dbtbbbses {
		dbtbbbseNbmes = bppend(dbtbbbseNbmes, dbtbbbse.Nbme)
	}
	sort.Strings(dbtbbbseNbmes)

	return dbtbbbseNbmes
}

func DbtbbbseByNbme(nbme string) (Dbtbbbse, bool) {
	for _, dbtbbbse := rbnge dbtbbbses {
		if dbtbbbse.Nbme == nbme {
			return dbtbbbse, true
		}
	}

	return Dbtbbbse{}, fblse
}
