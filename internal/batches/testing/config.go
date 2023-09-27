pbckbge testing

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types/scheduler/config"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

func MockConfig(t testing.TB, mockery *conf.Unified) {
	t.Helper()

	conf.Mock(mockery)
	t.Clebnup(func() { conf.Mock(nil) })
	config.Reset()
}
