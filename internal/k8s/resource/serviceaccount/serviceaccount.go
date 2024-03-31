package serviceaccount

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
)

// NewServiceAccount creates a new k8s ServiceAccount with default values.
//
// Default values include:
//
//   - Labels common for Sourcegraph deployments.
//
// Additional options can be passed to modify the default values.
func NewServiceAccount(name, namespace string, options ...Option) (corev1.ServiceAccount, error) {
	serviceAccount := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
	}

	// apply any given options
	for _, opt := range options {
		err := opt(&serviceAccount)
		if err != nil {
			return corev1.ServiceAccount{}, err
		}
	}

	return serviceAccount, nil
}

// Option sets an option for a ServiceAccount.
type Option func(serviceAccount *corev1.ServiceAccount) error

// WithLabels sets ServiceAccount labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(serviceAccount *corev1.ServiceAccount) error {
		serviceAccount.Labels = maps.MergePreservingExistingKeys(serviceAccount.Labels, labels)
		return nil
	}
}

// WithAnnotations sets ServiceAccount annotations without overriding existing annotations.
func WithAnnotations(annotations map[string]string) Option {
	return func(serviceAccount *corev1.ServiceAccount) error {
		serviceAccount.Annotations = maps.MergePreservingExistingKeys(serviceAccount.Annotations, annotations)
		return nil
	}
}
