package config

import corev1 "k8s.io/api/core/v1"

type StandardComponent interface {
	Disableable
	GetResources() map[string]corev1.ResourceRequirements
	GetServiceAccountAnnotations() map[string]string
	PrometheusPort() *int
}

type Disableable interface {
	IsDisabled() bool
}

type StandardConfig struct {
	Disabled                  bool                                   `json:"disabled,omitempty"`
	Resources                 map[string]corev1.ResourceRequirements `json:"resources,omitempty"`
	ServiceAccountAnnotations map[string]string                      `json:"serviceAccountAnnotations,omitempty"`
}

func (c StandardConfig) IsDisabled() bool                                     { return c.Disabled }
func (c StandardConfig) GetResources() map[string]corev1.ResourceRequirements { return c.Resources }
func (c StandardConfig) GetServiceAccountAnnotations() map[string]string {
	return c.ServiceAccountAnnotations
}
