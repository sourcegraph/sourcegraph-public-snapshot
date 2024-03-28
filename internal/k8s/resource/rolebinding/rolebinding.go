package rolebinding

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
)

// NewRoleBinding creates a new k8s RoleBinding with default values.
//
// Default values include:
//
//   - Labels common for Sourcegraph deployments.
//
// Additional options can be passed to modify the default values.
func NewRoleBinding(name, namespace string, options ...Option) (rbacv1.RoleBinding, error) {
	roleBinding := rbacv1.RoleBinding{
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
		err := opt(&roleBinding)
		if err != nil {
			return rbacv1.RoleBinding{}, err
		}
	}

	return roleBinding, nil
}

// Option sets an option for a RoleBinding.
type Option func(rolebinding *rbacv1.RoleBinding) error

// WithLabels sets RoleBinding labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(rolebinding *rbacv1.RoleBinding) error {
		rolebinding.Labels = maps.MergePreservingExistingKeys(rolebinding.Labels, labels)
		return nil
	}
}

// WithAnnotations sets RoleBinding annotations without overriding existing labels.
func WithAnnotations(annotations map[string]string) Option {
	return func(rolebinding *rbacv1.RoleBinding) error {
		rolebinding.Annotations = maps.MergePreservingExistingKeys(rolebinding.Annotations, annotations)
		return nil
	}
}

// WithRoleRef sets the given RoleRef for the RoleBinding.
func WithRoleRef(roleRef rbacv1.RoleRef) Option {
	return func(rolebinding *rbacv1.RoleBinding) error {
		rolebinding.RoleRef = roleRef
		return nil
	}
}

// WithSubjects sets the given Subjects for the RoleBinding.
func WithSubjects(subjects []rbacv1.Subject) Option {
	return func(rolebinding *rbacv1.RoleBinding) error {
		rolebinding.Subjects = subjects
		return nil
	}
}
