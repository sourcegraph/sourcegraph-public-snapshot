package pod

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewPodTemplate creates a new k8s PodTemplate with some default values set.
func NewPodTemplate(name string, cfg config.StandardComponent) corev1.PodTemplate {
	template := corev1.PodTemplate{
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

	if cfg != nil {
		template.Template.Spec.Affinity = cfg.GetPodTemplateConfig().Affinity
		template.Template.Spec.ImagePullSecrets = cfg.GetPodTemplateConfig().ImagePullSecrets
		template.Template.Spec.NodeSelector = cfg.GetPodTemplateConfig().NodeSelector
		template.Template.Spec.Tolerations = cfg.GetPodTemplateConfig().Tolerations
	}

	return template
}

func NewVolumeFromPVC(name, claimName string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: claimName,
			},
		},
	}
}

func NewVolumeFromConfigMap(name, configMapName string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
				DefaultMode: pointers.Ptr[int32](0777),
			},
		},
	}
}

func NewVolumeFromSecret(name, secretName string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}

func NewVolumeHostPath(name, path string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: path,
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
