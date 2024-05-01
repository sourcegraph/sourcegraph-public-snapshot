package config

import corev1 "k8s.io/api/core/v1"

type StandardComponent interface {
	Disableable
	GetPodTemplateConfig() PodTemplateConfig
	GetResources() map[string]corev1.ResourceRequirements
	GetServiceAccountAnnotations() map[string]string
	GetPrometheusPort() *int
}

type Disableable interface {
	IsDisabled() bool
}

type StandardConfig struct {
	Disabled                  bool                                   `json:"disabled,omitempty"`
	PodTemplateConfig         PodTemplateConfig                      `json:"podTemplateConfig,omitempty"`
	PrometheusPort            *int                                   `json:"prometheusPort,omitempty"`
	Resources                 map[string]corev1.ResourceRequirements `json:"resources,omitempty"`
	ServiceAccountAnnotations map[string]string                      `json:"serviceAccountAnnotations,omitempty"`
}

// Config that applies to all Pod templates produced by a Service. If this needs
// to differ between pod templates, split another service definition.
type PodTemplateConfig struct {
	Affinity         *corev1.Affinity              `json:"affinity,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	NodeSelector     map[string]string             `json:"nodeSelector,omitempty"`
	Tolerations      []corev1.Toleration           `json:"tolerations,omitempty"`
}

func (c StandardConfig) IsDisabled() bool                                     { return c.Disabled }
func (c StandardConfig) GetPodTemplateConfig() PodTemplateConfig              { return c.PodTemplateConfig }
func (c StandardConfig) GetPrometheusPort() *int                              { return c.PrometheusPort }
func (c StandardConfig) GetResources() map[string]corev1.ResourceRequirements { return c.Resources }
func (c StandardConfig) GetServiceAccountAnnotations() map[string]string {
	return c.ServiceAccountAnnotations
}
