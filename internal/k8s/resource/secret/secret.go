package secret

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewSecret creates a new k8s Secret with some default values set.
func NewSecret(name, namespace, version string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/component": name,
				"app.kubernetes.io/name":      "sourcegraph",
				"app.kubernetes.io/version":   version,
				"deploy":                      "sourcegraph",
			},
		},
		Type: corev1.SecretTypeOpaque,
	}
}
