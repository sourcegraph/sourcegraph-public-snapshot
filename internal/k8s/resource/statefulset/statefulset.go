package statefulset

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewStatefulSet creates a new k8s StatefulSet with default values.
//
// Default values include:
//
//   - MinReadySeconds: 10
//   - RevisionHistoryLimit: 10
//   - UpdateStrategy: RollingUpdateStrategy
//   - PodManagementPolicy: OrderedReadyPodManagement
//   - ServiceName: <name>
//   - Selector: "app": <name>
//   - Replicas: 1
//
// Additional options can be passed to modify the default values.
func NewStatefulSet(name, namespace string, options ...Option) (appsv1.StatefulSet, error) {
	statefulSet := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: pointers.Ptr[int32](1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			ServiceName:         name,
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			RevisionHistoryLimit: pointers.Ptr[int32](10),
			MinReadySeconds:      int32(10),
		},
	}

	// apply any given options
	for _, opt := range options {
		err := opt(&statefulSet)
		if err != nil {
			return appsv1.StatefulSet{}, err
		}
	}

	return statefulSet, nil
}

// Option sets an option for a StatefulSet.
type Option func(statefulSet *appsv1.StatefulSet) error

// WithLabels sets StatefulSet labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Labels = maps.MergePreservingExistingKeys(statefulSet.Labels, labels)
		return nil
	}
}

// WithAnnotations sets StatefulSet annotations without overriding existing annotations.
func WithAnnotations(annotations map[string]string) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Annotations = maps.MergePreservingExistingKeys(statefulSet.Annotations, annotations)
		return nil
	}
}

// WithReplicas sets the given number of Replicas for the StatefulSet.
func WithReplicas(replicas int32) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.Replicas = pointers.Ptr[int32](replicas)
		return nil
	}
}

// WithSelector sets the given Selector for the StatefulSet.
func WithSelector(selector metav1.LabelSelector) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.Selector = &selector
		return nil
	}
}

// WithPodTemplateSpec sets the given PodTemplateSpec for the StatefulSet.
func WithPodTemplateSpec(template corev1.PodTemplateSpec) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.Template = template
		return nil
	}
}

// WithVolumeClaimTemplate sets the given PersistentVolumeClaim templates for the StatefulSet.
func WithVolumeClaimTemplate(template ...corev1.PersistentVolumeClaim) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.VolumeClaimTemplates = template
		return nil
	}
}

// WithServiceName sets the given Service name for the StatefulSet.
func WithServiceName(serviceName string) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.ServiceName = serviceName
		return nil
	}
}

// WithPodManagementPolicy sets the given PodManagementPolicy for the StatefulSet.
func WithPodManagementPolicy(podManagementPolicy appsv1.PodManagementPolicyType) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.PodManagementPolicy = podManagementPolicy
		return nil
	}
}

// WithUpdateStrategy sets the given StatefulSetUpdateStrategy for the StatefulSet.
func WithUpdateStrategy(updateStrategy appsv1.StatefulSetUpdateStrategy) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.UpdateStrategy = updateStrategy
		return nil
	}
}

// WithRevisionHistoryLimit sets the given RevisionHistoryLimit for the StatefulSet.
func WithRevisionHistoryLimit(limit int32) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.RevisionHistoryLimit = pointers.Ptr[int32](limit)
		return nil
	}
}

// WithMinReadySeconds sets the minimum number of seconds for which a newly created
// Pod should be ready without any of its containers crashing, for it to be considered available.
func WithMinReadySeconds(minReadySeconds int32) Option {
	return func(statefulSet *appsv1.StatefulSet) error {
		statefulSet.Spec.MinReadySeconds = minReadySeconds
		return nil
	}
}
