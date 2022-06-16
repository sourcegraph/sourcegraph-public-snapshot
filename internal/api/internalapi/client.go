package internalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var frontendInternal = env.Get("SRC_FRONTEND_INTERNAL", "sourcegraph-frontend-internal", "HTTP address for internal frontend HTTP API.")

type internalClient struct {
	// URL is the root to the internal API frontend server.
	URL string
}

var Client = &internalClient{URL: "http://" + frontendInternal}

var requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_frontend_internal_request_duration_seconds",
	Help:    "Time (in seconds) spent on request.",
	Buckets: prometheus.DefBuckets,
}, []string{"category", "code"})

func (c *internalClient) SettingsGetForSubject(
	ctx context.Context,
	subject api.SettingsSubject,
) (parsed *schema.Settings, settings *api.Settings, err error) {
	err = c.postInternal(ctx, "settings/get-for-subject", subject, &settings)
	if err == nil {
		err = jsonc.Unmarshal(settings.Contents, &parsed)
	}
	return parsed, settings, err
}

var MockOrgsListUsers func(orgID int32) (users []int32, err error)

func (c *internalClient) OrgsListUsers(ctx context.Context, orgID int32) (users []int32, err error) {
	if MockOrgsListUsers != nil {
		return MockOrgsListUsers(orgID)
	}
	err = c.postInternal(ctx, "orgs/list-users", orgID, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c *internalClient) OrgsGetByName(ctx context.Context, orgName string) (orgID *int32, err error) {
	err = c.postInternal(ctx, "orgs/get-by-name", orgName, &orgID)
	if err != nil {
		return nil, err
	}
	return orgID, nil
}

func (c *internalClient) UserEmailsGetEmail(ctx context.Context, userID int32) (email *string, err error) {
	err = c.postInternal(ctx, "user-emails/get-email", userID, &email)
	if err != nil {
		return nil, err
	}
	return email, nil
}

// TODO(slimsag): In the future, once we're no longer using environment
// variables to build ExternalURL, remove this in favor of services just reading it
// directly from the configuration file.
//
// TODO(slimsag): needs cleanup as part of upcoming configuration refactor.
func (c *internalClient) ExternalURL(ctx context.Context) (string, error) {
	var externalURL string
	err := c.postInternal(ctx, "app-url", nil, &externalURL)
	if err != nil {
		return "", err
	}
	return externalURL, nil
}

// TODO(slimsag): needs cleanup as part of upcoming configuration refactor.
func (c *internalClient) SendEmail(ctx context.Context, message txtypes.Message) error {
	return c.postInternal(ctx, "send-email", &message, nil)
}

// MockClientConfiguration mocks (*internalClient).Configuration.
var MockClientConfiguration func() (conftypes.RawUnified, error)

func (c *internalClient) Configuration(ctx context.Context) (conftypes.RawUnified, error) {
	if MockClientConfiguration != nil {
		return MockClientConfiguration()
	}
	var cfg conftypes.RawUnified
	err := c.postInternal(ctx, "configuration", nil, &cfg)
	return cfg, err
}

func (c *internalClient) ReposGetByName(ctx context.Context, repoName api.RepoName) (*api.Repo, error) {
	var repo api.Repo
	err := c.postInternal(ctx, "repos/"+string(repoName), nil, &repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (c *internalClient) PhabricatorRepoCreate(ctx context.Context, repo api.RepoName, callsign, url string) error {
	return c.postInternal(ctx, "phabricator/repo-create", api.PhabricatorRepoCreateRequest{
		RepoName: repo,
		Callsign: callsign,
		URL:      url,
	}, nil)
}

var MockExternalServiceConfigs func(kind string, result any) error

// ExternalServiceConfigs fetches external service configs of a single kind into the result parameter,
// which should be a slice of the expected config type.
func (c *internalClient) ExternalServiceConfigs(ctx context.Context, kind string, result any) error {
	if MockExternalServiceConfigs != nil {
		return MockExternalServiceConfigs(kind, result)
	}
	return c.postInternal(ctx, "external-services/configs", api.ExternalServiceConfigsRequest{
		Kind: kind,
	}, &result)
}

// ExternalServicesList returns all external services of the given kind.
func (c *internalClient) ExternalServicesList(
	ctx context.Context,
	opts api.ExternalServicesListRequest,
) ([]*api.ExternalService, error) {
	var extsvcs []*api.ExternalService
	return extsvcs, c.postInternal(ctx, "external-services/list", &opts, &extsvcs)
}

func (c *internalClient) LogTelemetry(ctx context.Context, reqBody any) error {
	return c.postInternal(ctx, "telemetry", reqBody, nil)
}

// postInternal sends an HTTP post request to the internal route.
func (c *internalClient) postInternal(ctx context.Context, route string, reqBody, respBody any) error {
	return c.meteredPost(ctx, "/.internal/"+route, reqBody, respBody)
}

func (c *internalClient) meteredPost(ctx context.Context, route string, reqBody, respBody any) error {
	start := time.Now()
	statusCode, err := c.post(ctx, route, reqBody, respBody)
	d := time.Since(start)

	code := strconv.Itoa(statusCode)
	if err != nil {
		code = "error"
	}
	requestDuration.WithLabelValues(route, code).Observe(d.Seconds())
	return err
}

// post sends an HTTP post request to the provided route. If reqBody is
// non-nil it will Marshal it as JSON and set that as the Request body. If
// respBody is non-nil the response body will be JSON unmarshalled to resp.
func (c *internalClient) post(ctx context.Context, route string, reqBody, respBody any) (int, error) {
	var data []byte
	if reqBody != nil {
		var err error
		data, err = json.Marshal(reqBody)
		if err != nil {
			return -1, err
		}
	}

	req, err := http.NewRequest("POST", c.URL+route, bytes.NewBuffer(data))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if err := checkAPIResponse(resp); err != nil {
		return resp.StatusCode, err
	}

	if respBody != nil {
		return resp.StatusCode, json.NewDecoder(resp.Body).Decode(respBody)
	}
	return resp.StatusCode, nil
}

func checkAPIResponse(resp *http.Response) error {
	if 200 > resp.StatusCode || resp.StatusCode > 299 {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		b := buf.Bytes()
		errString := string(b)
		if errString != "" {
			return errors.Errorf(
				"internal API response error code %d: %s (%s)",
				resp.StatusCode,
				errString,
				resp.Request.URL,
			)
		}
		return errors.Errorf("internal API response error code %d (%s)", resp.StatusCode, resp.Request.URL)
	}
	return nil
}
