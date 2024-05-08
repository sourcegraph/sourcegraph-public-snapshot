package secret

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewSecret creates a new k8s Secret with some default values set.
func NewSecret(name, namespace string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
		Immutable: pointers.Ptr(false),
		Type:      corev1.SecretTypeOpaque,
	}
}
