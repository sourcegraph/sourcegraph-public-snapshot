package codyapp

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
)

// RouteCodyAppLatestVersion is the name of the route that that returns a URL where to download the latest Cody App version
const RouteCodyAppLatestVersion = "codyapp.latest.version"

const gitHubReleaseBaseURL = "https://github.com/sourcegraph/sourcegraph/releases/tag/"

type latestVersion struct {
	logger           log.Logger
	manifestResolver UpdateManifestResolver
}

func (l *latestVersion) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		manifest, err := l.manifestResolver.Resolve(ctx)
		if err != nil {
			l.logger.Error("failed to resolve manifest", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		query := r.URL.Query()
		target := query.Get("target")
		arch := query.Get("arch")
		platform := PlatformString(arch, target) // x86_64-darwin

		releaseURL, err := url.Parse(gitHubReleaseBaseURL)
		if err != nil {
			l.logger.Error("failed to create release url from base release url", log.Error(err), log.String("releaseTag", manifest.GitHubReleaseTag()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		releaseURL = releaseURL.JoinPath(manifest.GitHubReleaseTag())

		releaseLoc, hasPlatform := manifest.Platforms[platform]
		// if we have the platform, get it's release URL and redirect to it.
		// if we don't have it or something goes wrong while converting to a URL, we
		// redirect to the GitHub release page
		if hasPlatform {
			u, err := url.Parse(releaseLoc.URL)
			if err != nil {
				l.logger.Error("failed to create release url for platform - redirecting to release page instead",
					log.Error(err),
					log.String("platform", platform),
					log.String("releaseTag", manifest.GitHubReleaseTag()),
				)
				http.Redirect(w, r, releaseURL.String(), http.StatusSeeOther)
				return
			}
			releaseURL = u
		}

		http.Redirect(w, r, patchReleaseURL(releaseURL.String()), http.StatusSeeOther)
	}
}

// (Hack) patch the release URL so that Mac users get a DMG instead of a .tar.gz download
func patchReleaseURL(u string) string {
	if suffix := ".aarch64.app.tar.gz"; strings.HasSuffix(u, suffix) {
		u = strings.ReplaceAll(u, "Cody.", "Cody_")
		u = strings.ReplaceAll(u, suffix, "_aarch64.dmg")
	}
	if suffix := ".x86_64.app.tar.gz"; strings.HasSuffix(u, suffix) {
		u = strings.ReplaceAll(u, "Cody.", "Cody_")
		u = strings.ReplaceAll(u, suffix, "_x64.dmg")
	}
	return u
}

func newLatestVersion(logger log.Logger, resolver UpdateManifestResolver) *latestVersion {
	return &latestVersion{
		logger:           logger,
		manifestResolver: resolver,
	}
}

func LatestVersionHandler(logger log.Logger) http.HandlerFunc {
	var bucket = ManifestBucket

	if deploy.IsDev(deploy.Type()) {
		bucket = ManifestBucketDev
	}

	resolver, err := NewGCSManifestResolver(context.Background(), bucket, ManifestName)
	if err != nil {
		logger.Error("failed to initialize GCS Manifest resolver",
			log.String("bucket", bucket),
			log.String("manifestName", ManifestName),
			log.Error(err),
		)
		return func(w http.ResponseWriter, _ *http.Request) {
			logger.Warn("GCS Manifest resolver not initialized. Unable to respond with latest App version")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	return newLatestVersion(logger, resolver).Handler()
}
