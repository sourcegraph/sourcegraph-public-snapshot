package background

import (
	"context"
	"time"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCVEDownloader(store store.Store, metrics *Metrics, interval time.Duration) goroutine.BackgroundRoutine {
	cveParser := &CVEParser{
		store:  store,
		logger: logger.Scoped("sentinel.parser", ""),
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.sentinel-cve-downloader", "Periodically syncs GitHub advisory records into Postgres.",
		interval,
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
	)
}

type CVEParser struct {
	store  store.Store
	logger logger.Logger
}

func NewCVEParser() *CVEParser {
	return &CVEParser{
		logger: logger.Scoped("sentinel.parser", ""),
	}
}

func (parser *CVEParser) handle(ctx context.Context) ([]shared.Vulnerability, error) {
	return parser.ReadGitHubAdvisoryDB(ctx, false)
}
