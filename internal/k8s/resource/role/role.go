package role

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewRole creates a new k8s RBAC role with some default values set.
func NewRole(name, namespace string) rbacv1.Role {
	return rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
	}

}

func NewClusterRole(name, namespace string) rbacv1.ClusterRole {
	return rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"for-namespace": namespace,
				"deploy":        "sourcegraph",
			},
		},
	}

}
