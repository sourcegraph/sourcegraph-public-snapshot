package deployment

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewDeployment creates a new k8s Deployment with some default values set.
func NewDeployment(name, namespace, version string) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/component": name,
				"app.kubernetes.io/name":      "sourcegraph",
				"app.kubernetes.io/version":   version,
				"deploy":                      "sourcegraph",
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
		},
	}
}
