package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/mxk/go-flowrate/flowrate"
	"github.com/neelance/parallel"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/codeintelutils"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"golang.org/x/net/context/ctxhttp"
)

// ErrNotFound occurs when the requested upload or bundle was evicted from disk.
var ErrNotFound = errors.New("data does not exist")

// BundleManagerClient is the interface to the precise-code-intel-bundle-manager service.
type BundleManagerClient interface {
	// BundleClient creates a client that can answer intelligence queries for a single dump.
	BundleClient(bundleID int) BundleClient

	// SendUpload transfers a raw LSIF upload to the bundle manager to be stored on disk.
	SendUpload(ctx context.Context, bundleID int, r io.Reader) error

	// SendUploadPart transfers a partial LSIF upload to the bundle manager to be stored on disk.
	SendUploadPart(ctx context.Context, bundleID, partIndex int, r io.Reader) error

	// StitchParts instructs the bundle manager to collapse multipart uploads into a single file.
	StitchParts(ctx context.Context, bundleID int) error

	// DeleteUpload removes the upload file with the given identifier from disk.
	DeleteUpload(ctx context.Context, bundleID int) error

	// GetUploads retrieves a raw LSIF upload from disk. The file is written to a file in the
	// given directory with a random filename. The generated filename is returned on success.
	GetUpload(ctx context.Context, bundleID int, dir string) (string, error)

	// SendDB transfers a converted database to the bundle manager to be stored on disk. This
	// will also remove the original upload file with the same identifier from disk.
	SendDB(ctx context.Context, bundleID int, filename string) error

	// Exists determines if a file exists on disk for all the supplied identifiers.
	Exists(ctx context.Context, bundleIDs []int) (map[int]bool, error)
}

type baseClient interface {
	QueryBundle(ctx context.Context, bundleID int, op string, qs map[string]interface{}, target interface{}) error
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
	httpClient       *http.Client
	httpLimiter      *parallel.Run
	bundleManagerURL string
	UserAgent        string
}

var _ BundleManagerClient = &bundleManagerClientImpl{}
var _ baseClient = &bundleManagerClientImpl{}

func New(bundleManagerURL string) BundleManagerClient {
	return &bundleManagerClientImpl{
		httpClient:       &http.Client{Transport: defaultTransport},
		httpLimiter:      parallel.NewRun(500),
		bundleManagerURL: bundleManagerURL,
		UserAgent:        filepath.Base(os.Args[0]),
	}
}

// BundleClient creates a client that can answer intelligence queries for a single dump.
func (c *bundleManagerClientImpl) BundleClient(bundleID int) BundleClient {
	return &bundleClientImpl{
		base:     c,
		bundleID: bundleID,
	}
}

// SendUpload transfers a raw LSIF upload to the bundle manager to be stored on disk.
func (c *bundleManagerClientImpl) SendUpload(ctx context.Context, bundleID int, r io.Reader) error {
	url, err := makeURL(c.bundleManagerURL, fmt.Sprintf("uploads/%d", bundleID), nil)
	if err != nil {
		return err
	}

	body, err := c.do(ctx, "POST", url, r)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// SendUploadPart transfers a partial LSIF upload to the bundle manager to be stored on disk.
func (c *bundleManagerClientImpl) SendUploadPart(ctx context.Context, bundleID, partIndex int, r io.Reader) error {
	url, err := makeURL(c.bundleManagerURL, fmt.Sprintf("uploads/%d/%d", bundleID, partIndex), nil)
	if err != nil {
		return err
	}

	body, err := c.do(ctx, "POST", url, r)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// StitchParts instructs the bundle manager to collapse multipart uploads into a single file.
func (c *bundleManagerClientImpl) StitchParts(ctx context.Context, bundleID int) error {
	url, err := makeURL(c.bundleManagerURL, fmt.Sprintf("uploads/%d/stitch", bundleID), nil)
	if err != nil {
		return err
	}

	body, err := c.do(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// DeleteUpload removes the upload file with the given identifier from disk.
func (c *bundleManagerClientImpl) DeleteUpload(ctx context.Context, bundleID int) error {
	url, err := makeURL(c.bundleManagerURL, fmt.Sprintf("uploads/%d", bundleID), nil)
	if err != nil {
		return err
	}

	body, err := c.do(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// GetUploads retrieves a raw LSIF upload from disk. The file is written to a file in the
// given directory with a random filename. The generated filename is returned on success.
func (c *bundleManagerClientImpl) GetUpload(ctx context.Context, bundleID int, dir string) (_ string, err error) {
	url, err := makeURL(c.bundleManagerURL, fmt.Sprintf("uploads/%d", bundleID), nil)
	if err != nil {
		return "", err
	}

	f, err := openRandomFile(dir)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	totalWritten := int64(0)
	for {
		q := url.Query()
		q.Set("seek", fmt.Sprintf("%d", totalWritten))
		url.RawQuery = q.Encode()

		body, err := c.do(ctx, "GET", url, nil)
		if err != nil {
			return "", err
		}
		defer body.Close()

		//
		// TODO - need to test these conditions specifically now
		//

		if n, err := io.Copy(f, body); err != nil {
			if isConnectionError(err) {
				log15.Error("Failure to download bundle from manager - error occurred on read", "n", n)
				totalWritten += n
				continue
			}

			return "", err
		}

		return f.Name(), nil
	}
}

// TODO(efritz) - envvar / configure
const MaxPayloadSizeBytes = 100 * 1000 * 1000

// SendDB transfers a converted database to the bundle manager to be stored on disk.
func (c *bundleManagerClientImpl) SendDB(ctx context.Context, bundleID int, filename string) error {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if fileInfo.Size() <= int64(MaxPayloadSizeBytes) {
		return c.sendDB(ctx, bundleID, filename)
	}

	return c.sendDBMultipart(ctx, bundleID, filename)
}

// TODO(efritz) - document
func (c *bundleManagerClientImpl) sendDB(ctx context.Context, bundleID int, filename string) error {
	url, err := makeURL(c.bundleManagerURL, fmt.Sprintf("dbs/%d", bundleID), nil)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	body, err := c.do(ctx, "POST", url, file)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// TODO(efritz) - document
func (c *bundleManagerClientImpl) sendDBMultipart(ctx context.Context, bundleID int, filename string) error {
	files, cleanup, err := codeintelutils.SplitFile(filename, MaxPayloadSizeBytes)
	if err != nil {
		return err
	}
	defer func() {
		err = cleanup(err)
	}()

	for i, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				err = multierror.Append(err, closeErr)
			}
		}()

		url, err := makeURL(c.bundleManagerURL, fmt.Sprintf("dbs/%d/%d", bundleID, i), nil)
		if err != nil {
			return err
		}

		// TODO(efritz) - will also need a way to clean up old dump files
		body, err := c.do(ctx, "POST", url, codeintelutils.Gzip(f))
		if err != nil {
			return err
		}
		body.Close()
	}

	url, err := makeURL(c.bundleManagerURL, fmt.Sprintf("dbs/%d/stitch", bundleID), nil)
	if err != nil {
		return err
	}

	body, err := c.do(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// Exists determines if a file exists on disk for all the supplied identifiers.
func (c *bundleManagerClientImpl) Exists(ctx context.Context, bundleIDs []int) (target map[int]bool, err error) {
	var bundleIDStrings []string
	for _, bundleID := range bundleIDs {
		bundleIDStrings = append(bundleIDStrings, fmt.Sprintf("%d", bundleID))
	}

	url, err := makeURL(c.bundleManagerURL, "exists", map[string]interface{}{
		"ids": strings.Join(bundleIDStrings, ","),
	})
	if err != nil {
		return nil, err
	}

	body, err := c.do(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if err := json.NewDecoder(body).Decode(&target); err != nil {
		return nil, err
	}

	return target, nil
}

func (c *bundleManagerClientImpl) QueryBundle(ctx context.Context, bundleID int, op string, qs map[string]interface{}, target interface{}) (err error) {
	url, err := makeBundleURL(c.bundleManagerURL, bundleID, op, qs)
	if err != nil {
		return err
	}

	body, err := c.do(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	defer body.Close()

	return json.NewDecoder(body).Decode(&target)
}

func (c *bundleManagerClientImpl) do(ctx context.Context, method string, url *url.URL, body io.Reader) (_ io.ReadCloser, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "BundleManagerClient.do")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	req, err := http.NewRequest(method, url.String(), limitTransferRate(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
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

func openRandomFile(dir string) (*os.File, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return os.Create(filepath.Join(dir, uuid.String()))
}

func makeURL(baseURL, path string, qs map[string]interface{}) (*url.URL, error) {
	values := url.Values{}
	for k, v := range qs {
		values[k] = []string{fmt.Sprintf("%v", v)}
	}

	url, err := url.Parse(fmt.Sprintf("%s/%s", baseURL, path))
	if err != nil {
		return nil, err
	}
	url.RawQuery = values.Encode()
	return url, nil
}

func makeBundleURL(baseURL string, bundleID int, op string, qs map[string]interface{}) (*url.URL, error) {
	return makeURL(baseURL, fmt.Sprintf("dbs/%d/%s", bundleID, op), qs)
}

// limitTransferRate applies a transfer limit to the given reader.
//
// In the case that the bundle manager is running on the same host as this service, an unbounded
// transfer rate can end up being so fast that we harm our own network connectivity. In order to
// prevent the disruption of other in-flight requests, we cap the transfer rate of r to 1Gbps.
func limitTransferRate(r io.Reader) io.ReadCloser {
	if r == nil {
		return nil
	}

	return flowrate.NewReader(r, 1000*1000*1000)
}

//
// Temporary network debugging code

var connectionErrors = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "src_bundle_manager_connection_reset_by_peer_read",
	Help: "The total number connection reset by peer errors (client) when trying to transfer upload payloads.",
})

func init() {
	prometheus.MustRegister(connectionErrors)
}

func isConnectionError(err error) bool {
	if err != nil && strings.Contains(err.Error(), "read: connection reset by peer") {
		connectionErrors.Inc()
		return true
	}

	return false
}
