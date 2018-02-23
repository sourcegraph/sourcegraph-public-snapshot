package updatecheck

import (
	"net/http"
	"testing"
)

func TestLatestVersionPushed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping due to network request against dockerhub")
	}

	resp, err := http.Get("https://index.docker.io/v1/repositories/sourcegraph/server/tags/" + latestReleaseBuild.Version)
	if err != nil {
		t.Skip("Failed to contact dockerhub", err)
	}
	if resp.StatusCode == 404 {
		t.Fatalf("sourcegraph/server:%s does not exist on dockerhub. https://hub.docker.com/r/sourcegraph/server/tags/", latestReleaseBuild.Version)
	}
	if resp.StatusCode != 200 {
		t.Skip("unexpected response from dockerhub", resp.StatusCode)
	}
}
