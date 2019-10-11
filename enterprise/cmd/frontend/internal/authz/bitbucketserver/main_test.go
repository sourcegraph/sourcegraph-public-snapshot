package bitbucketserver

import (
	"flag"
	"os"
	"testing"

	"gopkg.in/inconshreveable/log15.v2"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}
