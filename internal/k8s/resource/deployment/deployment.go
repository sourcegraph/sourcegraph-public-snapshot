package deployment

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewDeployment creates a new k8s Deployment with some default values set.
func NewDeployment(name, namespace, version string) appsv1.Deployment {
	// Note that we don't set a default spec.replicas because the default is 1
	// anyway, and if this field is explicitly set but also managed by an HPA,
	// configuration changes will cause downscale events. Even when the HPA
	// subsequently scales back out, there may have been a period of service
	// disruption as requests saturate the single replica.
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
			RevisionHistoryLimit: pointers.Ptr[int32](10),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
		},
	}
}
