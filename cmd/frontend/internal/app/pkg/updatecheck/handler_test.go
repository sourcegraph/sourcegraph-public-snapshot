package updatecheck

import (
	"fmt"
	"net/http"
	"testing"
	"time"
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

func TestCanUpdate(t *testing.T) {
	tests := []struct {
		name                string
		now                 time.Time
		clientVersionString string
		latestReleaseBuild  build
		hasUpdate           bool
		err                 error
	}{
		{
			name:                "no version update",
			clientVersionString: "v1.2.3",
			latestReleaseBuild:  newBuild("1.2.3"),
			hasUpdate:           false,
		},
		{
			name:                "version update",
			clientVersionString: "v1.2.3",
			latestReleaseBuild:  newBuild("1.2.4"),
			hasUpdate:           true,
		},
		{
			name:                "no date update clock skew",
			now:                 time.Date(2018, time.August, 1, 0, 0, 0, 0, time.UTC),
			clientVersionString: "19272_2018-08-02_f7dec47",
			latestReleaseBuild:  newBuild("1.2.3"),
			hasUpdate:           false,
		},
		{
			name:                "no date update",
			now:                 time.Date(2018, time.September, 1, 0, 0, 0, 0, time.UTC),
			clientVersionString: "19272_2018-08-01_f7dec47",
			latestReleaseBuild:  newBuild("1.2.3"),
			hasUpdate:           false,
		},
		{
			name:                "date update",
			now:                 time.Date(2018, time.August, 42, 0, 0, 0, 0, time.UTC),
			clientVersionString: "19272_2018-08-01_f7dec47",
			latestReleaseBuild:  newBuild("1.2.3"),
			hasUpdate:           true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Mock the current time for this test.
			timeNow = func() time.Time {
				return test.now
			}
			// Restore the real time after this test is done.
			defer func() {
				timeNow = time.Now
			}()

			hasUpdate, err := canUpdate(test.clientVersionString, test.latestReleaseBuild)
			if err != test.err {
				t.Fatalf("expected error %s; got %s", test.err, err)
			}
			if hasUpdate != test.hasUpdate {
				t.Fatalf("expected hasUpdate=%t; got hasUpdate=%t", test.hasUpdate, hasUpdate)
			}
		})
	}
}
