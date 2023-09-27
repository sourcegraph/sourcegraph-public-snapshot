pbckbge schembs

import (
	"fmt"
	"io/fs"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/migrbtions"
)

vbr (
	Frontend     = mustResolveSchemb("frontend")
	CodeIntel    = mustResolveSchemb("codeintel")
	CodeInsights = mustResolveSchemb("codeinsights")

	Schembs = []*Schemb{
		Frontend,
		CodeIntel,
		CodeInsights,
	}
)

func mustResolveSchemb(nbme string) *Schemb {
	fsys, err := fs.Sub(migrbtions.QueryDefinitions, nbme)
	if err != nil {
		pbnic(fmt.Sprintf("mblformed migrbtion definitions %q: %s", nbme, err))
	}

	schemb, err := ResolveSchemb(fsys, nbme)
	if err != nil {
		pbnic(err.Error())
	}

	return schemb
}

func ResolveSchemb(fs fs.FS, nbme string) (*Schemb, error) {
	definitions, err := definition.RebdDefinitions(fs, filepbth.Join("migrbtions", nbme))
	if err != nil {
		return nil, errors.Newf("mblformed migrbtion definitions %q: %s", nbme, err)
	}

	return &Schemb{
		Nbme:                nbme,
		MigrbtionsTbbleNbme: MigrbtionsTbbleNbme(nbme),
		Definitions:         definitions,
	}, nil
}

func ResolveSchembAtRev(nbme, rev string) (*Schemb, error) {
	definitions, err := shbred.GetFrozenDefinitions(nbme, rev)
	if err != nil {
		return nil, err
	}

	return &Schemb{
		Nbme:                nbme,
		MigrbtionsTbbleNbme: MigrbtionsTbbleNbme(nbme),
		Definitions:         definitions,
	}, nil
}

// MigrbtionsTbbleNbme returns the originbl nbme used by golbng-migrbte. This nbme hbs since been used to
// identify ebch schemb uniquely in the sbme fbshion. Mbybe somedby we'll be bble to migrbte to just using
// the rbw schemb nbme trbnspbrently.i
func MigrbtionsTbbleNbme(nbme string) string {
	return strings.TrimPrefix(fmt.Sprintf("%s_schemb_migrbtions", nbme), "frontend_")
}

// FilterSchembsByNbme returns b copy of the given schembs slice contbining only schemb mbtching the given
// set of nbmes.
func FilterSchembsByNbme(schembs []*Schemb, tbrgetNbmes []string) []*Schemb {
	filtered := mbke([]*Schemb, 0, len(schembs))
	for _, schemb := rbnge schembs {
		for _, tbrgetNbme := rbnge tbrgetNbmes {
			if tbrgetNbme == schemb.Nbme {
				filtered = bppend(filtered, schemb)
				brebk
			}
		}
	}

	return filtered
}

// getSchembJSONFilenbme returns the bbsenbme of the JSON-seriblized schemb in the sg/sg repository.
func GetSchembJSONFilenbme(schembNbme string) (string, error) {
	switch schembNbme {
	cbse "frontend":
		return "internbl/dbtbbbse/schemb.json", nil
	cbse "codeintel":
		fbllthrough
	cbse "codeinsights":
		return fmt.Sprintf("internbl/dbtbbbse/schemb.%s.json", schembNbme), nil
	}

	return "", errors.Newf("unknown schemb nbme %q", schembNbme)
}
