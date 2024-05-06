package service

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
)

// NewService creates a new k8s Service with some default values set.
func NewService(name, namespace string, cfg config.StandardComponent) corev1.Service {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":                         name,
				"app.kubernetes.io/component": name,
				"deploy":                      "sourcegraph",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	if cfg != nil {
		if promPort := cfg.GetPrometheusPort(); promPort != nil {
			annotations := map[string]string{
				"prometheus.io/port":            fmt.Sprintf("%d", *promPort),
				"sourcegraph.prometheus/scrape": "true",
			}
			svc.SetAnnotations(annotations)
		}
	}

	return svc
}
