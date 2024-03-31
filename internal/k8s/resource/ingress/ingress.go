package ingress

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
)

func NewIngress(name, namespace string, options ...Option) (networkingv1.Ingress, error) {
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
		Spec: networkingv1.IngressSpec{},
	}

	// apply any given options
	for _, opt := range options {
		err := opt(&ingress)
		if err != nil {
			return networkingv1.Ingress{}, err
		}
	}
	return ingress, nil
}

// Option sets an option for an Ingress.
type Option func(ingress *networkingv1.Ingress) error

// WithLabels sets ingress labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(ingress *networkingv1.Ingress) error {
		ingress.Labels = maps.MergePreservingExistingKeys(ingress.Labels, labels)
		return nil
	}
}

// WithAnnotations sets ingress annotations without overriding existing labels.
func WithAnnotations(annotations map[string]string) Option {
	return func(ingress *networkingv1.Ingress) error {
		ingress.Annotations = maps.MergePreservingExistingKeys(ingress.Annotations, annotations)
		return nil
	}
}

// WithIngressClassName sets the name of the IngressClass cluster resource.
func WithIngressClassName(ingressClassName string) Option {
	return func(ingress *networkingv1.Ingress) error {
		ingress.Spec.IngressClassName = &ingressClassName
		return nil
	}
}

// WithDefaultIngressBackend sets the default ingress backend.
func WithDefaultIngressBackend(defaultBackend networkingv1.IngressBackend) Option {
	return func(ingress *networkingv1.Ingress) error {
		ingress.Spec.DefaultBackend = &defaultBackend
		return nil
	}
}

// WithIngressTLS sets the ingress TLS settings.
func WithIngressTLS(tls []networkingv1.IngressTLS) Option {
	return func(ingress *networkingv1.Ingress) error {
		ingress.Spec.TLS = tls
		return nil
	}
}

// WithIngressRules sets the ingress rules.
func WithIngressRules(rules []networkingv1.IngressRule) Option {
	return func(ingress *networkingv1.Ingress) error {
		ingress.Spec.Rules = rules
		return nil
	}
}
