package deployment

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewDeployment creates a new k8s Deployment with default values.
//
// Default values include:
//
//   - MinReadySeconds: 10
//   - Replicas: 1
//   - RevisionHistoryLimit: 10
//   - Selector: "app": <name>
//   - DeploymentStrategy: RecreateDeploymentStrategy
//
// Additional options can be passed to modify the default values.
func NewDeployment(name, namespace string, options ...Option) (appsv1.Deployment, error) {
	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
		Spec: appsv1.DeploymentSpec{
			MinReadySeconds:      int32(10),
			Replicas:             pointers.Ptr[int32](1),
			RevisionHistoryLimit: pointers.Ptr[int32](10),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}

	// apply any given options
	for _, opt := range options {
		err := opt(&deployment)
		if err != nil {
			return appsv1.Deployment{}, err
		}
	}

	return deployment, nil
}

// Option sets an option for a Deployment.
type Option func(deployment *appsv1.Deployment) error

// WithLabels sets Deployment labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(deployment *appsv1.Deployment) error {
		deployment.Labels = maps.MergePreservingExistingKeys(deployment.Labels, labels)
		return nil
	}
}

// WithAnnotations sets Deployment annotations without overriding existing annotations.
func WithAnnotations(annotations map[string]string) Option {
	return func(deployment *appsv1.Deployment) error {
		deployment.Annotations = maps.MergePreservingExistingKeys(deployment.Annotations, annotations)
		return nil
	}
}

// WithMinReadySeconds sets the minimum number of seconds for which a newly
// created Pod should be ready without any of its containers crashing, for it to
// be considered available.
func WithMinReadySeconds(seconds int32) Option {
	return func(deployment *appsv1.Deployment) error {
		deployment.Spec.MinReadySeconds = seconds
		return nil
	}
}

// WithReplicas sets the number of replicas for the Deployment.
func WithReplicas(replicas int32) Option {
	return func(deployment *appsv1.Deployment) error {
		deployment.Spec.Replicas = pointers.Ptr[int32](replicas)
		return nil
	}
}

// WithRevisionHistoryLimit sets the revision history limit for the Deployment.
func WithRevisionHistoryLimit(revisionHistoryLimit int32) Option {
	return func(deployment *appsv1.Deployment) error {
		deployment.Spec.RevisionHistoryLimit = pointers.Ptr[int32](revisionHistoryLimit)
		return nil
	}
}

// WithSelector sets the selector for the Deployment. The selector determines which Pods are managed by the Deployment.
func WithSelector(selector metav1.LabelSelector) Option {
	return func(deployment *appsv1.Deployment) error {
		deployment.Spec.Selector = &selector
		return nil
	}
}

// WithDeploymentStrategy sets the deployment strategy for the Deployment.
// The deployment strategy determines how Pods are created/updated/deleted when the Deployment is updated.
func WithDeploymentStrategy(deploymentStrategy appsv1.DeploymentStrategy) Option {
	return func(deployment *appsv1.Deployment) error {
		deployment.Spec.Strategy = deploymentStrategy
		return nil
	}
}

// WithPodTemplateSpec sets the pod template spec for the Deployment. The pod template spec
// defines the pods that will be created by the Deployment.
func WithPodTemplateSpec(podTemplate corev1.PodTemplateSpec) Option {
	return func(deployment *appsv1.Deployment) error {
		deployment.Spec.Template = podTemplate
		return nil
	}
}
