package testing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/batches/types/scheduler/config"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func MockConfig(t testing.TB, mockery *conf.Unified) {
	t.Helper()

	conf.Mock(mockery)
	t.Cleanup(func() { conf.Mock(nil) })
	config.Reset()
}
