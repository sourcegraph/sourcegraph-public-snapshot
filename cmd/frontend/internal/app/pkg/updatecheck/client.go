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

	"golang.org/x/net/context/ctxhttp"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestatsdeprecated"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

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

func getAndMarshalSiteActivityJSON(ctx context.Context) (json.RawMessage, error) {
	days, weeks, months := 2, 1, 1
	siteActivity, err := usagestats.GetSiteUsageStatistics(ctx, &usagestats.SiteUsageStatisticsOptions{
		DayPeriods:   &days,
		WeekPeriods:  &weeks,
		MonthPeriods: &months,
	})
	if err != nil {
		return nil, err
	}
	contents, err := json.Marshal(siteActivity)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(contents), nil
}

func getAndMarshalCampaignsUsageJSON(ctx context.Context) (json.RawMessage, error) {
	campaignsUsage, err := usagestats.GetCampaignsUsageStatistics(ctx)
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
	days, weeks, months := 2, 1, 1
	codeIntelUsage, err := usagestats.GetCodeIntelUsageStatistics(ctx, &usagestats.CodeIntelUsageStatisticsOptions{
		DayPeriods:            &days,
		WeekPeriods:           &weeks,
		MonthPeriods:          &months,
		IncludeEventCounts:    !conf.Get().DisableNonCriticalTelemetry,
		IncludeEventLatencies: true,
	})
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
	days, weeks, months := 2, 1, 1
	searchUsage, err := usagestats.GetSearchUsageStatistics(ctx, &usagestats.SearchUsageStatisticsOptions{
		DayPeriods:         &days,
		WeekPeriods:        &weeks,
		MonthPeriods:       &months,
		IncludeEventCounts: !conf.Get().DisableNonCriticalTelemetry,
	})
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

	// TODO(Dan): migrate this to the new usagestats package.
	//
	// For the time being, instances will report daily active users through the legacy package via this argument,
	// as well as using the new package through the `act` argument below. This will allow comparison during the
	// transition.
	count, err := usagestatsdeprecated.GetUsersActiveTodayCount()
	if err != nil {
		logFunc("usagestatsdeprecated.GetUsersActiveTodayCount failed", "error", err)
	}
	totalUsers, err := db.Users.Count(ctx, &db.UsersListOptions{})
	if err != nil {
		logFunc("db.Users.Count failed", "error", err)
	}
	totalRepos, err := db.Repos.Count(ctx, db.ReposListOptions{})
	hasRepos := totalRepos > 0
	if err != nil {
		logFunc("db.Repos.Count failed", "error", err)
	}
	searchOccurred, err := usagestats.HasSearchOccurred()
	if err != nil {
		logFunc("usagestats.HasSearchOccurred failed", "error", err)
	}
	findRefsOccurred, err := usagestats.HasFindRefsOccurred()
	if err != nil {
		logFunc("usagestats.HasFindRefsOccurred failed", "error", err)
	}
	act, err := getAndMarshalSiteActivityJSON(ctx)
	if err != nil {
		logFunc("getAndMarshalSiteActivityJSON failed", "error", err)
	}
	campaignsUsage, err := getAndMarshalCampaignsUsageJSON(ctx)
	if err != nil {
		logFunc("getAndMarshalCampaignsUsageJSON failed", "error", err)
	}
	codeIntelUsage, err := getAndMarshalCodeIntelUsageJSON(ctx)
	if err != nil {
		logFunc("getAndMarshalCodeIntelUsageJSON failed", "error", err)
	}
	searchUsage, err := getAndMarshalSearchUsageJSON(ctx)
	if err != nil {
		logFunc("getAndMarshalSearchUsageJSON failed", "error", err)
	}
	initAdminEmail, err := db.UserEmails.GetInitialSiteAdminEmail(ctx)
	if err != nil {
		logFunc("db.UserEmails.GetInitialSiteAdminEmail failed", "error", err)
	}
	svcs, err := externalServiceKinds(ctx)
	if err != nil {
		logFunc("externalServicesKinds failed", "error", err)
	}
	contents, err := json.Marshal(&pingRequest{
		ClientSiteID:         siteid.Get(),
		DeployType:           conf.DeployType(),
		ClientVersionString:  version.Version(),
		AuthProviders:        authProviderTypes(),
		ExternalServices:     svcs,
		BuiltinSignupAllowed: conf.IsBuiltinSignupAllowed(),
		HasExtURL:            conf.UsingExternalURL(),
		UniqueUsers:          int32(count),
		Activity:             act,
		CampaignsUsage:       campaignsUsage,
		CodeIntelUsage:       codeIntelUsage,
		SearchUsage:          searchUsage,
		InitialAdminEmail:    initAdminEmail,
		TotalUsers:           int32(totalUsers),
		HasRepos:             hasRepos,
		EverSearched:         hasRepos && searchOccurred, // Searches only count if repos have been added.
		EverFindRefs:         findRefsOccurred,
	})
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
	services, err := db.ExternalServices.List(ctx, db.ExternalServicesListOptions{})
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
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		_, _ = check(ctx) // updates global state on its own, can safely ignore return value
		cancel()

		// Randomize sleep to prevent thundering herds.
		randomDelay := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(delay + randomDelay)
	}
}
