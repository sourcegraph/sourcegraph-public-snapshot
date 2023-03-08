package dockertools

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var tt = []struct {
	In   string
	Want ImageReference
}{
	{
		In: "sourcegraph/executor",
		Want: ImageReference{
			Namespace: "sourcegraph",
			Name:      "executor",
		},
	},
	{
		In: "sourcegraph/executor:insiders",
		Want: ImageReference{
			Namespace: "sourcegraph",
			Name:      "executor",
			Tag:       "insiders",
		},
	},
	{
		In: "index.docker.io/sourcegraph/executor:insiders",
		Want: ImageReference{
			Namespace: "sourcegraph",
			Registry:  "index.docker.io",
			Name:      "executor",
			Tag:       "insiders",
		},
	},
	{
		In: "index.docker.io/sourcegraph/executor:insiders@sha256:abc123",
		Want: ImageReference{
			Registry:  "index.docker.io", // index.docker.io strings are replaced by docker.io
			Namespace: "sourcegraph",
			Name:      "executor",
			Tag:       "insiders",
			Digest:    "sha256:abc123",
		},
	},
	{
		In: "us-central1-docker.pkg.dev/project-id/repo/sourcegraph/executor",
		Want: ImageReference{
			Registry:  "us-central1-docker.pkg.dev/project-id/repo",
			Namespace: "sourcegraph",
			Name:      "executor",
		},
	},
	{
		In: "us-central1-docker.pkg.dev/project-id/repo/sourcegraph/executor:insiders",
		Want: ImageReference{
			Registry:  "us-central1-docker.pkg.dev/project-id/repo",
			Namespace: "sourcegraph",
			Name:      "executor",
			Tag:       "insiders",
		},
	},
	{
		In: "us-central1-docker.pkg.dev/project-id/repo/sourcegraph/executor:insiders@sha256:abc123",
		Want: ImageReference{
			Registry:  "us-central1-docker.pkg.dev/project-id/repo",
			Namespace: "sourcegraph",
			Name:      "executor",
			Tag:       "insiders",
			Digest:    "sha256:abc123",
		},
	},
	{
		In: "us.gcr.io/sourcegraph-dev/searcher:insiders",
		Want: ImageReference{
			Registry:  "us.gcr.io",
			Namespace: "sourcegraph-dev",
			Name:      "searcher",
			Tag:       "insiders",
		},
	},
	{
		In: "nodejs:16",
		Want: ImageReference{
			Name: "nodejs",
			Tag:  "16",
		},
	},
	{
		In: "arm64/nodejs:16",
		Want: ImageReference{
			Namespace: "arm64",
			Name:      "nodejs",
			Tag:       "16",
		},
	},
	{
		In: "privateartifactory.sgdev.org/nodejs/nodejs:16",
		Want: ImageReference{
			Registry:  "privateartifactory.sgdev.org",
			Namespace: "nodejs",
			Name:      "nodejs",
			Tag:       "16",
		},
	},
	{
		In: "registry.example.com:5000/sourcegraph/executors:insiders",
		Want: ImageReference{
			Registry:  "registry.example.com:5000",
			Namespace: "sourcegraph",
			Name:      "executors",
			Tag:       "insiders",
		},
	},
	// https://docs.gitlab.com/ee/user/packages/container_registry/#naming-convention-for-your-container-images
	{
		In: "registry.example.com/mynamespace/myproject:some-tag",
		Want: ImageReference{
			Registry:  "registry.example.com",
			Namespace: "mynamespace",
			Name:      "myproject",
			Tag:       "some-tag",
		},
	},
	{
		In: "registry.example.com/mynamespace/myproject/image:latest",
		Want: ImageReference{
			Registry:  "registry.example.com/mynamespace",
			Namespace: "myproject",
			Name:      "image",
			Tag:       "latest",
		},
	},
	{
		In: "registry.example.com/mynamespace/myproject/my/image:rc1",
		Want: ImageReference{
			Registry:  "registry.example.com/mynamespace/myproject",
			Namespace: "my",
			Name:      "image",
			Tag:       "rc1",
		},
	},
}

func TestParseImageString(t *testing.T) {
	for idx, tc := range tt {
		name := fmt.Sprintf("%v_input=%s", idx, tc.In)
		t.Run(name, func(t *testing.T) {
			// string -> ImageReference
			got := ParseImageString(tc.In)
			if !cmp.Equal(tc.Want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.Want, got))
			}

			// ImageReference -> string
			s := got.String()
			if s != tc.In {
				t.Errorf("Expected: %s, got: %s\n", tc.In, s)
			}

		})
	}
}

func TestIsPublicDockerHub(t *testing.T) {
	testCases := []struct {
		Name      string
		Container string
		Public    bool
	}{
		{
			Name:      "NoRegistry",
			Container: "sourcegraph/executor",
			Public:    true,
		},
		{
			Name:      "LegacyDockerHub",
			Container: "index.docker.io/sourcegraph/executor:insiders",
			Public:    true,
		},
		{
			Name:      "DockerHub",
			Container: "docker.io/sourcegraph/executor:insiders",
			Public:    true,
		},
		{
			Name:      "ArtifactRegistry",
			Container: "us-central1-docker.pkg.dev/project-id/repo-name/sourcegraph/executor:insiders",
			Public:    false,
		},
		{
			Name:      "GitLab",
			Container: "registry.example.com/mynamespace/myproject/my/image:rc1",
			Public:    false,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			i := ParseImageString(test.Container)
			if i.IsPublicDockerHub() != test.Public {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(test.Public, i.IsPublicDockerHub()))
			}
		})
	}

}
