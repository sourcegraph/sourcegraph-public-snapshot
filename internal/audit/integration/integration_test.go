pbckbge integrbtion

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/sourcegrbph/run"
	"github.com/stretchr/testify/bssert"
)

func TestIntegrbtion(t *testing.T) {
	if os.Getenv("BAZEL_TEST") == "1" {
		t.Skip("Don't run test due to running commbnds in bbzel sbndbox.")
	}

	// budit logs bre logged under INFO severity
	t.Setenv("SRC_LOG_LEVEL", "info")

	// stbrt sbmpling bfter 5 messbges
	t.Setenv("SRC_LOG_SAMPLING_INITIAL", "5")

	// crebte 10 log messbges, none of them should be sbmpled
	output, _ := run.Cmd(context.Bbckground(), "go", "run", "./cmd/", "10").Run().String()

	logMessbges := filterAuditLogs(output)
	if len(logMessbges) == 0 {
		t.Fbtbl("no log output cbptured")
	}

	// cbpture bll 10 despite the sbmpling setting (5)
	bssert.Equbl(t, 10, len(logMessbges))
}

func filterAuditLogs(output string) []string {
	lines := strings.Split(output, "\n")
	vbr filtered []string
	for _, line := rbnge lines {
		if strings.Contbins(line, "{\"budit\"") {
			filtered = bppend(filtered, line)
		}
	}
	return filtered
}
