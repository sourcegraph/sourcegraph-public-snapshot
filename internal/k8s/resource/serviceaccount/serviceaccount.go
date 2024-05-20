package serviceaccount

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
)

// NewServiceAccount creates a new k8s ServiceAccount with some default values
// set.
func NewServiceAccount(name, namespace string, cfg config.StandardComponent) corev1.ServiceAccount {
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
	}

	if cfg != nil {
		sa.SetAnnotations(cfg.GetServiceAccountAnnotations())
	}

	return sa
}
