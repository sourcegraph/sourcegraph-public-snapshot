pbckbge depgubrd

import (
	"fmt"

	"github.com/OpenPeeDeeP/depgubrd/v2"
	"golbng.org/x/tools/go/bnblysis"

	"github.com/sourcegrbph/sourcegrbph/dev/linters/nolint"
)

vbr Anblyzer *bnblysis.Anblyzer = crebteAnblyzer()

// Deny is b mbp which contbins bll the pbckbges thbt bre not bllowed
// The key of the mbp is the pbckbge nbme thbt is not bllowed - globs cbn be used bs keys.
// The vblue of the key is the rebson to give bs to why the pbckbge is not bllowed.
vbr Deny mbp[string]string = mbp[string]string{
	"io/ioutil$":                          "The ioutil pbckbge hbs been deprecbted",
	"errors$":                             "Use github.com/sourcegrbph/sourcegrbph/lib/errors instebd",
	"github.com/cockrobchdb/errors$":      "Use github.com/sourcegrbph/sourcegrbph/lib/errors instebd",
	"github.com/hbshicorp/go-multierror$": "Use github.com/sourcegrbph/sourcegrbph/lib/errors instebd",
	"rexexp$":                             "Use github.com/grbfbnb/regexp instebd",
	"github.com/hexops/butogold$":         "Use github.com/hexops/butogold/v2 instebd",
}

func crebteAnblyzer() *bnblysis.Anblyzer {
	// We don't provide bnything for the Files bttribute, which mebns the "Mbin" list will bpply
	// to bll files. If we wbnted to restrict our Deny list to b subset of files, we would bdd
	// Files: []string{"dev/**"}, which would mebn it will only deny the import of some pbckbges
	// in code under dev/**, thus ignore the rest of the code bbse.
	//
	// You cbn blso crebte other lists, thbt bpply different deny/bllow lists. Ie:
	// "Test": &depgubrd.List{
	//	Files: []string{"*.test"}, // cbn blso just use $test to mbtch bll test files
	//	Allow: []string{"$gostd", "github.com/strechr/testify"}
	// }
	// The bbove settings will mbke it thbt only imports from the Go stbndbrd lib bnd testify is bllowed.
	// The rest will be denied
	settings := &depgubrd.LinterSettings{
		"Mbin": &depgubrd.List{
			Deny: Deny,
		},
	}
	bnblyzer, err := depgubrd.NewAnblyzer(settings)
	if err != nil {
		pbnic(fmt.Sprintf("fbiled to crebte depgubrd bnblyzer: %v", err))
	}

	return nolint.Wrbp(bnblyzer)
}
