package updatecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/siteid"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/versions"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// metricsRecorder records operational metrics for methods.
var metricsRecorder = metrics.NewREDMetrics(prometheus.DefaultRegisterer, "updatecheck_client", metrics.WithLabels("method"))

// Status of the check for software updates for Sourcegraph.
type Status struct {
	Date          time.Time // the time that the last check completed
	Err           error     // the error that occurred, if any. When present, indicates the instance is offline / unable to contact Sourcegraph.com
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

func logFuncFrom(logger log.Logger) func(string, ...log.Field) {
	logFunc := logger.Debug
	if envvar.SourcegraphDotComMode() {
		logFunc = logger.Warn
	}

	return logFunc
}

// recordOperation returns a record fn that is called on any given return err. If an error is encountered
// it will register the err metric. The err is never altered.
func recordOperation(method string) func(*error) {
	start := time.Now()
	return func(err *error) {
		metricsRecorder.Observe(time.Since(start).Seconds(), 1, err, method)
	}
}

func getAndMarshalSiteActivityJSON(ctx context.Context, db database.DB, criticalOnly bool) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalSiteActivityJSON")(&err)
	siteActivity, err := usagestats.GetSiteUsageStats(ctx, db, criticalOnly)
	if err != nil {
		return nil, err
	}

	return json.Marshal(siteActivity)
}

func hasSearchOccurred(ctx context.Context) (_ bool, err error) {
	defer recordOperation("hasSearchOccurred")(&err)
	return usagestats.HasSearchOccurred(ctx)
}

func hasFindRefsOccurred(ctx context.Context) (_ bool, err error) {
	defer recordOperation("hasSearchOccured")(&err)
	return usagestats.HasFindRefsOccurred(ctx)
}

func getTotalUsersCount(ctx context.Context, db database.DB) (_ int, err error) {
	defer recordOperation("getTotalUsersCount")(&err)
	return db.Users().Count(ctx, &database.UsersListOptions{})
}

func getTotalOrgsCount(ctx context.Context, db database.DB) (_ int, err error) {
	defer recordOperation("getTotalUsersCount")(&err)
	return db.Orgs().Count(ctx, database.OrgsListOptions{})
}

// hasRepo returns true when the instance has at least one repository that isn't
// soft-deleted nor blocked.
func hasRepos(ctx context.Context, db database.DB) (_ bool, err error) {
	defer recordOperation("hasRepos")(&err)
	rs, err := db.Repos().List(ctx, database.ReposListOptions{
		LimitOffset: &database.LimitOffset{Limit: 1},
	})
	return len(rs) > 0, err
}

func getUsersActiveTodayCount(ctx context.Context, db database.DB) (_ int, err error) {
	defer recordOperation("getUsersActiveTodayCount")(&err)
	return usagestats.GetUsersActiveTodayCount(ctx, db)
}

func getInitialSiteAdminInfo(ctx context.Context, db database.DB) (_ string, _ bool, err error) {
	defer recordOperation("getInitialSiteAdminInfo")(&err)
	return db.UserEmails().GetInitialSiteAdminInfo(ctx)
}

func getAndMarshalBatchChangesUsageJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalBatchChangesUsageJSON")(&err)

	batchChangesUsage, err := usagestats.GetBatchChangesUsageStatistics(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Marshal(batchChangesUsage)
}

func getAndMarshalGrowthStatisticsJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalGrowthStatisticsJSON")(&err)

	growthStatistics, err := usagestats.GetGrowthStatistics(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Marshal(growthStatistics)
}

func getAndMarshalSavedSearchesJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalSavedSearchesJSON")(&err)

	savedSearches, err := usagestats.GetSavedSearches(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Marshal(savedSearches)
}

func getAndMarshalHomepagePanelsJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalHomepagePanelsJSON")(&err)

	homepagePanels, err := usagestats.GetHomepagePanels(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Marshal(homepagePanels)
}

func getAndMarshalRepositoriesJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalRepositoriesJSON")(&err)

	repos, err := usagestats.GetRepositories(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Marshal(repos)
}

func getAndMarshalRetentionStatisticsJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalRetentionStatisticsJSON")(&err)

	retentionStatistics, err := usagestats.GetRetentionStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(retentionStatistics)
}

func getAndMarshalSearchOnboardingJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalSearchOnboardingJSON")(&err)

	searchOnboarding, err := usagestats.GetSearchOnboarding(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(searchOnboarding)
}

func getAndMarshalAggregatedCodeIntelUsageJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalAggregatedCodeIntelUsageJSON")(&err)

	codeIntelUsage, err := usagestats.GetAggregatedCodeIntelStats(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(codeIntelUsage)
}

func getAndMarshalAggregatedSearchUsageJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalAggregatedSearchUsageJSON")(&err)

	searchUsage, err := usagestats.GetAggregatedSearchStats(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(searchUsage)
}

func getAndMarshalExtensionsUsageStatisticsJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalExtensionsUsageStatisticsJSON")

	extensionsUsage, err := usagestats.GetExtensionsUsageStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(extensionsUsage)
}

func getAndMarshalCodeInsightsUsageJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalCodeInsightsUsageJSON")

	codeInsightsUsage, err := usagestats.GetCodeInsightsUsageStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(codeInsightsUsage)
}

func getAndMarshalCodeInsightsCriticalTelemetryJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalCodeInsightsUsageJSON")

	insightsCriticalTelemetry, err := usagestats.GetCodeInsightsCriticalTelemetry(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(insightsCriticalTelemetry)
}

func getAndMarshalCodeMonitoringUsageJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalCodeMonitoringUsageJSON")

	codeMonitoringUsage, err := usagestats.GetCodeMonitoringUsageStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(codeMonitoringUsage)
}

func getAndMarshalNotebooksUsageJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalNotebooksUsageJSON")

	notebooksUsage, err := usagestats.GetNotebooksUsageStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(notebooksUsage)
}

func getAndMarshalCodeHostIntegrationUsageJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalCodeHostIntegrationUsageJSON")

	codeHostIntegrationUsage, err := usagestats.GetCodeHostIntegrationUsageStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(codeHostIntegrationUsage)
}

func getAndMarshalIDEExtensionsUsageJSON(ctx context.Context, db database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalIDEExtensionsUsageJSON")

	ideExtensionsUsage, err := usagestats.GetIDEExtensionsUsageStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(ideExtensionsUsage)
}

func getAndMarshalCodeHostVersionsJSON(_ context.Context, _ database.DB) (_ json.RawMessage, err error) {
	defer recordOperation("getAndMarshalCodeHostVersionsJSON")(&err)

	versions, err := versions.GetVersions()
	if err != nil {
		return nil, err
	}
	return json.Marshal(versions)
}

func getDependencyVersions(ctx context.Context, db database.DB, logger log.Logger) (json.RawMessage, error) {
	logFunc := logFuncFrom(logger.Scoped("getDependencyVersions", "gets the version of various dependency services"))
	var (
		err error
		dv  dependencyVersions
	)
	// get redis cache server version
	dv.RedisCacheVersion, err = getRedisVersion(redispool.Cache.Dial)
	if err != nil {
		logFunc("unable to get Redis cache version", log.Error(err))
	}

	// get redis store server version
	dv.RedisStoreVersion, err = getRedisVersion(redispool.Store.Dial)
	if err != nil {
		logFunc("unable to get Redis store version", log.Error(err))
	}

	// get postgres version
	err = db.QueryRowContext(ctx, "SHOW server_version").Scan(&dv.PostgresVersion)
	if err != nil {
		logFunc("unable to get Postgres version", log.Error(err))
	}
	return json.Marshal(dv)
}

func getRedisVersion(dialFunc func() (redis.Conn, error)) (string, error) {
	conn, err := dialFunc()
	if err != nil {
		return "", err
	}
	buf, err := redis.Bytes(conn.Do("INFO"))
	if err != nil {
		return "", err
	}

	m, err := parseRedisInfo(buf)
	return m["redis_version"], err
}

func parseRedisInfo(buf []byte) (map[string]string, error) {
	var (
		lines = bytes.Split(buf, []byte("\n"))
		m     = make(map[string]string, len(lines))
	)

	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("#")) || len(line) == 0 {
			continue
		}

		parts := bytes.Split(line, []byte(":"))
		if len(parts) != 2 {
			return nil, errors.Errorf("expected a key:value line, got %q", string(line))
		}
		m[string(parts[0])] = string(parts[1])
	}

	return m, nil
}

func updateBody(ctx context.Context, logger log.Logger, db database.DB) (io.Reader, error) {
	scopedLog := logger.Scoped("telemetry", "track and update various usages stats")
	logFunc := scopedLog.Debug
	if envvar.SourcegraphDotComMode() {
		logFunc = scopedLog.Warn
	}

	r := &pingRequest{
		ClientSiteID:                  siteid.Get(),
		DeployType:                    deploy.Type(),
		ClientVersionString:           version.Version(),
		LicenseKey:                    conf.Get().LicenseKey,
		CodeIntelUsage:                []byte("{}"),
		NewCodeIntelUsage:             []byte("{}"),
		SearchUsage:                   []byte("{}"),
		BatchChangesUsage:             []byte("{}"),
		GrowthStatistics:              []byte("{}"),
		SavedSearches:                 []byte("{}"),
		HomepagePanels:                []byte("{}"),
		Repositories:                  []byte("{}"),
		RetentionStatistics:           []byte("{}"),
		SearchOnboarding:              []byte("{}"),
		ExtensionsUsage:               []byte("{}"),
		CodeInsightsUsage:             []byte("{}"),
		CodeInsightsCriticalTelemetry: []byte("{}"),
		CodeMonitoringUsage:           []byte("{}"),
		NotebooksUsage:                []byte("{}"),
		CodeHostIntegrationUsage:      []byte("{}"),
		IDEExtensionsUsage:            []byte("{}"),
	}

	totalUsers, err := getTotalUsersCount(ctx, db)
	if err != nil {
		logFunc("database.Users.Count failed", log.Error(err))
	}
	r.TotalUsers = int32(totalUsers)
	r.InitialAdminEmail, r.TosAccepted, err = getInitialSiteAdminInfo(ctx, db)
	if err != nil {
		logFunc("database.UserEmails.GetInitialSiteAdminInfo failed", log.Error(err))
	}

	r.DependencyVersions, err = getDependencyVersions(ctx, db, logger)
	if err != nil {
		logFunc("getDependencyVersions failed", log.Error(err))
	}

	// Yes dear reader, this is a feature ping in critical telemetry. Why do you ask? Because for the purposes of
	// licensing enforcement, we need to know how many insights our customers have created. Please see RFC 584
	// for the original approval of this ping. (https://docs.google.com/document/d/1J-fnZzRtvcZ_NWweCZQ5ipDMh4NdgQ8rlxXsa8vHWlQ/edit#)
	r.CodeInsightsCriticalTelemetry, err = getAndMarshalCodeInsightsCriticalTelemetryJSON(ctx, db)
	if err != nil {
		logFunc("getAndMarshalCodeInsightsCriticalTelemetry failed", log.Error(err))
	}

	if !conf.Get().DisableNonCriticalTelemetry {
		// TODO(Dan): migrate this to the new usagestats package.
		//
		// For the time being, instances will report daily active users through the legacy package via this argument,
		// as well as using the new package through the `act` argument below. This will allow comparison during the
		// transition.
		count, err := getUsersActiveTodayCount(ctx, db)
		if err != nil {
			logFunc("getUsersActiveToday failed", log.Error(err))
		}
		r.UniqueUsers = int32(count)

		totalOrgs, err := getTotalOrgsCount(ctx, db)
		if err != nil {
			logFunc("database.Orgs.Count failed", log.Error(err))
		}
		r.TotalOrgs = int32(totalOrgs)

		r.HasRepos, err = hasRepos(ctx, db)
		if err != nil {
			logFunc("hasRepos failed", log.Error(err))
		}

		r.EverSearched, err = hasSearchOccurred(ctx)
		if err != nil {
			logFunc("hasSearchOccurred failed", log.Error(err))
		}
		r.EverFindRefs, err = hasFindRefsOccurred(ctx)
		if err != nil {
			logFunc("hasFindRefsOccurred failed", log.Error(err))
		}
		r.BatchChangesUsage, err = getAndMarshalBatchChangesUsageJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalBatchChangesUsageJSON failed", log.Error(err))
		}
		r.GrowthStatistics, err = getAndMarshalGrowthStatisticsJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalGrowthStatisticsJSON failed", log.Error(err))
		}

		r.SavedSearches, err = getAndMarshalSavedSearchesJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalSavedSearchesJSON failed", log.Error(err))
		}

		r.HomepagePanels, err = getAndMarshalHomepagePanelsJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalHomepagePanelsJSON failed", log.Error(err))
		}

		r.SearchOnboarding, err = getAndMarshalSearchOnboardingJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalSearchOnboardingJSON failed", log.Error(err))
		}

		r.Repositories, err = getAndMarshalRepositoriesJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalRepositoriesJSON failed", log.Error(err))
		}

		r.RetentionStatistics, err = getAndMarshalRetentionStatisticsJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalRetentionStatisticsJSON failed", log.Error(err))
		}

		r.ExtensionsUsage, err = getAndMarshalExtensionsUsageStatisticsJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalExtensionsUsageStatisticsJSON failed", log.Error(err))
		}

		r.CodeInsightsUsage, err = getAndMarshalCodeInsightsUsageJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalCodeInsightsUsageJSON failed", log.Error(err))
		}

		r.CodeMonitoringUsage, err = getAndMarshalCodeMonitoringUsageJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalCodeMonitoringUsageJSON failed", log.Error(err))
		}

		r.NotebooksUsage, err = getAndMarshalNotebooksUsageJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalNotebooksUsageJSON failed", log.Error(err))
		}

		r.CodeHostIntegrationUsage, err = getAndMarshalCodeHostIntegrationUsageJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalCodeHostIntegrationUsageJSON failed", log.Error(err))
		}

		r.IDEExtensionsUsage, err = getAndMarshalIDEExtensionsUsageJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalIDEExtensionsUsageJSON failed", log.Error(err))
		}

		r.CodeHostVersions, err = getAndMarshalCodeHostVersionsJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalCodeHostVersionsJSON failed", log.Error(err))
		}

		r.ExternalServices, err = externalServiceKinds(ctx, db)
		if err != nil {
			logFunc("externalServicesKinds failed", log.Error(err))
		}

		r.HasExtURL = conf.UsingExternalURL()
		r.BuiltinSignupAllowed = conf.IsBuiltinSignupAllowed()
		r.AuthProviders = authProviderTypes()

		// The following methods are the most expensive to calculate, so we do them in
		// parallel.

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Activity, err = getAndMarshalSiteActivityJSON(ctx, db, false)
			if err != nil {
				logFunc("getAndMarshalSiteActivityJSON failed", log.Error(err))
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			r.NewCodeIntelUsage, err = getAndMarshalAggregatedCodeIntelUsageJSON(ctx, db)
			if err != nil {
				logFunc("getAndMarshalAggregatedCodeIntelUsageJSON failed", log.Error(err))
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			r.SearchUsage, err = getAndMarshalAggregatedSearchUsageJSON(ctx, db)
			if err != nil {
				logFunc("getAndMarshalAggregatedSearchUsageJSON failed", log.Error(err))
			}
		}()

		wg.Wait()
	} else {
		r.Repositories, err = getAndMarshalRepositoriesJSON(ctx, db)
		if err != nil {
			logFunc("getAndMarshalRepositoriesJSON failed", log.Error(err))
		}

		r.Activity, err = getAndMarshalSiteActivityJSON(ctx, db, true)
		if err != nil {
			logFunc("getAndMarshalSiteActivityJSON failed", log.Error(err))
		}
	}

	contents, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	err = db.EventLogs().Insert(ctx, &database.Event{
		UserID:          0,
		Name:            "ping",
		URL:             "",
		AnonymousUserID: "backend",
		Source:          "BACKEND",
		Argument:        contents,
		Timestamp:       time.Now().UTC(),
	})

	return bytes.NewReader(contents), err
}

func authProviderTypes() []string {
	ps := conf.Get().AuthProviders
	types := make([]string, len(ps))
	for i, p := range ps {
		types[i] = conf.AuthProviderType(p)
	}
	return types
}

func externalServiceKinds(ctx context.Context, db database.DB) (kinds []string, err error) {
	defer recordOperation("externalServiceKinds")(&err)
	kinds, err = db.ExternalServices().DistinctKinds(ctx)
	return kinds, err
}

// check performs an update check and updates the global state.
func check(logger log.Logger, db database.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	doCheck := func() (updateVersion string, err error) {
		body, err := updateBody(ctx, logger, db)
		if err != nil {
			return "", err
		}

		req, err := http.NewRequest("POST", "https://sourcegraph.com/.api/updates", body)
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)

		resp, err := httpcli.ExternalDoer.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			var description string
			if body, err := io.ReadAll(io.LimitReader(resp.Body, 30)); err != nil {
				description = err.Error()
			} else if len(body) == 0 {
				description = "(no response body)"
			} else {
				description = strconv.Quote(string(bytes.TrimSpace(body)))
			}
			return "", errors.Errorf("update endpoint returned HTTP error %d: %s", resp.StatusCode, description)
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
	if err != nil {
		logger.Error("updatecheck failed", log.Error(err))
	}

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
}

var started bool

// Start starts checking for software updates periodically.
func Start(logger log.Logger, db database.DB) {
	if started {
		panic("already started")
	}
	started = true

	if channel := conf.UpdateChannel(); channel != "release" {
		return // no update check
	}

	const delay = 30 * time.Minute
	scopedLog := logger.Scoped("updatecheck", "checks for updates of services and updates usage telemetry")
	for {
		check(scopedLog, db)

		// Randomize sleep to prevent thundering herds.
		randomDelay := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(delay + randomDelay)
	}
}
