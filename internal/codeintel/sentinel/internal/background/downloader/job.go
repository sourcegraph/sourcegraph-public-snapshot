package downloader

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewCVEDownloader(store store.Store, observationCtx *observation.Context, config *Config) goroutine.BackgroundRoutine {
	cveParser := &CVEParser{
		store:  store,
		logger: log.Scoped("sentinel.parser"),
	}
	metrics := newMetrics(observationCtx)

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(context.Background()),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			vulnerabilities, err := cveParser.handle(ctx)
			if err != nil {
				return err
			}

			numVulnerabilitiesInserted, err := store.InsertVulnerabilities(ctx, vulnerabilities)
			if err != nil {
				return err
			}

			metrics.numVulnerabilitiesInserted.Add(float64(numVulnerabilitiesInserted))
			return nil
		}),
		goroutine.WithName("codeintel.sentinel-cve-downloader"),
		goroutine.WithDescription("Periodically syncs GitHub advisory records into Postgres."),
		goroutine.WithInterval(config.DownloaderInterval),
	)
}

type CVEParser struct {
	store  store.Store
	logger log.Logger
}

func NewCVEParser() *CVEParser {
	return &CVEParser{
		logger: log.Scoped("sentinel.parser"),
	}
}

func (parser *CVEParser) handle(ctx context.Context) ([]shared.Vulnerability, error) {
	return parser.ReadGitHubAdvisoryDB(ctx, false)
}
