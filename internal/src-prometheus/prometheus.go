// Package srcprometheus defines an API to interact with Sourcegraph Prometheus, including
// prom-wrapper. See https://docs.sourcegraph.com/dev/background-information/observability/prometheus
package srcprometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

// ErrPrometheusUnavailable is raised specifically when prometheusURL is unset or when
// prometheus API access times out, both of which indicate that the server API has likely
// been configured to explicitly disallow access to prometheus, or that prometheus is not
// deployed at all. The website checks for this error in `fetchMonitoringStats`, for example.
var ErrPrometheusUnavailable = errors.New("prometheus API is unavailable")

// PrometheusURL is the configured Prometheus instance.
var PrometheusURL = env.Get("PROMETHEUS_URL", "", "prometheus server URL")

type Client interface {
	GetAlertsStatus(ctx context.Context) (*AlertsStatus, error)
	GetAlertsHistory(ctx context.Context, timespan time.Duration) (*AlertsHistory, error)
	GetConfigStatus(ctx context.Context) (*ConfigStatus, error)
}

type client struct {
	http    http.Client
	promURL url.URL
}

// NewClient provides a client for interacting with Sourcegraph Prometheus. It errors if
// the target Prometheus URL is invalid, or if no Prometheus URL is configured at all.
// Users should check for the latter case by asserting against `ErrPrometheusUnavailable`
// to avoid rendering an error.
//
// See https://docs.sourcegraph.com/dev/background-information/observability/prometheus
func NewClient(prometheusURL string) (Client, error) {
	if prometheusURL == "" {
		return nil, ErrPrometheusUnavailable
	}
	promURL, err := url.Parse(prometheusURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	return &client{
		http: http.Client{
			Transport: &roundTripper{},
		},
		promURL: *promURL,
	}, nil
}

func (c *client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	return resp, nil
}

const PathPrefixAlertsStatus = "/prom-wrapper/alerts-status"

func (c *client) GetAlertsStatus(ctx context.Context) (*AlertsStatus, error) {
	requestURL := c.promURL
	requestURL.Path = PathPrefixAlertsStatus
	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("prometheus misconfigured: %w", err)
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("alerts status: %w", err)
	}

	var alertsStatus AlertsStatus
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&alertsStatus); err != nil {
		return nil, err
	}
	return &alertsStatus, nil
}

func (c *client) GetAlertsHistory(ctx context.Context, timespan time.Duration) (*AlertsHistory, error) {
	requestURL := c.promURL
	requestURL.Path = PathPrefixAlertsStatus + "/history"
	query := make(url.Values)
	query.Add("timespan", timespan.String())
	requestURL.RawQuery = query.Encode()
	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("prometheus misconfigured: %w", err)
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("alerts history: %w", err)
	}

	var alertsHistory AlertsHistory
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&alertsHistory); err != nil {
		return nil, err
	}
	return &alertsHistory, nil
}

const PathPrefixConfigSubscriber = "/prom-wrapper/config-subscriber"

func (c *client) GetConfigStatus(ctx context.Context) (*ConfigStatus, error) {
	requestURL := c.promURL
	requestURL.Path = PathPrefixConfigSubscriber
	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("prometheus misconfigured: %w", err)
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("config status: %w", err)
	}

	var status ConfigStatus
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return &status, nil
}

// roundTripper treats certain connection errors as `ErrPrometheusUnavailable` which can be
// handled explicitly for environments without Prometheus available.
type roundTripper struct{}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(req)

	// there isn't a great way to check for conn refused, sadly https://github.com/golang/go/issues/9424
	// so check for specific syscall errors to detect if the provided prometheus server is
	// not accessible in this deployment. we also treat deadline exceeds as an indicator.
	var syscallErr *os.SyscallError
	if errors.As(err, &syscallErr) {
		if syscallErr.Err == syscall.ECONNREFUSED || syscallErr.Err == syscall.EHOSTUNREACH {
			err = ErrPrometheusUnavailable
		}
	} else if errors.Is(err, context.DeadlineExceeded) {
		err = ErrPrometheusUnavailable
	}

	return resp, err
}
