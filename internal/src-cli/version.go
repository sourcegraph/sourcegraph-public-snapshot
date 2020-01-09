package srccli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/linkheader"
)

type releaseMeta struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

const githubAPIReleasesEndpoint = "https://api.github.com/repos/sourcegraph/src-cli/releases"

// Version returns the highest public version currently available via the GitHub release page
// that has the same major and minor versions as the configured minimum version. This allows
// us to recommend patch updates without having to release a sourcegraph instance with a bumped
// constant.
func Version() (string, error) {
	minimumVersion, err := semver.Make(MinimumVersion)
	if err != nil {
		return "", errors.Wrap(err, "non-semantic minimum src-cli version")
	}

	versions, err := releaseVersions(githubAPIReleasesEndpoint)
	if err != nil {
		return "", errors.Wrap(err, "fetching src-cli release versions")
	}

	recommendedVersion, err := highestMatchingVersion(minimumVersion, versions)
	if err != nil {
		return "", errors.Wrap(err, "comparing versions")
	}

	return recommendedVersion.String(), nil
}

// highestMatchingVersion returns the highest version with the same major and
// minor value as the given minimum version.
func highestMatchingVersion(minimumVersion semver.Version, versions []semver.Version) (semver.Version, error) {
	constraint := fmt.Sprintf(">=%d.%d.%d <%d.%d.0",
		minimumVersion.Major, minimumVersion.Minor, minimumVersion.Patch,
		minimumVersion.Major, minimumVersion.Minor+1,
	)

	checkRange, err := semver.ParseRange(constraint)
	if err != nil {
		return semver.Version{}, errors.Wrap(err, "invalid range")
	}

	var matching []semver.Version
	for _, version := range versions {
		if checkRange(version) {
			matching = append(matching, version)
		}
	}

	if len(matching) == 0 {
		return minimumVersion, nil
	}

	semver.Sort(matching)
	return matching[len(matching)-1], nil
}

// releaseVersions requests the given URL and all subsequent pages of
// releases. Returns the non-draft, non-prerelease items with a valid
// semver tag.
func releaseVersions(url string) ([]semver.Version, error) {
	versions, nextURL, err := releaseVersionsPage(url)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		tail, err := releaseVersions(nextURL)
		if err != nil {
			return nil, err
		}
		versions = append(versions, tail...)
	}

	return versions, nil
}

// releaseVersionsPage requests the given URL and returns the non-draft,
// non-prerelease items with a valid semver tag and the url for the next page
// of results (if one exists).
func releaseVersionsPage(url string) ([]semver.Version, string, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	respContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var releases []releaseMeta
	if err := json.Unmarshal(respContent, &releases); err != nil {
		return nil, "", err
	}

	versions := []semver.Version{}
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}

		version, err := semver.Make(release.TagName)
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	nextURL, _ := linkheader.ExtractNextURL(resp)
	return versions, nextURL, nil
}
