package container

import (
	"github.com/grafana/regexp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var imageRegexp = regexp.MustCompile(`(.+)/([^:]+):(.+)`)

// NewContainer creates a new k8s Container with some default values set.
func NewContainer(name string, cfg config.StandardComponent, defaults config.ContainerConfig) corev1.Container {
	ctr := corev1.Container{
		Name:                     name,
		Image:                    defaults.Image,
		ImagePullPolicy:          corev1.PullIfNotPresent,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
		Resources:                *defaults.Resources,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                pointers.Ptr[int64](100),
			RunAsGroup:               pointers.Ptr[int64](101),
			AllowPrivilegeEscalation: pointers.Ptr(false),
			ReadOnlyRootFilesystem:   pointers.Ptr(true),
		},
	}

	if cfg != nil {
		if ctrConfig, ok := cfg.GetContainerConfig()[name]; ok {
			if ctrConfig.BestEffortQOS {
				ctr.Resources = corev1.ResourceRequirements{}
			} else if ctrConfig.Resources != nil {
				ctr.Resources = *ctrConfig.Resources
			}

			if ctrConfig.Image != "" {
				ctr.Image = imageRegexp.ReplaceAllString(ctr.Image, "$1/"+ctrConfig.Image)
			}
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

func EnvVarsPostgres(secretName string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POSTGRES_DATABASE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "database",
				},
			},
		},
		{
			Name: "POSTGRES_HOST",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "host",
				},
			},
		},
		{
			Name: "POSTGRES_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "password",
				},
			},
		},
		{
			Name: "POSTGRES_PORT",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "port",
				},
			},
		},
		{
			Name: "POSTGRES_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "user",
				},
			},
		},
		{
			Name:  "POSTGRES_DB",
			Value: "$(POSTGRES_DATABASE)",
		},
	}
}

func EnvVarsPostgresExporter(secretName string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "DATA_SOURCE_DB",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "database",
				},
			},
		},
		{
			Name: "DATA_SOURCE_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "password",
				},
			},
		},
		{
			Name: "DATA_SOURCE_PORT",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "port",
				},
			},
		},
		{
			Name: "DATA_SOURCE_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "user",
				},
			},
		},
		{
			Name:  "DATA_SOURCE_URI",
			Value: "localhost:$(DATA_SOURCE_PORT)/$(DATA_SOURCE_DB)?sslmode=disable",
		},
		{
			Name:  "PG_EXPORTER_EXTEND_QUERY_PATH",
			Value: "/config/queries.yaml",
		},
	}
}
