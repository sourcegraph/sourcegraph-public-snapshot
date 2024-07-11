package container

import (
	"sort"

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
			ctr.Env = append(ctr.Env, newSortedEnvVars(ctrConfig.EnvVars)...)

			if ctrConfig.BestEffortQOS {
				// Preserve ephemeral-storage
				delete(ctr.Resources.Requests, corev1.ResourceCPU)
				delete(ctr.Resources.Requests, corev1.ResourceMemory)
				delete(ctr.Resources.Limits, corev1.ResourceCPU)
				delete(ctr.Resources.Limits, corev1.ResourceMemory)
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
		NewEnvVarSecretKeyRef("POSTGRES_DATABASE", secretName, "database"),
		NewEnvVarSecretKeyRef("POSTGRES_HOST", secretName, "host"),
		NewEnvVarSecretKeyRef("POSTGRES_PASSWORD", secretName, "password"),
		NewEnvVarSecretKeyRef("POSTGRES_PORT", secretName, "port"),
		NewEnvVarSecretKeyRef("POSTGRES_USER", secretName, "user"),
		{
			Name:  "POSTGRES_DB",
			Value: "$(POSTGRES_DATABASE)",
		},
	}
}

func EnvVarsPostgresExporter(secretName string) []corev1.EnvVar {
	return []corev1.EnvVar{
		NewEnvVarSecretKeyRef("DATA_SOURCE_DB", secretName, "database"),
		NewEnvVarSecretKeyRef("DATA_SOURCE_PASS", secretName, "password"),
		NewEnvVarSecretKeyRef("DATA_SOURCE_PORT", secretName, "port"),
		NewEnvVarSecretKeyRef("DATA_SOURCE_USER", secretName, "user"),
		{
			Name:  "DATA_SOURCE_URI",
			Value: "127.0.0.1:$(DATA_SOURCE_PORT)/$(DATA_SOURCE_DB)?sslmode=disable",
		},
	}
}

func newSortedEnvVars(vars map[string]string) []corev1.EnvVar {
	keys := make([]string, len(vars))
	i := 0
	for key := range vars {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	ret := make([]corev1.EnvVar, len(vars))
	for i, key := range keys {
		ret[i] = corev1.EnvVar{Name: key, Value: vars[key]}
	}
	return ret
}
