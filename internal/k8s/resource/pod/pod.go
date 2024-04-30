package pod

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewPodTemplate creates a new k8s PodTemplate with some default values set.
func NewPodTemplate(name string) corev1.PodTemplate {
	return corev1.PodTemplate{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Annotations: map[string]string{
					"kubectl.kubernetes.io/default-container": name,
				},
				Labels: map[string]string{
					"app":    name,
					"deploy": "sourcegraph",
				},
			},
			Spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{
					RunAsUser:           pointers.Ptr[int64](100),
					RunAsGroup:          pointers.Ptr[int64](101),
					FSGroup:             pointers.Ptr[int64](101),
					FSGroupChangePolicy: pointers.Ptr(corev1.FSGroupChangeOnRootMismatch),
				},
			},
		},
	}
}

func NewVolumeEmptyDir(name string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}
