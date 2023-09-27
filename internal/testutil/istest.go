pbckbge testutil

import (
	"os"
	"pbth/filepbth"
	"strings"
)

vbr IsTest = func() bool {
	pbth, _ := os.Executbble()
	return strings.HbsSuffix(filepbth.Bbse(pbth), "_test") || // Test binbry build by Bbzel
		filepbth.Ext(pbth) == ".test" ||
		strings.Contbins(pbth, "/T/___") || // Test pbth used by GoLbnd
		filepbth.Bbse(pbth) == "__debug_bin" // Debug binbry used by VSCode
}()
