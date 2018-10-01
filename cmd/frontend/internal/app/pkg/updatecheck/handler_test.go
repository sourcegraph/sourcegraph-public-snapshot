package updatecheck

import (
	"fmt"
	"net/http"
	"testing"
)

func TestLatestServerVersionPushed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping due to network request against dockerhub")
	}

	url := fmt.Sprintf("https://index.docker.io/v1/repositories/sourcegraph/server/tags/%s", latestReleaseServerBuild.Version)
	resp, err := http.Get(url)
	if err != nil {
		t.Skip("Failed to contact dockerhub", err)
	}
	if resp.StatusCode == 404 {
		t.Fatalf("sourcegraph/server:%s does not exist on dockerhub. %s", latestReleaseServerBuild.Version, url)
	}
	if resp.StatusCode != 200 {
		t.Skip("unexpected response from dockerhub", resp.StatusCode)
	}
}

func TestLatestDataCenterVersionPushed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping due to network request")
	}
	url := fmt.Sprintf("https://github.com/sourcegraph/deploy-sourcegraph/releases/tag/v%v", latestReleaseDataCenterBuild.Version)
	resp, err := http.Head(url)
	if err != nil || resp.StatusCode != 200 {
		t.Errorf("Could not find Data Center release %s on GitHub. Respnose code %s from %s, err: %v", latestReleaseDataCenterBuild.Version, resp.Status, url, err)
	}
}
