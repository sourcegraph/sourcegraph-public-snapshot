package role

import (
	"sort"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
)

// NewRole creates a new k8s RBAC role with default values.
//
// Default values include:
//
//   - Labels common for Sourcegraph deployments.
//
// Additional options can be passed to modify the default values.
func NewRole(name, namespace string, options ...Option) (rbacv1.Role, error) {
	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
	}

	// apply any options
	for _, opt := range options {
		err := opt(&role)
		if err != nil {
			return rbacv1.Role{}, err
		}
	}

	return role, nil
}

// Option sets an option for a Role.
type Option func(role *rbacv1.Role) error

// WithLabels sets Role labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(role *rbacv1.Role) error {
		role.Labels = maps.MergePreservingExistingKeys(role.Labels, labels)
		return nil
	}
}

// WithAnnotations sets Role annotations without overriding existing annotations.
func WithAnnotations(annotations map[string]string) Option {
	return func(role *rbacv1.Role) error {
		role.Annotations = maps.MergePreservingExistingKeys(role.Annotations, annotations)
		return nil
	}
}

// WithRules sets Role Policy Rules, sorting by name for accurate comparison of resources.
func WithRules(rules []rbacv1.PolicyRule) Option {
	return func(role *rbacv1.Role) error {
		role.Rules = rules

		sort.SliceStable(role.Rules, func(i, j int) bool {
			return role.Rules[i].Verbs[0] < role.Rules[j].Verbs[0]
		})

		return nil
	}
}
