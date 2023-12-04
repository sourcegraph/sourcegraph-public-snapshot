package codyapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
)

// RouteAppUpdateCheck is the name of the route that the Cody App will use to check if there are updates
const RouteAppUpdateCheck = "app.update.check"

// ManifestBucket the name of the bucket where the Cody App update manifest is stored
const ManifestBucket = "sourcegraph-app"

// ManifestBucketDev the name of the bucket where the Cody App update manifest is stored for dev instances
const ManifestBucketDev = "sourcegraph-app-dev"

// ManifestName is the name of the manifest object that is in the ManifestBucket
const ManifestName = "app.update.prod.manifest.json"

// noUpdateConstraint clients on or prior to this version are using the "Cody App" version, which is the version prior to the
// "Cody App" version which does not have search. Therefore, clients that match this constraint should be told that there is NOT a
// new version for them to update to with the Tauri updater. Instead we will notify them with a banner in the app - which is not
// part of the Tauri updater.
var noUpdateConstraint = mustConstraint("<= 2023.6.13")

type AppUpdateResponse struct {
	Version   string    `json:"version"`
	Notes     string    `json:"notes,omitempty"`
	PubDate   time.Time `json:"pub_date"`
	Signature string    `json:"signature"`
	URL       string    `json:"url"`
}

type AppUpdateChecker struct {
	logger           log.Logger
	manifestResolver UpdateManifestResolver
}

type AppNoopUpdateChecker struct{}

func NewAppUpdateChecker(logger log.Logger, resolver UpdateManifestResolver) *AppUpdateChecker {
	return &AppUpdateChecker{
		logger:           logger.Scoped("app.update.checker"),
		manifestResolver: resolver,
	}
}

func (checker *AppUpdateChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appClientVersion := readClientAppVersion(r.URL)
		if err := appClientVersion.validate(); err != nil {
			checker.logger.Error("app client version failed validation", log.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		manifest, err := checker.manifestResolver.Resolve(ctx)
		if err != nil {
			checker.logger.Error("failed to resolve update manifest", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		checker.logger.Info("app update check", log.Object("App",
			log.String("target", appClientVersion.Target),
			log.String("version", appClientVersion.Version),
			log.String("arch", appClientVersion.Arch),
		))

		if canUpdate, err := checker.canUpdate(appClientVersion, manifest); err != nil {
			checker.logger.Error("failed to check app client version for update",
				log.String("clientVersion", appClientVersion.Version), log.String("manifestVersion", manifest.Version))
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if !canUpdate {
			// No update
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var platformLoc AppLocation
		if p, ok := manifest.Platforms[appClientVersion.Platform()]; !ok {
			// we don't have this platform in our manifest, so this is just a bad request
			checker.logger.Error("platform not found in App Update Manifest", log.String("platform", appClientVersion.Platform()))
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			platformLoc = p
		}

		checker.logger.Debug("found client platform in App Update Manifest", log.Object("platform", log.String("signature", platformLoc.Signature), log.String("url", platformLoc.URL)))

		var notes = "A new Sourcegraph version is available! For more information see https://github.com/sourcegraph/sourcegraph/releases"
		if len(manifest.Notes) > 0 {
			notes = manifest.Notes
		}

		updateResp := AppUpdateResponse{
			Version:   manifest.Version,
			PubDate:   manifest.PubDate,
			Notes:     notes,
			Signature: platformLoc.Signature,
			URL:       platformLoc.URL,
		}

		// notify the app client that they can update
		err = json.NewEncoder(w).Encode(updateResp)
		if err != nil {
			checker.logger.Error("failed to encode App Update Response", log.Error(err), log.Object("resp",
				log.String("version", updateResp.Version),
				log.Time("PubDate", updateResp.PubDate),
				log.String("Notes", updateResp.Notes),
				log.String("Signature", updateResp.Signature),
				log.String("URL", updateResp.URL),
			))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func readClientAppVersion(reqURL *url.URL) *AppVersion {
	queryValues := reqURL.Query()
	var appClientVersion = AppVersion{}
	for key, attr := range map[string]*string{
		"target":          &appClientVersion.Target,
		"current_version": &appClientVersion.Version,
		"arch":            &appClientVersion.Arch,
	} {
		if v, ok := queryValues[key]; ok && len(v) > 0 {
			*attr = v[0]
		}
	}

	// The app versions contain '+' and Tauri is not encoding the updater url
	// this is being interpreted as a blank space and breaking the semver check.
	// Trimming all leading/trailing spaces then replacing spaces with '+' to get auto updates working.
	appClientVersion.Version = strings.ReplaceAll(strings.TrimSpace(appClientVersion.Version), " ", "+")

	return &appClientVersion
}

func (checker *AppUpdateChecker) canUpdate(client *AppVersion, manifest *AppUpdateManifest) (bool, error) {
	clientVersion, err := semver.NewVersion(client.Version)
	if err != nil {
		return false, err
	}
	manifestVersion, err := semver.NewVersion(manifest.Version)
	if err != nil {
		return false, err
	}

	// no updates for clients that match this constraint!
	if noUpdateConstraint.Check(clientVersion) {
		return false, nil
	}

	// if the manifest version is higher than then the clientVersion, then the client can upgrade
	return manifestVersion.Compare(clientVersion) > 0, nil
}

func (checker *AppNoopUpdateChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// No update
		w.WriteHeader(http.StatusNoContent)
	}
}

func AppUpdateHandler(logger log.Logger) http.HandlerFunc {
	// We store the Cody App manifest in a different GCS bucket, since buckets are globally unique we use different names
	var bucket = ManifestBucket
	if deploy.IsDev(deploy.Type()) {
		bucket = ManifestBucketDev
	}
	resolver, err := NewGCSManifestResolver(context.Background(), bucket, ManifestName)
	if err != nil {
		logger.Error("failed to initialize GCS Manifest resolver. Using NoopUpdateChecker which will tell all clients that there are no updates", log.Error(err))
		return (&AppNoopUpdateChecker{}).Handler()
	} else {
		return NewAppUpdateChecker(logger, resolver).Handler()
	}
}

func mustConstraint(c string) *semver.Constraints {
	constraint, err := semver.NewConstraint(c)
	if err != nil {
		panic(fmt.Sprintf("invalid constraint %q: %v", c, err))
	}

	return constraint
}
