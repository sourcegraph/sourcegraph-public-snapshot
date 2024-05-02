package container

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewContainer creates a new k8s Container with some default values set.
func NewContainer(name string, cfg config.StandardComponent, defaultResources corev1.ResourceRequirements) corev1.Container {
	ctr := corev1.Container{
		Name:                     name,
		ImagePullPolicy:          corev1.PullIfNotPresent,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
		Resources:                defaultResources,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                pointers.Ptr[int64](100),
			RunAsGroup:               pointers.Ptr[int64](101),
			AllowPrivilegeEscalation: pointers.Ptr(false),
			ReadOnlyRootFilesystem:   pointers.Ptr(true),
		},
	}

	if cfg != nil {
		if ctrConfig, ok := cfg.GetContainerConfig()[name]; ok {
			ctr.Resources = ctrConfig.Resources
		}
	}

	return ctr
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

func NewEnvVarSecretKeyRef(name, secretName, secretKey string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
}

func NewEnvVarFieldRef(name, fieldPath string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: fieldPath,
			},
		},
	}
}

func EnvVarsRedis() []corev1.EnvVar {
	return []corev1.EnvVar{
		NewEnvVarSecretKeyRef("REDIS_CACHE_ENDPOINT", "redis-cache", "endpoint"),
		NewEnvVarSecretKeyRef("REDIS_STORE_ENDPOINT", "redis-store", "endpoint"),
	}
}

func EnvVarsOtel() []corev1.EnvVar {
	return []corev1.EnvVar{
		// OTEL_AGENT_HOST must be defined before OTEL_EXPORTER_OTLP_ENDPOINT to substitute the node IP on which the DaemonSet pod instance runs in the latter variable
		NewEnvVarFieldRef("OTEL_AGENT_HOST", "status.hostIP"),
		{Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: "http://$(OTEL_AGENT_HOST):4317"},
	}
}
