package campaigns

import (
	"flag"
	"fmt"
	"os"
	"testing"

	secretsPkg "github.com/sourcegraph/sourcegraph/internal/secrets"

	"github.com/inconshreveable/log15"
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
