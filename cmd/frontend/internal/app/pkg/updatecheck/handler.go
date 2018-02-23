package updatecheck

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventlogger"

	"github.com/coreos/go-semver/semver"
	"github.com/prometheus/client_golang/prometheus"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	// ProductVersion is a semver version string that corresponds to this
	// product's version number, without any build or tag information. This is
	// compared against the remote handler's build.Assets.ProductVersion
	// field.
	//
	// When we speak to sourcegraph.com we report this version. Usually should
	// be equal to latestReleaseBuild.Version, unless we are in the process of
	// doing a release, in which case it should be one version ahead.
	ProductVersion = "2.5.12"

	// latestReleaseBuild is only used by Sourcegraph.com to tell existing
	// installations what the latest version is. The version here _must_ be
	// available at https://hub.docker.com/r/sourcegraph/server/tags/ before
	// landing in master.
	latestReleaseBuild   = newBuild(1519249026, "2.5.12")
	latestReleaseVersion = *semver.New(latestReleaseBuild.Version)
)

func newBuild(timestamp int64, version string) build {
	return build{
		Timestamp:  timestamp,
		Version:    version,
		IsReleased: true,
		Assets: []asset{
			{
				Name:           "docker-image",
				Version:        version,
				ProductVersion: version,
				Platform:       "docker",
				Type:           "docker-image",
				URL:            "docker.io/sourcegraph/server:" + version,
			},
		},
	}
}

// Handler is an HTTP handler that responds with information about software updates
// for Sourcegraph Server.
func Handler(w http.ResponseWriter, r *http.Request) {
	requestCounter.Inc()

	q := r.URL.Query()
	clientVersionString := q.Get("version")
	clientSiteID := q.Get("site")
	uniqueUsers := q.Get("u")
	hasCodeIntelligence := q.Get("codeintel")
	if clientVersionString == "" {
		http.Error(w, "no version specified", http.StatusBadRequest)
		return
	}
	if clientVersionString == "dev" {
		// No updates for dev servers.
		w.WriteHeader(http.StatusNoContent)
		return
	}
	clientVersion, err := semver.NewVersion(clientVersionString)
	if err != nil {
		http.Error(w, "bad version string: "+err.Error(), http.StatusBadRequest)
		return
	}
	if clientSiteID == "" {
		http.Error(w, "no site ID specified", http.StatusBadRequest)
		return
	}

	hasUpdate := clientVersion.LessThan(latestReleaseVersion)

	{
		// Log update check.
		var clientAddr string
		if v := r.Header.Get("x-forwarded-for"); v != "" {
			clientAddr = v
		} else {
			clientAddr = r.RemoteAddr
		}

		eventlogger.LogEvent("", "ServerUpdateCheck", map[string]interface{}{
			"remote_ip":             clientAddr,
			"remote_site_version":   clientVersionString,
			"remote_site_id":        clientSiteID,
			"has_update":            strconv.FormatBool(hasUpdate),
			"unique_users_today":    uniqueUsers,
			"has_code_intelligence": hasCodeIntelligence,
		})
	}

	if !hasUpdate {
		// No newer version.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("content-type", "application/json; charset=utf-8")
	body, err := json.Marshal(latestReleaseBuild)
	if err != nil {
		log15.Error("error preparing update check response", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	requestHasUpdateCounter.Inc()
	_, _ = w.Write(body)
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
