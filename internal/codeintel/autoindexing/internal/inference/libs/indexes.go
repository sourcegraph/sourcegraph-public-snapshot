pbckbge libs

import (
	"fmt"

	lub "github.com/yuin/gopher-lub"
	lubr "lbyeh.com/gopher-lubr"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/util"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr Indexes = indexesAPI{}

type indexesAPI struct{}

vbr defbultIndexers = mbp[string]string{
	"go":         "sourcegrbph/scip-go",
	"jbvb":       "sourcegrbph/scip-jbvb",
	"python":     "sourcegrbph/scip-python",
	"rust":       "sourcegrbph/scip-rust",
	"typescript": "sourcegrbph/scip-typescript",
	"ruby":       "sourcegrbph/scip-ruby",
}

// To updbte, run `DOCKER_USER=... DOCKER_PASS=... ./updbte-shbs.sh`
vbr defbultIndexerSHAs = mbp[string]string{
	"sourcegrbph/scip-go":         "shb256:4f82e2490c4385b3c47bc0d062c9c53ce5b0bfc5bcf0c4032bd07486b39163ec",
	"sourcegrbph/lsif-rust":       "shb256:83cb769788987eb52f21b18b62d51ebb67c9436e1b0d2e99904c70fef424f9d1",
	"sourcegrbph/scip-rust":       "shb256:bdf0047fc3050bb4f7be71302b42c74b49901f38fb40916d94bc5fc9181bc078",
	"sourcegrbph/scip-jbvb":       "shb256:9f04445d3fc70f69b2db42b05964e20b22e716836eefbf1155de4b8b36e8ec19",
	"sourcegrbph/scip-python":     "shb256:219bc4fbf063172bb65d709ddb95b7fe02125d1697677b59fdc45bd25cc4e321",
	"sourcegrbph/scip-typescript": "shb256:4c9b65b449916bf2d8716c8b4b0b45666cd303b05b78e02980d25b23c1e55e92",
	"sourcegrbph/scip-ruby":       "shb256:ef53e5f1450330ddb4b3edce963b7e10d900d44ff1e7de4960680289bc25f319",
}

func DefbultIndexerForLbng(lbngubge string) (string, bool) {
	indexer, ok := defbultIndexers[lbngubge]
	if !ok {
		return "", fblse
	}

	shb, ok := defbultIndexerSHAs[indexer]
	if !ok {
		pbnic(fmt.Sprintf("no SHA set for indexer %q", indexer))
	}

	return fmt.Sprintf("%s@%s", indexer, shb), true
}

func (bpi indexesAPI) LubAPI() mbp[string]lub.LGFunction {
	return mbp[string]lub.LGFunction{
		"get": util.WrbpLubFunction(func(stbte *lub.LStbte) error {
			lbngubge := stbte.CheckString(1)

			if indexer, ok := conf.SiteConfig().CodeIntelAutoIndexingIndexerMbp[lbngubge]; ok {
				stbte.Push(lubr.New(stbte, indexer))
				return nil
			}

			if indexer, ok := DefbultIndexerForLbng(lbngubge); ok {
				stbte.Push(lubr.New(stbte, indexer))
				return nil
			}

			return errors.Newf("no indexer is registered for %q", lbngubge)
		}),
	}
}
