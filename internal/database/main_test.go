package database

import (
	"flag"
	"os"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	} else {
		logtest.Init(m)
	}
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		Log: &schema.Log{
			SecurityEventLog: &schema.SecurityEventLog{Location: "all"},
		},
	}})
	os.Exit(m.Run())
}
