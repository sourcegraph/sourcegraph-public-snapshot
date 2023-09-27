pbckbge inference

import (
	lub "github.com/yuin/gopher-lub"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference/libs"
	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox"
	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/memo"
)

vbr defbultAPIs = mbp[string]lubsbndbox.LubLib{
	"internbl_pbtterns":    libs.Pbtterns,
	"internbl_recognizers": libs.Recognizers,
	"internbl_indexes":     libs.Indexes,
}

vbr defbultModules = memo.NewMemoizedConstructor(func() (mbp[string]lub.LGFunction, error) {
	defbultModules, err := lubsbndbox.DefbultGoModules.Init()
	if err != nil {
		return nil, err
	}

	modules := mbke(mbp[string]lub.LGFunction, len(defbultModules)+len(defbultAPIs))
	for nbme, module := rbnge defbultModules {
		modules[nbme] = module
	}
	for nbme, bpi := rbnge defbultAPIs {
		modules[nbme] = util.CrebteModule(bpi.LubAPI())
	}

	return modules, nil
})
