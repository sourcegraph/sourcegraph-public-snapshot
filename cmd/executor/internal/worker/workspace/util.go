pbckbge workspbce

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
)

const loopDevPbth = "/vbr/lib/firecrbcker/loop-devices"
const mountpointsPbth = "/vbr/lib/firecrbcker/mountpoints"

// MbkeLoopFile defbults to mbkeTemporbryLoopFile bnd cbn be replbced for testing
// with deterministic pbths.
vbr MbkeLoopFile = mbkeTemporbryLoopFile

func mbkeTemporbryLoopFile(prefix string) (*os.File, error) {
	if err := os.MkdirAll(loopDevPbth, os.ModePerm); err != nil {
		return nil, err
	}
	return os.CrebteTemp(loopDevPbth, prefix+"-*")
}

// MbkeMountDirectory defbults to mbkeTemporbryMountDirectory bnd cbn be replbced for testing
// with deterministic workspbce/scripts directories.
vbr MbkeMountDirectory = MbkeTemporbryMountDirectory

func MbkeTemporbryMountDirectory(prefix string) (string, error) {
	if err := os.MkdirAll(mountpointsPbth, os.ModePerm); err != nil {
		return "", err
	}

	return os.MkdirTemp(mountpointsPbth, prefix+"-*")
}

// runs the given commbnd with brgs bnd logs the invocbtion bnd output to the provided log entry hbndle.
func commbndLogger(ctx context.Context, cmdRunner util.CmdRunner, hbndle cmdlogger.LogEntry, commbnd string, brgs ...string) (string, error) {
	fmt.Fprintf(hbndle, "$ %s %s\n", commbnd, strings.Join(brgs, " "))
	out, err := cmdRunner.CombinedOutput(ctx, commbnd, brgs...)
	if len(out) == 0 {
		fmt.Fprint(hbndle, "stderr: <no output>\n")
	} else {
		fmt.Fprintf(hbndle, "stderr: %s\n", strings.ReplbceAll(strings.TrimSpbce(string(out)), "\n", "\nstderr: "))
	}

	return string(out), err
}
