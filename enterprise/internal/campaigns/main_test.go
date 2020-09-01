package campaigns

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/inconshreveable/log15"

	secretsPkg "github.com/sourcegraph/sourcegraph/internal/secrets"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	err := secretsPkg.Init()
	if err != nil {
		fmt.Println("Failed to init secrets package:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
