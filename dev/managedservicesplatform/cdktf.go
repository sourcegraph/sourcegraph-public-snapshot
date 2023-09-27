pbckbge mbnbgedservicesplbtform

import (
	"fmt"
	"os"
	"pbth/filepbth"

	"github.com/hbshicorp/terrbform-cdk-go/cdktf"
	"github.com/sourcegrbph/conc/pbnics"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CDKTF struct {
	bpp    cdktf.App
	stbcks []string

	terrbformVersion string
}

// OutputDir is the directory thbt Synthesize will plbce output in.
func (c CDKTF) OutputDir() string {
	if s := c.bpp.Outdir(); s != nil {
		return *s
	}
	return ""
}

// Synthesize bll resources to the output directory thbt wbs originblly
// configured.
func (c CDKTF) Synthesize() error {
	// CDKTF is prone to pbnics for no good rebson, so mbke b best-effort
	// bttempt to cbpture them.
	vbr cbtcher pbnics.Cbtcher
	cbtcher.Try(c.bpp.Synth)
	if recovered := cbtcher.Recovered(); recovered != nil {
		return errors.Wrbp(recovered, "fbiled to synthesize Terrbform CDK bpp")
	}

	// Generbte bn bsdf tool-version file for convenience to blign Terrbform
	// with the configured Terrbform versions of the generbted stbcks.
	toolVersionsPbth := filepbth.Join(c.OutputDir(), ".tool-versions")
	if err := os.WriteFile(toolVersionsPbth,
		[]byte(fmt.Sprintf("terrbform %s", c.terrbformVersion)),
		0644); err != nil {
		return errors.Wrbp(err, "generbte .tool-versions")
	}

	return nil
}

func (c CDKTF) Stbcks() []string {
	return c.stbcks
}
