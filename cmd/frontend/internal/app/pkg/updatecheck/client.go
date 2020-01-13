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
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context/ctxhttp"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats"
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

func getSiteActivityJSON() ([]byte, error) {
	days, weeks, months := 2, 1, 1
	siteActivity, err := usagestats.GetSiteUsageStatistics(&usagestats.SiteUsageStatisticsOptions{
		DayPeriods:   &days,
		WeekPeriods:  &weeks,
		MonthPeriods: &months,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(siteActivity)
}

func updateURL(ctx context.Context) string {
	logFunc := log15.Debug
	if envvar.SourcegraphDotComMode() {
		logFunc = log15.Warn
	}

	q := url.Values{}
	q.Set("version", version.Version())
	q.Set("site", siteid.Get())
	q.Set("auth", strings.Join(authProviderTypes(), ","))
	q.Set("deployType", conf.DeployType())
	q.Set("hasExtURL", strconv.FormatBool(conf.UsingExternalURL()))
	q.Set("signup", strconv.FormatBool(conf.IsBuiltinSignupAllowed()))

	count, err := usagestats.GetUsersActiveTodayCount()
	if err != nil {
		logFunc("usagestats.GetUsersActiveTodayCount failed", "error", err)
	}
	q.Set("u", strconv.Itoa(count))
	totalUsers, err := db.Users.Count(ctx, &db.UsersListOptions{})
	if err != nil {
		logFunc("db.Users.Count failed", "error", err)
	}
	q.Set("totalUsers", strconv.Itoa(totalUsers))
	totalRepos, err := db.Repos.Count(ctx, db.ReposListOptions{Enabled: true, Disabled: true})
	hasRepos := totalRepos > 0
	if err != nil {
		logFunc("db.Repos.Count failed", "error", err)
	}
	q.Set("repos", strconv.FormatBool(hasRepos))
	searchOccurred, err := usagestats.HasSearchOccurred()
	if err != nil {
		logFunc("usagestats.HasSearchOccurred failed", "error", err)
	}
	// Searches only count if repos have been added.
	q.Set("searched", strconv.FormatBool(hasRepos && searchOccurred))
	findRefsOccurred, err := usagestats.HasFindRefsOccurred()
	if err != nil {
		logFunc("usagestats.HasFindRefsOccurred failed", "error", err)
	}
	q.Set("refs", strconv.FormatBool(findRefsOccurred))
	if act, err := getSiteActivityJSON(); err != nil {
		logFunc("getSiteActivityJSON failed", "error", err)
	} else {
		q.Set("act", string(act))
	}
	initAdminEmail, err := db.UserEmails.GetInitialSiteAdminEmail(ctx)
	if err != nil {
		logFunc("db.UserEmails.GetInitialSiteAdminEmail failed", "error", err)
	}
	q.Set("initAdmin", initAdminEmail)
	svcs, err := externalServiceKinds(ctx)
	if err != nil {
		logFunc("externalServicesKinds failed", "error", err)
	}
	q.Set("extsvcs", strings.Join(svcs, ","))
	return baseURL.ResolveReference(&url.URL{RawQuery: q.Encode()}).String()
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
		resp, err := ctxhttp.Get(ctx, nil, updateURL(ctx))
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
