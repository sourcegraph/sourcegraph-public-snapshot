package updatecheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/eventlogger"
	"github.com/sourcegraph/sourcegraph/pkg/hubspot"
	"github.com/sourcegraph/sourcegraph/pkg/pubsub/pubsubutil"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	// latestReleaseDockerServerImageBuild is only used by sourcegraph.com to tell existing
	// non-cluster installations what the latest version is. The version here _must_ be
	// available at https://hub.docker.com/r/sourcegraph/server/tags/ before
	// landing in master.
	latestReleaseDockerServerImageBuild = newBuild("3.0.1")

	// latestReleaseKubernetesBuild is only used by sourcegraph.com to tell existing Sourcegraph
	// cluster deployments what the latest version is. The version here _must_ be available in
	// a tag at https://github.com/sourcegraph/deploy-sourcegraph before landing in master.
	latestReleaseKubernetesBuild = newBuild("3.0.1")
)

func getLatestRelease(deployType string) build {
	if conf.IsDeployTypeCluster(deployType) {
		return latestReleaseKubernetesBuild
	}
	return latestReleaseDockerServerImageBuild
}

// Handler is an HTTP handler that responds with information about software updates
// for Sourcegraph.
func Handler(w http.ResponseWriter, r *http.Request) {
	requestCounter.Inc()

	q := r.URL.Query()
	clientSiteID := q.Get("site")
	deployType := q.Get("deployType")
	clientVersionString := q.Get("version")
	if clientSiteID == "" {
		log15.Error("updatecheck: no site ID specified")
		http.Error(w, "no site ID specified", http.StatusBadRequest)
		return
	}
	if clientVersionString == "" {
		log15.Error("updatecheck: no version specified")
		http.Error(w, "no version specified", http.StatusBadRequest)
		return
	}
	if clientVersionString == "dev" {
		// No updates for dev servers.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	latestReleaseBuild := getLatestRelease(deployType)
	hasUpdate, err := canUpdate(clientVersionString, latestReleaseBuild)
	if err != nil {
		// Still log pings on malformed version strings.
		logPing(r, clientVersionString, false)

		http.Error(w, clientVersionString+" is a bad version string: "+err.Error(), http.StatusBadRequest)
		return
	}

	logPing(r, clientVersionString, hasUpdate)

	if !hasUpdate {
		// No newer version.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("content-type", "application/json; charset=utf-8")
	body, err := json.Marshal(latestReleaseBuild)
	if err != nil {
		log15.Error("updatecheck: error preparing update check response", "error", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	requestHasUpdateCounter.Inc()
	_, _ = w.Write(body)
}

// canUpdate returns true if the latestReleaseBuild is newer than the clientVersionString.
func canUpdate(clientVersionString string, latestReleaseBuild build) (bool, error) {
	// Check for a date in the version string to handle developer builds that don't have a semver.
	// If there is an error parsing a date out of the version string, then we ignore the error
	// and parse it as a semver.
	if hasDateUpdate, err := canUpdateDate(clientVersionString); err == nil {
		return hasDateUpdate, nil
	}

	// Released builds will have a semantic version that we can compare.
	return canUpdateVersion(clientVersionString, latestReleaseBuild)
}

// canUpdateVersion returns true if the latest released build is newer than
// the clientVersionString. It returns an error if clientVersionString is not a semver.
func canUpdateVersion(clientVersionString string, latestReleaseBuild build) (bool, error) {
	clientVersionString = strings.TrimPrefix(clientVersionString, "v")
	clientVersion, err := semver.NewVersion(clientVersionString)
	if err != nil {
		return false, err
	}
	return clientVersion.LessThan(latestReleaseBuild.Version), nil
}

var dateRegex = regexp.MustCompile("_([0-9]{4}-[0-9]{2}-[0-9]{2})_")
var timeNow = time.Now

// canUpdateDate returns true if clientVersionString contains a date
// more than 40 days in the past. It returns an error if there is no
// parsable date in clientVersionString
func canUpdateDate(clientVersionString string) (bool, error) {
	match := dateRegex.FindStringSubmatch(clientVersionString)
	if len(match) != 2 {
		return false, fmt.Errorf("no date in version string %q", clientVersionString)
	}

	t, err := time.ParseInLocation("2006-01-02", match[1], time.UTC)
	if err != nil {
		// This shouldn't ever happen if the above code is correct.
		return false, err
	}

	// Assume that we release a new version at least every 40 days.
	return timeNow().After(t.Add(40 * 24 * time.Hour)), nil
}

func logPing(r *http.Request, clientVersionString string, hasUpdate bool) {
	q := r.URL.Query()
	clientSiteID := q.Get("site")
	authProviders := q.Get("auth")
	builtinSignupAllowed := q.Get("signup")
	hasExtURL := q.Get("hasExtURL")
	uniqueUsers := q.Get("u")
	activity := q.Get("act")
	initialAdminEmail := q.Get("initAdmin")
	deployType := q.Get("deployType")
	totalUsers := q.Get("totalUsers")
	hasRepos := q.Get("repos")
	everSearched := q.Get("searched")
	everFindRefs := q.Get("refs")

	// Log update check.
	var clientAddr string
	if v := r.Header.Get("x-forwarded-for"); v != "" {
		clientAddr = v
	} else {
		clientAddr = r.RemoteAddr
	}

	// Prevent nil activity data (i.e., "") from breaking json marshaling.
	// This is an issue with all instances at versions < 2.7.0.
	if activity == "" {
		activity = `{}`
	}

	message := fmt.Sprintf(`{
		"remote_ip": "%s",
		"remote_site_version": "%s",
		"remote_site_id": "%s",
		"has_update": "%s",
		"unique_users_today": "%s",
		"site_activity": %s,
		"installer_email": "%s",
		"auth_providers": "%s",
		"builtin_signup_allowed": "%s",
		"deploy_type": "%s",
		"total_user_accounts": "%s",
		"has_external_url": "%s",
		"has_repos": "%s",
		"ever_searched": "%s",
		"ever_find_refs": "%s",
		"timestamp": "%s"
	}`,
		clientAddr,
		clientVersionString,
		clientSiteID,
		strconv.FormatBool(hasUpdate),
		uniqueUsers,
		activity,
		initialAdminEmail,
		authProviders,
		builtinSignupAllowed,
		deployType,
		totalUsers,
		hasExtURL,
		hasRepos,
		everSearched,
		everFindRefs,
		time.Now().UTC().Format(time.RFC3339),
	)

	eventlogger.LogEvent("", "ServerUpdateCheck", json.RawMessage(message))

	if pubsubutil.Enabled() {
		err := pubsubutil.Publish(pubsubutil.PubSubTopicID, message)
		if err != nil {
			log15.Warn("pubsubutil.Publish: failed to Publish", "message", message, "error", err)
		}
	}

	// Sync the initial administrator email in HubSpot.
	if initialAdminEmail != "" && strings.Contains(initialAdminEmail, "@") {
		go tracking.SyncUser(initialAdminEmail, "", &hubspot.ContactProperties{IsServerAdmin: true})
	}
}

var (
	requestCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "updatecheck",
		Name:      "requests",
		Help:      "Number of requests to the update check handler.",
	})
	requestHasUpdateCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "updatecheck",
		Name:      "requests_has_update",
		Help:      "Number of requests to the update check handler where an update is available.",
	})
)

func init() {
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestHasUpdateCounter)
}
