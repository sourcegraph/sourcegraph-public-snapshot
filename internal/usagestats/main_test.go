package usagestats

import (
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
		log15.Root().SetHandler(log15.DiscardHandler())
	} else {
		logtest.Init(m)
	}
	os.Exit(m.Run())
}
