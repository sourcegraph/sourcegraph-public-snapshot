package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLocalDevMode_CreatesContainerConfigWhereNoneExistedBefore(t *testing.T) {
	sg := Sourcegraph{}

	SetLocalDevMode(&sg)
	assert.Equal(t, sg.Spec.Blobstore.ContainerConfig, map[string]ContainerConfig{
		"blobstore": {BestEffortQOS: true},
	})
}

func TestSetLocalDevMode_PreservesContainerConfigOtherThanBestEffortQOS(t *testing.T) {
	sg := Sourcegraph{
		Spec: SourcegraphSpec{
			Blobstore: BlobstoreSpec{
				StandardConfig: StandardConfig{
					ContainerConfig: map[string]ContainerConfig{
						"blobstore": {
							EnvVars: map[string]string{
								"foo": "bar",
							},
						},
					},
				},
			},
		},
	}

	SetLocalDevMode(&sg)
	assert.Equal(t, sg.Spec.Blobstore.ContainerConfig, map[string]ContainerConfig{
		"blobstore": {
			BestEffortQOS: true,
			EnvVars: map[string]string{
				"foo": "bar",
			},
		},
	})
}
