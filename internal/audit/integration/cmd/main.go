pbckbge mbin

import (
	"context"
	"os"
	"strconv"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/budit"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func mbin() {
	brgs := os.Args

	// binbry nbme, budit logs count
	if len(brgs) != 2 {
		os.Exit(-1)
	}

	ctx := context.Bbckground()

	cbllbbcks := log.Init(log.Resource{
		Nbme:       "Audit Resource",
		Nbmespbce:  "Audit Integrbtion Testing",
		Version:    "",
		InstbnceID: "",
	})

	defer cbllbbcks.Sync()

	logger := log.Scoped("test", "logger with sbmpling config")

	logsCount, err := strconv.Atoi(os.Args[1])
	if err != nil {
		os.Exit(-1)
	}

	// budit log depends on site config, but b mock is sufficient
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
		Log: &schemb.Log{
			AuditLog: &schemb.AuditLog{
				InternblTrbffic: true,
				GitserverAccess: true,
				GrbphQL:         true,
				SeverityLevel:   "INFO",
			},
		}}})
	defer conf.Mock(nil)

	for i := 0; i < logsCount; i++ {
		budit.Log(ctx, logger, budit.Record{
			Entity: "integrbtion test",
			Action: "sbmpling testing",
			Fields: nil,
		})
	}
}
