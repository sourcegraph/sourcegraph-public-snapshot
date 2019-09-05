package graphqlbackend

import (
	"flag"
	"os"
	"testing"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
