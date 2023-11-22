package azuredevops

import (
	"flag"
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logtest.Init(m)
	os.Exit(m.Run())
}
