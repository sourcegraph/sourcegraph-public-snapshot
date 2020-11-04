package client

import (
	"compress/gzip"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/efritz/glock"
	"github.com/inconshreveable/log15"
	"github.com/neelance/parallel"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/postgres"
	uploadstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/upload_store"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"golang.org/x/net/context/ctxhttp"
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

	// SendUpload transfers a raw LSIF upload to the bundle manager to be stored on disk. This method returns the
	// size of the file on disk.
	SendUpload(ctx context.Context, bundleID int, r io.Reader) (int64, error)

	// SendUploadPart transfers a partial LSIF upload to the bundle manager to be stored on disk.
	SendUploadPart(ctx context.Context, bundleID, partIndex int, r io.Reader) error

	// StitchParts instructs the bundle manager to collapse multipart uploads into a single file. This method
	// returns the size of the stitched file on disk.
	StitchParts(ctx context.Context, bundleID, numParts int) (int64, error)

	// DeleteUpload removes the upload file with the given identifier from disk.
	DeleteUpload(ctx context.Context, bundleID int) error

	// GetUpload retrieves a reader containing the content of a raw, uncompressed LSIF upload
	// from the bundle manager.
	GetUpload(ctx context.Context, bundleID int) (io.ReadCloser, error)
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

// SendUpload transfers a raw LSIF upload to the bundle manager to be stored on disk. This method returns the
// size of the file on disk.
func (c *bundleManagerClientImpl) SendUpload(ctx context.Context, bundleID int, r io.Reader) (int64, error) {
	return c.uploadStore.Upload(ctx, uploadName(bundleID), r)
}

// SendUploadPart transfers a partial LSIF upload to the bundle manager to be stored on disk.
func (c *bundleManagerClientImpl) SendUploadPart(ctx context.Context, bundleID, partIndex int, r io.Reader) error {
	_, err := c.uploadStore.Upload(ctx, uploadPartName(bundleID, partIndex), r)
	return err
}

// StitchParts instructs the bundle manager to collapse multipart uploads into a single file. This method
// returns the size of the stitched file on disk.
func (c *bundleManagerClientImpl) StitchParts(ctx context.Context, bundleID, numParts int) (int64, error) {
	var sources []string
	for partNumber := 0; partNumber < numParts; partNumber++ {
		sources = append(sources, uploadPartName(bundleID, partNumber))
	}

	return c.uploadStore.Compose(ctx, uploadName(bundleID), sources...)
}

// DeleteUpload removes the upload file with the given identifier from disk.
func (c *bundleManagerClientImpl) DeleteUpload(ctx context.Context, bundleID int) error {
	return c.uploadStore.Delete(ctx, uploadName(bundleID))
}

// GetUpload retrieves a reader containing the content of a raw, uncompressed LSIF upload
// from the bundle manager.
func (c *bundleManagerClientImpl) GetUpload(ctx context.Context, bundleID int) (io.ReadCloser, error) {
	url, err := url.Parse(fmt.Sprintf("%s/uploads/%d", c.bundleManagerURL, bundleID))
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		seek := int64(0)
		zeroPayloadIterations := 0

		for {
			n, err := c.getUploadChunk(ctx, pw, bundleID, url, seek)
			if err != nil {
				if !isConnectionError(err) {
					_ = pw.CloseWithError(err)
					return
				}

				if n == 0 {
					zeroPayloadIterations++

					// Ensure that we don't spin infinitely when when a reset error
					// happens at the beginning of the requested payload. We'll just
					// give up if this happens a few times in a row, which should be
					// very unlikely.
					if zeroPayloadIterations > MaxZeroPayloadIterations {
						_ = pw.CloseWithError(ErrNoDownloadProgress)
						return
					}
				} else {
					zeroPayloadIterations = 0
				}

				// We have a transient error. Make another request but skip the
				// first seek + n bytes as we've already written these to disk.
				seek += n
				log15.Warn("Transient error while reading payload", "error", err)
				continue
			}

			return
		}
	}()

	return gzip.NewReader(pr)
}

// getUploadChunk retrieves a raw LSIF upload from the bundle manager starting from the offset as
// indicated by seek. The number of bytes written to the given writer is returned, along with any
// error.
func (c *bundleManagerClientImpl) getUploadChunk(ctx context.Context, w io.Writer, bundleID int, url *url.URL, seek int64) (int64, error) {
	q := url.Query()
	q.Set("seek", strconv.FormatInt(seek, 10))
	url.RawQuery = q.Encode()

	body, err := c.do(ctx, url)
	if err != nil {
		if err == ErrNotFound {
			body, err := c.uploadStore.Get(context.Background(), uploadName(bundleID), seek)
			if err != nil {
				return 0, err
			}
			defer body.Close()

			return c.ioCopy(w, body)
		}

		return 0, err
	}
	defer body.Close()

	return c.ioCopy(w, body)
}

// do performs an HTTP request to the bundle manager and returns the body content as a reader.
func (c *bundleManagerClientImpl) do(ctx context.Context, url *url.URL) (_ io.ReadCloser, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "BundleManagerClient.do")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req = req.WithContext(ctx)

	if c.httpLimiter != nil {
		span.LogKV("event", "Waiting on HTTP limiter")
		c.httpLimiter.Acquire()
		defer c.httpLimiter.Release()
		span.LogKV("event", "Acquired HTTP limiter")
	}

	req, ht := nethttp.TraceRequest(
		span.Tracer(),
		req,
		nethttp.OperationName("Precise Code Intel Bundle Manager Client"),
		nethttp.ClientTrace(false),
	)
	defer ht.Finish()

	resp, err := ctxhttp.Do(req.Context(), c.httpClient, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func isConnectionError(err error) bool {
	if err != nil && strings.Contains(err.Error(), "read: connection reset by peer") {
		return true
	}

	return false
}

func uploadName(bundleID int) string {
	return fmt.Sprintf("upload-%d.lsif.gz", bundleID)
}

func uploadPartName(bundleID, partNumber int) string {
	return fmt.Sprintf("upload-%d.%d.lsif.gz", bundleID, partNumber)
}
