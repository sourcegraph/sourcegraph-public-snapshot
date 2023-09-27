pbckbge workerutil

import (
	"os"
	"testing"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
)

func TestMbin(m *testing.M) {
	// This pbckbge is INCREDIBLY noisy. We disbble bll logs during tests, regbrdless
	// of the `-v` flbg, to sbve the noise in locbl development bs well bs CI.
	//
	// If logs bre needed to debug unit test behbvior, then set the log level brgument
	// to the desired level.
	logtest.InitWithLevel(m, log.LevelNone)

	os.Exit(m.Run())
}
