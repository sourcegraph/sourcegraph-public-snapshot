package secret

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewSecret creates a new k8s Secret with default values.
//
// Default values include:
//
//   - Immutable set to false.
//   - Default type of `SecretTypeOpaque`
//
// Additional options can be passed to modify the default values.
func NewSecret(name, namespace string, options ...Option) (corev1.Secret, error) {
	secret := corev1.Secret{
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

	// apply any options
	for _, opt := range options {
		err := opt(&secret)
		if err != nil {
			return corev1.Secret{}, err
		}
	}
	return secret, nil
}

// Option sets an option for a Secret.
type Option func(secret *corev1.Secret) error

// WithLabels sets Secret labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(secret *corev1.Secret) error {
		secret.Labels = maps.MergePreservingExistingKeys(secret.Labels, labels)
		return nil
	}
}

// WithAnnotations sets Secret annotations without overriding existing annotations.
func WithAnnotations(annotations map[string]string) Option {
	return func(secret *corev1.Secret) error {
		secret.Annotations = maps.MergePreservingExistingKeys(secret.Annotations, annotations)
		return nil
	}
}

// Immutable sets a secret to be immutable.
func Immutable() Option {
	return func(secret *corev1.Secret) error {
		secret.Immutable = pointers.Ptr(true)
		return nil
	}
}

// WithData sets the given binary data to a Secret.
func WithData(data map[string][]byte) Option {
	return func(secret *corev1.Secret) error {
		secret.Data = data
		return nil
	}
}

// WithStringData sets the given string data to a Secret.
func WithStringData(data map[string]string) Option {
	return func(secret *corev1.Secret) error {
		secret.StringData = data
		return nil
	}
}

// OfType sets the given SecretType to the Secret.
func OfType(secretType corev1.SecretType) Option {
	return func(secret *corev1.Secret) error {
		secret.Type = secretType
		return nil
	}
}
