package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/go-github/v55/github"
	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type resetFlag struct {
	dryRun bool
}

var resetFlags resetFlag

type updateManifestFlag struct {
	bucket           string
	build            int
	tag              string
	updateSignatures bool
	noUpload         bool
}

var (
	manifestFlags updateManifestFlag
	appCommand    = &cli.Command{
		Name:  "app",
		Usage: "Manage releases and update manifests used to let Cody App clients know that a new update is available",
		UsageText: `
# Update the updater manifest
sg app update-manifest

# Update the updater manifest based on a particular github release
sg app update-manifest --release-tag app-v2023.07.07

# Do everything except upload the updated manifest
sg app update-manifest --no-upload

# Update the manifest but don't update the signatures from the release - useful if the release comes from the same build
sg app update-manifest --update-signatures

# Resets the dev app's db and web cache
sg app reset

# Prints the locations to be removed without deleting
sg app reset --dry-run
`,
		Description: `
Various commands to handle management of releases, and processes around Cody App.

`,
		ArgsUsage: "",
		Category:  category.Dev,
		Subcommands: []*cli.Command{
			{
				Name:  "update-manifest",
				Usage: "update the manifest used by the updater endpoint on dotCom",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "bucket",
						HasBeenSet:  true,
						Value:       "sourcegraph-app",
						Destination: &manifestFlags.bucket,
						Usage:       "Bucket where the updated manifest should be uploaded to once updated.",
					},
					&cli.IntFlag{
						Name:        "build",
						Value:       -1,
						Destination: &manifestFlags.build,
						Usage:       "Build number to retrieve the update-manifest from. If no build number is given, the latest build will be used",
						DefaultText: "latest",
					},
					&cli.StringFlag{
						Name:        "release-tag",
						Value:       "latest",
						Destination: &manifestFlags.tag,
						Usage:       "GitHub release tag which should be used to update the manifest with. If no tag is given the latest GitHub release is used",
						DefaultText: "latest",
					},
					&cli.BoolFlag{
						Name:        "update-signatures",
						Destination: &manifestFlags.updateSignatures,
						Usage:       "update the signatures in the update manifest by retrieving the signature content from the GitHub release",
					},
					&cli.BoolFlag{
						Name:        "no-upload",
						Destination: &manifestFlags.noUpload,
						Usage:       "do everything except upload the final manifest",
					},
				},
				Action: UpdateCodyAppManifest,
			},
			{
				Name:  "reset",
				Usage: "Resets the dev app's db and web cache",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "dry-run",
						Destination: &resetFlags.dryRun,
						Usage:       "write out paths to be removed",
					},
				},
				Action: ResetApp,
			},
		},
	}
)

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

func UpdateCodyAppManifest(ctx *cli.Context) error {
	client, err := bk.NewClient(ctx.Context, std.Out)
	if err != nil {
		return err
	}

	pipeline := "cody-app-release"
	branch := "app-release/stable"

	var build *buildkite.Build

	pending := std.Out.Pending(output.Line(output.EmojiHourglass, output.StylePending, "Updating manifest"))
	destroyPending := true
	defer func() {
		if destroyPending {
			pending.Destroy()
		}
	}()

	if manifestFlags.build == -1 {
		pending.Update("Retrieving latest build")
		build, err = client.GetMostRecentBuild(ctx.Context, pipeline, branch)
	} else {
		pending.Updatef(fmt.Sprintf("Retrieving build %d", manifestFlags.build))
		build, err = client.GetBuildByNumber(ctx.Context, pipeline, strconv.Itoa(manifestFlags.build))
	}
	if err != nil {
		return err
	}

	pending.Update("Looking for app.update.manifest artifact on build")
	manifestArtifact, err := findArtifactByBuild(ctx.Context, client, build, "app.update.manifest")
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	pending.Update("Downloading app.update.manifest artifact")
	err = client.DownloadArtifact(*manifestArtifact, buf)
	if err != nil {
		return err
	}

	manifest := appUpdateManifest{}
	err = json.NewDecoder(buf).Decode(&manifest)
	if err != nil {
		return err
	}

	githubClient := github.NewClient(http.DefaultClient)

	pending.Update(fmt.Sprintf("Retrieving GitHub release with tag %q", manifestFlags.tag))
	release, err := getAppGitHubRelease(ctx.Context, githubClient, manifestFlags.tag)
	if err != nil {
		return errors.Wrapf(err, "failed to get Cody App release with tag %q", manifestFlags.tag)
	}

	var updateSignatures bool
	if manifestFlags.updateSignatures {
		updateSignatures = false
	} else {
		// the tag is the version just with 'app-v' prepended
		// we only update the signatures if the tags differ
		updateSignatures = (fmt.Sprintf("app-v%s", manifest.Version) != release.GetTagName())
	}

	pending.Update(fmt.Sprintf("Updating manifest with data from the release - update sinatures: %v", updateSignatures))
	manifest, err = updateManifestFromRelease(manifest, release, updateSignatures)
	if err != nil {
		return err
	}

	destroyPending = false
	pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Manifest updated!"))

	if !manifestFlags.noUpload {
		std.Out.Writef("Please ensure you have the necassery permission requested via Entitle to upload to GCP buckets")

		std.Out.WriteNoticef("Uploading manitfest to bucket %q", manifestFlags.bucket)
		storageClient, err := storage.NewClient(ctx.Context)
		if err != nil {
			return err
		}
		storageWriter := storageClient.Bucket(manifestFlags.bucket).Object("app.update.prod.manifest.json").NewWriter(ctx.Context)
		err = json.NewEncoder(storageWriter).Encode(&manifest)
		defer func() {
			if err := storageWriter.Close(); err != nil {
				std.Out.WriteFailuref("Google Storage Writer failed on close: %v", err)
			}
		}()
		if err != nil {
			return err
		}

		if err := storageWriter.Close(); err != nil {
			return err
		}
		std.Out.WriteSuccessf("Updated manifest uploaded!")
		return nil
	}

	buf = bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent(" ", " ")
	if err := enc.Encode(&manifest); err != nil {
		return err
	} else {
		std.Out.WriteCode("json", buf.String())
	}
	return nil
}

func updateManifestFromRelease(manifest appUpdateManifest, release *github.RepositoryRelease, updateSignatures bool) (appUpdateManifest, error) {
	platformMatch := map[string]*regexp.Regexp{
		// note the regular expression will capture
		// .tar.gz
		// AND
		// .tar.gz.sig
		"aarch64-darwin": regexp.MustCompile("^Cody.*.aarch64.app.tar.gz"),
		"x86_64-darwin":  regexp.MustCompile("^Cody.*.x86_64.app.tar.gz"),
		// note the LOWERCASE cody
		"x86_64-linux": regexp.MustCompile("^cody.*_amd64.AppImage.tar.gz"),
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
		sig := appPlatform.Signature

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

	manifest.Notes = release.GetBody()

	return manifest, nil
}

func downloadSignatureContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	defer resp.Body.Close()
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
			return strings.Contains(release.GetName(), "Cody App")
		}
	} else {
		releaseCompareFn = func(release *github.RepositoryRelease) bool {
			return strings.Contains(release.GetName(), "Cody App") && release.GetTagName() == tag
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
		return nil, errors.Newf("failed to find Cody App Release tag %q", tag)
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

func ResetApp(ctx *cli.Context) error {
	if runtime.GOOS != "darwin" {
		return errors.Newf("this command is not supported on %s", runtime.GOOS)
	}
	var appDataDir, appCacheDir, appWebCacheDir, dbSocketDir string
	userHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	userCache, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dbSocketDir = filepath.Join(userHome, ".sourcegraph-psql")
	appCacheDir = filepath.Join(userCache, "sourcegraph-dev")
	appDataDir = filepath.Join(userConfig, "sourcegraph-dev")
	appWebCacheDir = filepath.Join(userHome, "Library/WebKit/Sourcegraph")

	appPaths := []string{dbSocketDir, appCacheDir, appDataDir, appWebCacheDir}
	msg := "removing"
	if resetFlags.dryRun {
		msg = "skipping"
	}
	for _, path := range appPaths {
		std.Out.Writef("%s: %s", msg, path)
		if resetFlags.dryRun {
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}
