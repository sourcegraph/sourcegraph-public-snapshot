package service

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
)

// NewService creates a new k8s Service with default values.
//
// Default values include:
//   - Selector based on the provided service name.
//   - Service type of Cluster IP.
//
// Additional options can be passed to modify the default values.
func NewService(name, namespace string, options ...Option) (corev1.Service, error) {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":    name,
				"deploy": "sourcegraph",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// apply any given options
	for _, opt := range options {
		err := opt(&service)
		if err != nil {
			return corev1.Service{}, err
		}
	}

	return service, nil
}

// Option sets an option for a Service.
type Option func(service *corev1.Service) error

// WithLabels sets Service labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(service *corev1.Service) error {
		service.Labels = maps.MergePreservingExistingKeys(service.Labels, labels)
		return nil
	}
}

// WithAnnotations sets Service annotations without overriding existing annotations.
func WithAnnotations(annotations map[string]string) Option {
	return func(service *corev1.Service) error {
		service.Annotations = maps.MergePreservingExistingKeys(service.Annotations, annotations)
		return nil
	}
}

// WithPorts adds the provided ServicePorts to the Service if they do not already exist.
// It also sorts the Ports by name for accurate comparison of resources.
func WithPorts(ports ...corev1.ServicePort) Option {
	return func(service *corev1.Service) error {
		for _, p := range ports {
			if !portExists(p.Name, service) {
				service.Spec.Ports = append(service.Spec.Ports, p)
			}
		}

		// sort ports by name
		sort.SliceStable(service.Spec.Ports, func(i, j int) bool {
			return service.Spec.Ports[i].Name < service.Spec.Ports[j].Name
		})
		return nil
	}
}

func portExists(name string, service *corev1.Service) bool {
	for _, p := range service.Spec.Ports {
		if p.Name == name {
			return true
		}
	}
	return false
}

// WithSelector sets the given selector as the Selector on the Service.
func WithSelector(selector map[string]string) Option {
	return func(service *corev1.Service) error {
		service.Spec.Selector = selector
		return nil
	}
}

// WithServiceType sets the given serviceType on the Service.
func WithServiceType(serviceType corev1.ServiceType) Option {
	return func(service *corev1.Service) error {
		service.Spec.Type = serviceType
		return nil
	}
}

// WithClusterIP sets the Service cluster IP manually.
func WithClusterIP(clusterIP string) Option {
	return func(service *corev1.Service) error {
		service.Spec.ClusterIP = clusterIP
		return nil
	}
}
