package oobmigration

import (
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"
)

func TestMain(m *testing.M) {
	logtest.Init(m)
	os.Exit(m.Run())
}
