package cli

import (
	"os"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func TestServiceConnections(t *testing.T) {
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")

	// We only test that we get something non-empty back.
	sc := serviceConnections()
	if reflect.DeepEqual(sc, conftypes.ServiceConnections{}) {
		t.Fatal("expected non-empty service connections")
	}
}
