package client

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/efritz/glock"
	"github.com/neelance/parallel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/postgres"
	uploadstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/upload_store"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// ErrNotFound occurs when the requested upload or bundle was evicted from disk.
var ErrNotFound = errors.New("data does not exist")

// ErrNoDownloadProgress occurs when there are multiple transient errors in a row that prevent
// the client from receiving a raw upload payload from the bundle manager.
var ErrNoDownloadProgress = errors.New("no download progress")

// The maximum number of iterations where we make no progress while fetching an upload.
const MaxZeroPayloadIterations = 3

// BundleManagerClient is the interface to the precise-code-intel-bundle-manager service.
type BundleManagerClient interface {
	// BundleClient creates a client that can answer intelligence queries for a single dump.
	BundleClient(bundleID int) BundleClient
}

var requestMeter = metrics.NewRequestMeter("precise_code_intel_bundle_manager", "Total number of requests sent to precise code intel bundle manager.")

const MaxIdleConnectionsPerHost = 500

// defaultTransport is the default transport used in the default client and the
// default reverse proxy. ot.Transport will propagate opentracing spans.
var defaultTransport = &ot.Transport{
	RoundTripper: requestMeter.Transport(&http.Transport{
		// Default is 2, but we can send many concurrent requests
		MaxIdleConnsPerHost: MaxIdleConnectionsPerHost,
	}, func(u *url.URL) string {
		// Extract the operation from a path like `/dbs/{id}/{operation}`
		if segments := strings.Split(u.Path, "/"); len(segments) == 4 {
			return segments[3]
		}

		// All other methods are uploading/downloading upload or bundle files.
		// These can all go into a single bucket which  are meant not necessarily
		// meant to have low latency.
		return "transfer"
	}),
}

type bundleManagerClientImpl struct {
	codeIntelDB         *sql.DB
	observationContext  *observation.Context
	httpClient          *http.Client
	httpLimiter         *parallel.Run
	uploadStore         uploadstore.Store
	bundleManagerURL    string
	userAgent           string
	maxPayloadSizeBytes int
	clock               glock.Clock
	ioCopy              func(io.Writer, io.Reader) (int64, error)
}

var _ BundleManagerClient = &bundleManagerClientImpl{}

func New(
	codeIntelDB *sql.DB,
	observationContext *observation.Context,
	bundleManagerURL string,
	uploadStore uploadstore.Store,
) BundleManagerClient {
	return &bundleManagerClientImpl{
		codeIntelDB:         codeIntelDB,
		observationContext:  observationContext,
		httpClient:          &http.Client{Transport: defaultTransport},
		httpLimiter:         parallel.NewRun(500),
		uploadStore:         uploadStore,
		bundleManagerURL:    bundleManagerURL,
		userAgent:           filepath.Base(os.Args[0]),
		maxPayloadSizeBytes: 100 * 1000 * 1000, // 100Mb
		clock:               glock.NewRealClock(),
		ioCopy:              io.Copy,
	}
}

// BundleClient creates a client that can answer intelligence queries for a single dump.
func (c *bundleManagerClientImpl) BundleClient(bundleID int) BundleClient {
	return &bundleClientImpl{
		bundleID: bundleID,
		store:    persistence.NewObserved(postgres.NewStore(c.codeIntelDB, bundleID), c.observationContext),
		databaseOpener: func(ctx context.Context, filename string, store persistence.Store) (database.Database, error) {
			db, err := database.OpenDatabase(ctx, filename, store)
			if err != nil {
				return nil, err
			}

			return database.NewObserved(db, filename, c.observationContext), nil
		},
	}
}
