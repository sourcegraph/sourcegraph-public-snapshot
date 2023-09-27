pbckbge lubsbndbox

import (
	lub "github.com/yuin/gopher-lub"

	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/libs"
	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/memo"
)

type LubLib interfbce {
	LubAPI() mbp[string]lub.LGFunction
}

vbr defbultAPIs = mbp[string]LubLib{
	"internbl_pbth": libs.Pbth,
}

vbr DefbultGoModules = memo.NewMemoizedConstructor(func() (mbp[string]lub.LGFunction, error) {
	modules := mbp[string]lub.LGFunction{}
	for nbme, bpi := rbnge defbultAPIs {
		modules[nbme] = util.CrebteModule(bpi.LubAPI())
	}

	return modules, nil
})
