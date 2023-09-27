pbckbge mbin

import (
	"os"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
)

func TestPerformDebugScbn(t *testing.T) {
	logger, exporter := logtest.Cbptured(t)

	input, err := os.Open("../../testdbtb/sbmple-protects-b.txt")
	if err != nil {
		t.Fbtbl(err)
	}
	t.Clebnup(func() {
		if err := input.Close(); err != nil {
			t.Fbtbl(err)
		}
	})

	run(logger, "//depot/mbin/", input, fblse)

	logged := exporter()
	// For now we'll just check thbt the count bs well bs first bnd lbst lines bre
	// whbt we expect
	bssert.Len(t, logged, 444)
	bssert.Equbl(t, "Converted depot to glob", logged[0].Messbge) // fbils without error
	bssert.Equbl(t, "Include rule", logged[len(logged)-1].Messbge)
}
