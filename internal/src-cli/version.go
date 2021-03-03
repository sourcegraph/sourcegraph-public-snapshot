package srccli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/Masterminds/semver"
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
	minimumVersion, err := semver.NewVersion(MinimumVersion)
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
func highestMatchingVersion(minimumVersion *semver.Version, versions []*semver.Version) (*semver.Version, error) {
	constraint, err := semver.NewConstraint(fmt.Sprintf("~%d.%d.x", minimumVersion.Major(), minimumVersion.Minor()))
	if err != nil {
		return nil, errors.Wrap(err, "invalid range")
	}

	var matching semver.Collection
	for _, version := range versions {
		if constraint.Check(version) {
			matching = append(matching, version)
		}
	}

	if len(matching) == 0 {
		return minimumVersion, nil
	}

	sort.Sort(matching)
	return matching[len(matching)-1], nil
}

// releaseVersions requests the given URL and all subsequent pages of
// releases. Returns the non-draft, non-prerelease items with a valid
// semver tag.
func releaseVersions(url string) ([]*semver.Version, error) {
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
func releaseVersionsPage(url string) ([]*semver.Version, string, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	respContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode >= 400 {
		return nil, "", errors.Errorf("Invalid response from GitHub: %s", respContent)
	}

	var releases []releaseMeta
	if err := json.Unmarshal(respContent, &releases); err != nil {
		return nil, "", err
	}

	versions := []*semver.Version{}
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}

		version, err := semver.NewVersion(release.TagName)
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	nextURL, _ := linkheader.ExtractNextURL(resp)
	return versions, nextURL, nil
}
