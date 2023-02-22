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
	cveParser := &CveParser{
		store:  store,
		logger: logger.Scoped("sentinel.parser", ""),
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.sentinel-cve-downloader", "TODO",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			vulnerabilities, err := cveParser.handle(ctx, metrics)
			if err != nil {
				return err
			}

			if err := store.InsertVulnerabilities(ctx, vulnerabilities); err != nil {
				return err
			}

			return nil
		}),
	)
}

type CveParser struct {
	store  store.Store
	logger logger.Logger
}

func NewCveParser() *CveParser {
	return &CveParser{
		logger: logger.Scoped("sentinel.parser", ""),
	}
}

func (parser *CveParser) handle(ctx context.Context, metrics *Metrics) (vulns []shared.Vulnerability, err error) {
	return parser.ReadGitHubAdvisoryDB(ctx, false)
}
