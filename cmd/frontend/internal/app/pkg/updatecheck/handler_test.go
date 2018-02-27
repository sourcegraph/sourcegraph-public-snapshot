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
	platforms := []string{
		"darwin_amd64",
		"linux_amd64",
	}
	for _, platform := range platforms {
		url := fmt.Sprintf("https://storage.googleapis.com/sourcegraph-assets/sourcegraph-server-gen/%s/%s/sourcegraph-server-gen", latestReleaseDataCenterBuild.Version, platform)
		resp, err := http.Head(url)
		if err != nil {
			t.Skip("failed to contact googleapis.com", err)
		}
		if resp.StatusCode == 403 {
			t.Errorf("sourcegraph-server-gen %s %s is not uploaded to Google Storage. %s from %s", latestReleaseDataCenterBuild.Version, platform, resp.Status, url)
			continue
		}
		if resp.StatusCode != 200 {
			t.Skip("Unexpected response from googleapis.com", resp.Status, url)
		}
	}
}
