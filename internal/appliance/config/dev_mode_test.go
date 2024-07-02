package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestSetLocalDevMode_CreatesContainerConfigWhereNoneExistedBefore(t *testing.T) {
	sg := &Sourcegraph{}

	sg.SetLocalDevMode()
	assert.Equal(t, sg.Spec.Blobstore.ContainerConfig, map[string]ContainerConfig{
		"blobstore": {BestEffortQOS: true},
	})
}

func TestSetLocalDevMode_PreservesContainerConfigOtherThanBestEffortQOS(t *testing.T) {
	sg := &Sourcegraph{
		Spec: SourcegraphSpec{
			Blobstore: BlobstoreSpec{
				StandardConfig: StandardConfig{
					ContainerConfig: map[string]ContainerConfig{
						"blobstore": {
							EnvVars: map[string]string{
								"foo": "bar",
							},
							Resources: &corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:              resource.MustParse("2"),
									corev1.ResourceMemory:           resource.MustParse("2G"),
									corev1.ResourceEphemeralStorage: resource.MustParse("4Gi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:              resource.MustParse("2"),
									corev1.ResourceMemory:           resource.MustParse("4G"),
									corev1.ResourceEphemeralStorage: resource.MustParse("8Gi"),
								},
							},
						},
					},
				},
			},
		},
	}

	sg.SetLocalDevMode()
	assert.Equal(t, sg.Spec.Blobstore.ContainerConfig, map[string]ContainerConfig{
		"blobstore": {
			BestEffortQOS: true,
			EnvVars: map[string]string{
				"foo": "bar",
			},

			// We really do want to preserve the resources block, if its already
			// set, for one reason: in case the admin has overriden
			// ephemeral-resources. Dev mode only affects CPU and memory, but
			// this is implemented in
			// internal/k8s/resource/container/container.go, not here. We set
			// the BestEffortQOS flag to influence the behavior in that file.
			// While this is admittedly a bit of a confusing leaky abstraction,
			// the golden tests should catch any regressions in the interactions
			// between these 2 bits of code.
			Resources: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("2"),
					corev1.ResourceMemory:           resource.MustParse("2G"),
					corev1.ResourceEphemeralStorage: resource.MustParse("4Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("2"),
					corev1.ResourceMemory:           resource.MustParse("4G"),
					corev1.ResourceEphemeralStorage: resource.MustParse("8Gi"),
				},
			},
		},
	})
}
