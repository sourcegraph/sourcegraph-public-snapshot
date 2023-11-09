package iam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestExtractImageGoogleProject(t *testing.T) {
	for _, tc := range []struct {
		name        string
		image       string
		wantProject *string
	}{
		{
			name:        "public image",
			image:       "index.docker.io/sourcegraph/telemetry-gateway",
			wantProject: nil,
		},
		{
			name:        "gcr",
			image:       "us.gcr.io/sourcegraph-dev/abuse-banbot",
			wantProject: pointers.Ptr("sourcegraph-dev"),
		},
		{
			name:        "GCP docker",
			image:       "us-central1-docker.pkg.dev/control-plane-5e9ee072/docker/apiserver",
			wantProject: pointers.Ptr("control-plane-5e9ee072"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := extractImageGoogleProject(tc.image)
			if tc.wantProject == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, *tc.wantProject, *got)
			}
		})
	}
}
