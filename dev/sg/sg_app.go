package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/go-github/v41/github"
	"github.com/urfave/cli/v2"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var appCommand = &cli.Command{
	Name:  "app",
	Usage: "Manage releases and update manifests used to let Sourcegraph App clients know that a new update is available",
	UsageText: `
# Update the updater manifest
sg app update-manifest

# Update the updater manifest based on a particular github release
sg app update-manifest --release-tag app-v2023.07.07

# Do a dry run of updating the manifest
sg app update-manifest --dry-run
`,
	Description: `
Various commands to handle management of releases, and processes around Sourcegraph App.

`,
	ArgsUsage: "",
	Category:  CategoryDev,
	Subcommands: []*cli.Command{
		{
			Name:   "update-manifest",
			Usage:  "update the manifest used by the updater endpoint on dotCom",
			Action: UpdateSourcegraphAppManifest,
		},
	},
}

// appUpdateManifest is copied from cmd/frontend/internal/app/updatecheck/app_update_checker
type appUpdateManifest struct {
	Version   string      `json:"version"`
	Notes     string      `json:"notes"`
	PubDate   time.Time   `json:"pub_date"`
	Platforms appPlatform `json:"platforms"`
}

type appPlatform map[string]appLocation

type appLocation struct {
	Signature string `json:"signature"`
	URL       string `json:"url"`
}

func UpdateSourcegraphAppManifest(ctx *cli.Context) error {
	client, err := bk.NewClient(ctx.Context, std.Out)
	if err != nil {
		return err
	}

	pipeline := "sourcegraph-app-release"
	branch := "app-release/stable"

	build, err := client.GetMostRecentBuild(ctx.Context, pipeline, branch)
	if err != nil {
		return err
	}

	manifestArtifact, err := findArtifactByBuild(ctx.Context, client, build, "app.update.manifest")
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	client.DownloadArtifact(*manifestArtifact, buf)

	manifest := appUpdateManifest{}
	err = json.NewDecoder(buf).Decode(&manifest)
	if err != nil {
		return err
	}

	githubClient := github.NewClient(http.DefaultClient)
	release, err := getAppGitHubRelease(ctx.Context, githubClient, "latest")
	if err != nil {
		return err
	}

	updateSignatures := (fmt.Sprintf("app-v%s", manifest.Version) != release.GetTagName())
	manifest, err = updateManifestFromRelease(manifest, release, updateSignatures)
	if err != nil {
		return err
	}

	storageClient, err := storage.NewClient(ctx.Context)
	if err != nil {
		return err
	}

	storageWriter := storageClient.Bucket("sourcegraph-app-dev").Object("app.update.prod.manifest.json").NewWriter(ctx.Context)
	err = json.NewEncoder(storageWriter).Encode(&manifest)
	defer func() {
		if err := storageWriter.Close(); err != nil {
			std.Out.WriteFailuref("Google Storage Writer failed on close: %v", err)
		}
	}()
	if err != nil {
		return err
	}

	return storageWriter.Close()
}

func updateManifestFromRelease(manifest appUpdateManifest, release *github.RepositoryRelease, updateSignatures bool) (appUpdateManifest, error) {
	platformMatch := map[string]*regexp.Regexp{
		// note the regular expression will capture
		// .tar.gz
		// AND
		// .tar.gz.sig
		"aarch64-darwin": regexp.MustCompile("^Sourcegraph.*.aarch64.app.tar.gz"),
		"x86_64-darwin":  regexp.MustCompile("^Sourcegraph.*.x86_64.app.tar.gz"),
		"x86_64-linux":   regexp.MustCompile("^sourcegraph.*_amd64.AppImage.tar.gz"),
	}

	platformAssets := map[string][]*github.ReleaseAsset{
		"aarch64-darwin": make([]*github.ReleaseAsset, 2),
		"x86_64-darwin":  make([]*github.ReleaseAsset, 2),
		"x86_64-linux":   make([]*github.ReleaseAsset, 2),
	}

	for _, asset := range release.Assets {
		for platform, re := range platformMatch {
			if re.MatchString(asset.GetName()) {
				if strings.HasSuffix(asset.GetName(), ".sig") {
					platformAssets[platform][1] = asset
				} else {
					platformAssets[platform][0] = asset
				}
			}

		}
	}

	// update the manifest
	for platform, assets := range platformAssets {
		appPlatform := manifest.Platforms[platform]
		u := assets[0].GetBrowserDownloadURL()
		if u == "" {
			return manifest, errors.Newf("failed to get download url for asset: %q", assets[0].GetName())
		}
		var sig = appPlatform.Signature

		if updateSignatures {
			b, err := downloadSignatureContent(assets[1].GetBrowserDownloadURL())
			if err != nil {
				return manifest, errors.Wrapf(err, "failed to content of signature asset %q", assets[1].GetName())
			}
			sig = string(b)
		}

		appPlatform.URL = u
		appPlatform.Signature = sig

		manifest.Platforms[platform] = appPlatform
	}

	return manifest, nil
}

func downloadSignatureContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getAppGitHubRelease(ctx context.Context, client *github.Client, tag string) (*github.RepositoryRelease, error) {

	releases, _, err := client.Repositories.ListReleases(ctx, "sourcegraph", "sourcegraph", &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	var releaseCompareFn func(release *github.RepositoryRelease) bool

	// if tag is empty, we take the latest release, otherwise we look for a release with the tag
	if tag == "latest" {
		releaseCompareFn = func(release *github.RepositoryRelease) bool {
			return strings.Contains(release.GetName(), "Sourcegraph App")
		}
	} else {
		releaseCompareFn = func(release *github.RepositoryRelease) bool {
			return strings.Contains(release.GetName(), "Sourcegraph App") && release.GetTagName() == tag
		}
	}

	var appRelease *github.RepositoryRelease
	for _, r := range releases {
		if ok := releaseCompareFn(r); ok {
			appRelease = r
			break
		}
	}
	if appRelease == nil {
		return nil, errors.Newf("failed to find Sourcegraph App Release tag %q", tag)
	}
	return appRelease, nil
}

func findArtifactByBuild(ctx context.Context, client *bk.Client, build *buildkite.Build, artifactName string) (*buildkite.Artifact, error) {
	buildNumber := strconv.Itoa(*build.Number)
	artifacts, err := client.ListArtifactsByBuildNumber(ctx, *build.Pipeline.Slug, buildNumber)
	if err != nil {
		return nil, err
	}

	for _, a := range artifacts {
		name := *a.Filename
		if name == artifactName {
			return &a, nil
		}
	}

	return nil, errors.Newf("failed to find artifact %q on build %q", artifactName, buildNumber)
}
