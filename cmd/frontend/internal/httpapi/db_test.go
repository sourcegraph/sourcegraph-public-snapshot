package httpapi

import (
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func TestMain(m *testing.M) {
	flag.Parse()
	dbtesting.DBNameSuffix = "httpapidb"
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}
