package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/go-github/v55/github"
	"github.com/grafana/regexp"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

//go:embed gcs_versions.json
var gcsVersionsRaw []byte

var gcsFilenames = []string{
	"internal_database_schema.json",
	"internal_database_schema.codeintel.json",
	"internal_database_schema.codeinsights.json",
}
var githubFilenames = []string{
	"schema.json",
	"schema.codeintel.json",
	"schema.codeinsights.json",
}

var gcsVersions []semver.Version

func init() {
	if err := json.Unmarshal(gcsVersionsRaw, &gcsVersions); err != nil {
		panic("invalid JSON for gcs_versions.json: " + err.Error())
	}
}

func main() {
	ctx := context.Background()
	currentVersion := os.Args[1]
	if currentVersion == "" {
		fmt.Println("Current version argument is required.")
		fmt.Println("usage: airgappedgen vX.Y.Z <path to folder>")
		os.Exit(1)
	}
	exportPath := os.Args[2]
	if exportPath == "" {
		fmt.Println("Export path argument is required.")
		fmt.Println("usage: airgappedgen vX.Y.Z <path to folder>")
		os.Exit(1)
	}

	schemasGCS, err := downloadGCSVersions(ctx, gcsVersions)
	if err != nil {
		panic(err)
	}

	githubVersions, err := listRemoteTaggedVersions(ctx)
	if err != nil {
		panic(err)
	}

	schemasGitHub, err := downloadRemoteTaggedVersions(ctx, githubVersions)
	if err != nil {
		panic(err)
	}

	for _, sd := range append(schemasGCS, schemasGitHub...) {
		if err := sd.Export(exportPath); err != nil {
			panic(err)
		}
	}
}

func downloadRemoteTaggedVersions(ctx context.Context, versions []semver.Version) ([]*schemaDescription, error) {
	urlFmt := "https://raw.githubusercontent.com/sourcegraph/sourcegraph/v%s/internal/database/%s"

	p := pool.NewWithResults[*schemaDescription]().WithMaxGoroutines(5).WithErrors()

	for _, version := range versions {
		version := version
		p.Go(func() (*schemaDescription, error) {
			sd := schemaDescription{
				version: version,
				files:   map[string][]byte{},
			}

			for _, filename := range githubFilenames {
				url := fmt.Sprintf(urlFmt, version.String(), filename)
				resp, err := http.Get(url)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to download remote schema %q from GCS", versions[0].String())
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 500 {
					return nil, fmt.Errorf("server error downloading remote schema %q: %s", url, resp.Status)
				}
				if resp.StatusCode == 404 {
					return nil, fmt.Errorf("server error downloading remote schema %q: %s", url, resp.Status)
				}

				b, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to read response body")
				}
				sd.files[fmt.Sprintf("internal_database_%s", filename)] = b
			}
			return &sd, nil
		})
	}
	return p.Wait()
}

func downloadGCSVersions(ctx context.Context, versions []semver.Version) ([]*schemaDescription, error) {
	urlFmt := "https://storage.googleapis.com/sourcegraph-assets/migrations/drift/v%s-%s"

	p := pool.NewWithResults[*schemaDescription]().WithMaxGoroutines(5).WithErrors()

	for _, version := range versions {
		version := version
		p.Go(func() (*schemaDescription, error) {
			sd := schemaDescription{
				version: version,
				files:   map[string][]byte{},
			}

			for _, filename := range gcsFilenames {
				url := fmt.Sprintf(urlFmt, version.String(), filename)
				resp, err := http.Get(url)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to download remote schema %q from GCS", versions[0].String())
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 500 {
					return nil, fmt.Errorf("server error downloading remote schema %q: %s", url, resp.Status)
				}
				if resp.StatusCode == 404 {
					continue
				}

				b, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to read response body")
				}
				sd.files[filename] = b

			}
			return &sd, nil
		})
	}
	return p.Wait()
}

func listRemoteTaggedVersions(ctx context.Context) ([]semver.Version, error) {
	var ghc *github.Client
	if tok := os.Getenv("GH_TOKEN"); tok != "" {
		ghc = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")},
		)))
	} else {
		ghc = github.NewClient(http.DefaultClient)
	}

	// Unauthenticated requests can take a very long time if we get throttled.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	allTags := []string{}
	page := 0
	for {
		tags, resp, err := ghc.Repositories.ListTags(ctx, "sourcegraph", "sourcegraph", &github.ListOptions{Page: page})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list tags from GitHub")
		}
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage

		for _, tag := range tags {
			if isTagSourcegraphRelease(tag.GetName()) {
				allTags = append(allTags, tag.GetName())
			}
		}
	}

	allVersions := []semver.Version{}
	for _, tag := range allTags {
		v, err := semver.NewVersion(tag)
		if err != nil {
			return nil, errors.Wrapf(err, "list remote release tags: %w")
		}
		allVersions = append(allVersions, *v)
	}
	return allVersions, nil
}

var versionRegexp = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

func isTagSourcegraphRelease(tag string) bool {
	if !versionRegexp.MatchString(tag) {
		return false
	}
	return semver.MustParse(tag).GreaterThan(&gcsVersions[len(gcsVersions)-1])
}
