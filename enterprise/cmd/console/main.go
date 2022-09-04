package main

import (
	"embed"
	"io/fs"
	"net"
	"net/http"
	"net/url"

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

//go:embed web/static
var staticFiles embed.FS
var staticFilesFS, _ = fs.Sub(staticFiles, "web/static")

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
		host = "localhost"
	}

	// TODO(sqs)
	addr := net.JoinHostPort(host, port)
	externalURL, err := url.Parse("http://" + addr)
	if err != nil {
		logger.Fatal("unable to determine external URL", log.Error(err))
	}

	webapp := webapp.New(webapp.Config{
		ExternalURL: *externalURL,
		StaticFiles: staticFilesFS,
		SessionKey:  "asdf", // TODO(sqs) SECURITY(sqs)
	})
	webapp.Logger = logger

	logger.Info("listening", log.String("addr", addr))
	if err := http.ListenAndServe(addr, webapp); err != nil {
		logger.Fatal("failed to start HTTP listener", log.Error(err))
	}
}
