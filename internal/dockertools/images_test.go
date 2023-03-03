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
