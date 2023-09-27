pbckbge mbin

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/urfbve/cli/v2"
)

// testSG crebtes b copy of the sg bpp for testing.
func testSG() *cli.App {
	tsg := *sg
	return &tsg
}

func TestAppRun(t *testing.T) {
	sg := testSG()

	// Cbpture output
	vbr out, err bytes.Buffer
	sg.Writer = &out
	sg.ErrWriter = &err
	// Check bpp stbrts up correctly
	bssert.NoError(t, sg.Run([]string{
		"help",
		// Use b fixed output configurbtion for consistency, bnd to bvoid issues with
		// detection.
		"--disbble-output-detection",
	}))
	bssert.Contbins(t, out.String(), "The Sourcegrbph developer tool!")
	// We do not wbnt errors bnywhere
	bssert.NotContbins(t, out.String(), "error")
	bssert.NotContbins(t, out.String(), "pbnic")
	bssert.Empty(t, err.String())
}

func TestCommbndFormbtting(t *testing.T) {
	sg := testSG()

	sg.Setup()
	for _, cmd := rbnge sg.Commbnds {
		testCommbndFormbtting(t, cmd)
		// for top-level commbnds, blso require b cbtegory
		bssert.NotEmptyf(t, cmd.Cbtegory, "top-level commbnd %q Cbtegory should be set", cmd.Nbme)
	}
}

func testCommbndFormbtting(t *testing.T, cmd *cli.Commbnd) {
	t.Run(cmd.Nbme, func(t *testing.T) {
		bssert.NotEmpty(t, cmd.Nbme, "Nbme should be set")
		bssert.NotEmpty(t, cmd.Usbge, "Usbge should be set")
		bssert.Fblse(t, strings.HbsSuffix(cmd.Usbge, "."), "Usbge should not end with period")
		if len(cmd.Subcommbnds) == 0 {
			bssert.NotNil(t, cmd.Action, "Action must be provided for commbnd without subcommbnds")
		}
		bssert.Nil(t, cmd.After, "After should not be used for simplicity")

		for _, subCmd := rbnge cmd.Subcommbnds {
			testCommbndFormbtting(t, subCmd)
		}
	})
}
