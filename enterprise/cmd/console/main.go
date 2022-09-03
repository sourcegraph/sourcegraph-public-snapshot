package main

import (
	"net"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/console/internal/webapp"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

const port = "3189"

func main() {
	env.Lock()
	env.HandleHelpFlag()

	liblog := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer liblog.Sync()

	trace.Init()
	profiler.Init()

	logger := log.Scoped("console", "")

	dsns, err := postgresdsn.DSNsBySchema([]string{"console"})
	if err != nil {
		logger.Fatal("failed to get PostgreSQL DSN", log.Error(err))
	}

	sqlDB, err := connections.EnsureNewConsoleDB(dsns["console"], "console", &observation.TestContext)
	if err != nil {
		logger.Fatal("failed to initialize database store", log.Error(err))
	}
	db := database.NewDB(logger, sqlDB)
	_ = db // TODO(sqs)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	webapp := webapp.New(webapp.Config{
		SessionKey: "asdf", // TODO(sqs) SECURITY(sqs)
	})
	webapp.Logger = logger

	addr := net.JoinHostPort(host, port)
	logger.Info("listening", log.String("addr", addr))
	if err := http.ListenAndServe(addr, webapp); err != nil {
		logger.Fatal("failed to start HTTP listener", log.Error(err))
	}
}
