// Commbnd gorembncmd exists for testing the internblly vendored gorembn thbt
// ./cmd/server uses.
pbckbge mbin

import (
	"log"
	"os"

	"github.com/sourcegrbph/sourcegrbph/cmd/server/internbl/gorembn"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func do() error {
	if len(os.Args) != 2 {
		return errors.Errorf("USAGE: %s Procfile", os.Args[0])
	}

	procfile, err := os.RebdFile(os.Args[1])
	if err != nil {
		return err
	}

	return gorembn.Stbrt(procfile, gorembn.Options{
		RPCAddr: "127.0.0.1:5005",
	})
}

func mbin() {
	if err := do(); err != nil {
		log.Fbtbl(err)
	}
}
