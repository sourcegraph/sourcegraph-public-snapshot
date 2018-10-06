package updatecheck

import (
	"fmt"
	"net/http"
	"testing"
)

func TestLatestDockerVersionPushed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping due to network request against dockerhub")
	}

	url := fmt.Sprintf("https://index.docker.io/v1/repositories/sourcegraph/server/tags/%s", latestReleaseDockerServerImageBuild.Version)
	resp, err := http.Get(url)
	if err != nil {
		t.Skip("Failed to contact dockerhub", err)
	}
	if resp.StatusCode == 404 {
		t.Fatalf("sourcegraph/server:%s does not exist on dockerhub. %s", latestReleaseDockerServerImageBuild.Version, url)
	}
	if resp.StatusCode != 200 {
		t.Skip("unexpected response from dockerhub", resp.StatusCode)
	}
}

func TestLatestKubernetesVersionPushed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping due to network request")
	}
	url := fmt.Sprintf("https://github.com/sourcegraph/deploy-sourcegraph/releases/tag/v%v", latestReleaseKubernetesBuild.Version)
	resp, err := http.Head(url)
	if err != nil || resp.StatusCode != 200 {
		t.Errorf("Could not find Kubernetes release %s on GitHub. Response code %s from %s, err: %v", latestReleaseKubernetesBuild.Version, resp.Status, url, err)
	}
}
