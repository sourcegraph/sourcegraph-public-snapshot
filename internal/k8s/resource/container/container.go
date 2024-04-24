package container

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewContainer creates a new k8s Container with some default values set.
func NewContainer(name string) corev1.Container {
	return corev1.Container{
		Name:                     name,
		ImagePullPolicy:          corev1.PullIfNotPresent,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("500Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                pointers.Ptr[int64](100),
			RunAsGroup:               pointers.Ptr[int64](101),
			AllowPrivilegeEscalation: pointers.Ptr(false),
			ReadOnlyRootFilesystem:   pointers.Ptr(true),
		},
	}
}

// NewDefaultLivenessProbe creates a default LivenessProbe that is commonly used
// for Sourcegraph services.
func NewDefaultLivenessProbe(portName string) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/",
				Port:   intstr.FromString(portName),
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 60,
		TimeoutSeconds:      5,
	}
}

// NewDefaultReadinessProbe creates a default LivenessProbe that is commonly used
// for Sourcegraph services.
func NewDefaultReadinessProbe(portName string) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/",
				Port:   intstr.FromString(portName),
				Scheme: corev1.URISchemeHTTP,
			},
		},
		PeriodSeconds:  5,
		TimeoutSeconds: 5,
	}
}
