package updatecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context/ctxhttp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestatsdeprecated"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// recorder records operational metrics for methods.
var recorder = metrics.NewOperationMetrics(prometheus.DefaultRegisterer, "updatecheck", metrics.WithLabels("method"))

// Status of the check for software updates for Sourcegraph.
type Status struct {
	Date          time.Time // the time that the last check completed
	Err           error     // the error that occurred, if any
	UpdateVersion string    // the version string of the updated version, if any
}

// HasUpdate reports whether the status indicates an update is available.
func (s Status) HasUpdate() bool { return s.UpdateVersion != "" }

var (
	mu         sync.Mutex
	startedAt  *time.Time
	lastStatus *Status
)

// Last returns the status of the last-completed software update check.
func Last() *Status {
	mu.Lock()
	defer mu.Unlock()
	if lastStatus == nil {
		return nil
	}
	tmp := *lastStatus
	return &tmp
}

// IsPending returns whether an update check is in progress.
func IsPending() bool {
	mu.Lock()
	defer mu.Unlock()
	return startedAt != nil
}

var baseURL = &url.URL{
	Scheme: "https",
	Host:   "sourcegraph.com",
	Path:   "/.api/updates",
}

// recordOperation returns a record fn that is called on any given return err. If an error is encountered
// it will register the err metric. The err is never altered.
func recordOperation(method string) func(error) error {
	start := time.Now()
	return func(err error) error {
		recorder.Observe(time.Since(start).Seconds(), 1, &err, method)
		return err
	}
}

func getAndMarshalSiteActivityJSON(ctx context.Context, criticalOnly bool) (json.RawMessage, error) {
	rec := recordOperation("getAndMarshalSiteActivityJSON")

	var days, weeks, months int
	if criticalOnly {
		months = 1
	} else {
		days, weeks, months = 2, 1, 1
	}
	siteActivity, err := usagestats.GetSiteUsageStatistics(ctx, &usagestats.SiteUsageStatisticsOptions{
		DayPeriods:   &days,
		WeekPeriods:  &weeks,
		MonthPeriods: &months,
	})
	defer rec(err)

	if err != nil {
		return nil, err
	}
	contents, err := json.Marshal(siteActivity)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(contents), nil
}

func hasSearchOccurred(ctx context.Context) (bool, error) {
	rec := recordOperation("hasSearchOccurred")
	searchOccurred, err := usagestats.HasSearchOccurred()
	return searchOccurred, rec(err)
}

func hasFindRefsOccurred(ctx context.Context) (bool, error) {
	rec := recordOperation("hasSearchOccured")
	findRefsOccurred, err := usagestats.HasFindRefsOccurred()
	return findRefsOccurred, rec(err)
}

func getTotalUsersCount(ctx context.Context) (int, error) {
	rec := recordOperation("getTotalUsersCount")
	totalUsers, err := db.Users.Count(ctx, &db.UsersListOptions{})
	return totalUsers, rec(err)
}

func getTotalReposCount(ctx context.Context) (int, error) {
	rec := recordOperation("getTotalReposCount")
	totalRepos, err := db.Repos.Count(ctx, db.ReposListOptions{})
	return totalRepos, rec(err)
}

func getUsersActiveTodayCount(ctx context.Context) (int, error) {
	rec := recordOperation("getUsersActiveTodayCount")
	count, err := usagestatsdeprecated.GetUsersActiveTodayCount()
	return count, rec(err)
}

func getInitialSiteAdminEmail(ctx context.Context) (string, error) {
	rec := recordOperation("getInitialSiteAdminEmail")
	initAdminEmail, err := db.UserEmails.GetInitialSiteAdminEmail(ctx)
	return initAdminEmail, rec(err)
}

func getAndMarshalCampaignsUsageJSON(ctx context.Context) (json.RawMessage, error) {
	rec := recordOperation("getAndMarshalCampaignsUsageJSON")
	campaignsUsage, err := usagestats.GetCampaignsUsageStatistics(ctx)
	defer rec(err)
	if err != nil {
		return nil, err
	}
	contents, err := json.Marshal(campaignsUsage)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(contents), nil
}

func getAndMarshalCodeIntelUsageJSON(ctx context.Context) (json.RawMessage, error) {
	rec := recordOperation("getAndMarshalCodeIntelUsageJSON")
	days, weeks, months := 2, 1, 1
	codeIntelUsage, err := usagestats.GetCodeIntelUsageStatistics(ctx, &usagestats.CodeIntelUsageStatisticsOptions{
		DayPeriods:            &days,
		WeekPeriods:           &weeks,
		MonthPeriods:          &months,
		IncludeEventCounts:    true,
		IncludeEventLatencies: true,
	})
	defer rec(err)
	if err != nil {
		return nil, err
	}
	contents, err := json.Marshal(codeIntelUsage)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(contents), nil
}

func getAndMarshalSearchUsageJSON(ctx context.Context) (json.RawMessage, error) {
	rec := recordOperation("getAndMarshalSearchUsageJSON")
	days, weeks, months := 2, 1, 1
	searchUsage, err := usagestats.GetSearchUsageStatistics(ctx, &usagestats.SearchUsageStatisticsOptions{
		DayPeriods:         &days,
		WeekPeriods:        &weeks,
		MonthPeriods:       &months,
		IncludeEventCounts: true,
	})
	defer rec(err)
	if err != nil {
		return nil, err
	}
	contents, err := json.Marshal(searchUsage)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(contents), nil
}

func updateURL(ctx context.Context) string {
	return baseURL.String()
}

func updateBody(ctx context.Context) (io.Reader, error) {
	logFunc := log15.Debug
	if envvar.SourcegraphDotComMode() {
		logFunc = log15.Warn
	}

	r := &pingRequest{
		ClientSiteID:        siteid.Get(),
		DeployType:          conf.DeployType(),
		ClientVersionString: version.Version(),
		LicenseKey:          conf.Get().LicenseKey,
		CodeIntelUsage:      []byte("{}"),
		SearchUsage:         []byte("{}"),
		CampaignsUsage:      []byte("{}"),
	}

	totalUsers, err := db.Users.Count(ctx, &db.UsersListOptions{})
	if err != nil {
		logFunc("db.Users.Count failed", "error", err)
	}
	r.TotalUsers = int32(totalUsers)
	r.InitialAdminEmail, err = db.UserEmails.GetInitialSiteAdminEmail(ctx)
	if err != nil {
		logFunc("db.UserEmails.GetInitialSiteAdminEmail failed", "error", err)
	}

	if !conf.Get().DisableNonCriticalTelemetry {
		// TODO(Dan): migrate this to the new usagestats package.
		//
		// For the time being, instances will report daily active users through the legacy package via this argument,
		// as well as using the new package through the `act` argument below. This will allow comparison during the
		// transition.
		count, err := getUsersActiveTodayCount(ctx)
		if err != nil {
			logFunc("updatecheck.getUsersActiveToday failed", "error", err)
		}
		r.UniqueUsers = int32(count)
		totalRepos, err := getTotalReposCount(ctx)
		if err != nil {
			logFunc("updatecheck.getTotalReposCount failed", "error", err)
		}
		r.HasRepos = totalRepos > 0

		r.EverSearched, err = hasSearchOccurred(ctx)
		if err != nil {
			logFunc("updatecheck.hasSearchOccurred failed", "error", err)
		}
		r.EverFindRefs, err = hasFindRefsOccurred(ctx)
		if err != nil {
			logFunc("updatecheck.hasFindRefsOccurred failed", "error", err)
		}
		r.Activity, err = getAndMarshalSiteActivityJSON(ctx, false)
		if err != nil {
			logFunc("updatecheck.getAndMarshalSiteActivityJSON failed", "error", err)
		}
		r.CampaignsUsage, err = getAndMarshalCampaignsUsageJSON(ctx)
		if err != nil {
			logFunc("updatecheck.getAndMarshalCampaignsUsageJSON failed", "error", err)
		}
		r.CodeIntelUsage, err = getAndMarshalCodeIntelUsageJSON(ctx)
		if err != nil {
			logFunc("updatecheck.getAndMarshalCodeIntelUsageJSON failed", "error", err)
		}
		r.SearchUsage, err = getAndMarshalSearchUsageJSON(ctx)
		if err != nil {
			logFunc("updatecheck.getAndMarshalSearchUsageJSON failed", "error", err)
		}
		r.ExternalServices, err = externalServiceKinds(ctx)
		if err != nil {
			logFunc("externalServicesKinds failed", "error", err)
		}

		r.HasExtURL = conf.UsingExternalURL()
		r.BuiltinSignupAllowed = conf.IsBuiltinSignupAllowed()
		r.AuthProviders = authProviderTypes()
	} else {
		r.Activity, err = getAndMarshalSiteActivityJSON(ctx, true)
	}

	contents, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(contents), nil
}

func authProviderTypes() []string {
	ps := conf.Get().AuthProviders
	types := make([]string, len(ps))
	for i, p := range ps {
		types[i] = conf.AuthProviderType(p)
	}
	return types
}

func externalServiceKinds(ctx context.Context) ([]string, error) {
	rec := recordOperation("externalServiceKinds")
	services, err := db.ExternalServices.List(ctx, db.ExternalServicesListOptions{})
	defer rec(err)
	if err != nil {
		return nil, err
	}
	kinds := make([]string, len(services))
	for i, s := range services {
		kinds[i] = s.Kind
	}
	return kinds, nil
}

// check performs an update check. It returns the result and updates the global state
// (returned by Last and IsPending).
func check(ctx context.Context) (*Status, error) {
	doCheck := func() (updateVersion string, err error) {
		body, err := updateBody(ctx)
		if err != nil {
			return "", err
		}
		resp, err := ctxhttp.Post(ctx, nil, updateURL(ctx), "application/json", body)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			var description string
			if body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 30)); err != nil {
				description = err.Error()
			} else if len(body) == 0 {
				description = "(no response body)"
			} else {
				description = strconv.Quote(string(bytes.TrimSpace(body)))
			}
			return "", fmt.Errorf("update endpoint returned HTTP error %d: %s", resp.StatusCode, description)
		}

		if resp.StatusCode == http.StatusNoContent {
			return "", nil // no update available
		}

		var latestBuild build
		if err := json.NewDecoder(resp.Body).Decode(&latestBuild); err != nil {
			return "", err
		}
		return latestBuild.Version.String(), nil
	}

	mu.Lock()
	thisCheckStartedAt := time.Now()
	startedAt = &thisCheckStartedAt
	mu.Unlock()

	updateVersion, err := doCheck()

	mu.Lock()
	if startedAt != nil && !startedAt.After(thisCheckStartedAt) {
		startedAt = nil
	}
	lastStatus = &Status{
		Date:          time.Now(),
		Err:           err,
		UpdateVersion: updateVersion,
	}
	mu.Unlock()

	return lastStatus, err
}

var started bool

// Start starts checking for software updates periodically.
func Start() {
	if started {
		panic("already started")
	}
	started = true

	if channel := conf.UpdateChannel(); channel != "release" {
		return // no update check
	}

	ctx := context.Background()
	const delay = 30 * time.Minute
	for {
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		_, _ = check(ctx) // updates global state on its own, can safely ignore return value
		cancel()

		// Randomize sleep to prevent thundering herds.
		randomDelay := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(delay + randomDelay)
	}
}
