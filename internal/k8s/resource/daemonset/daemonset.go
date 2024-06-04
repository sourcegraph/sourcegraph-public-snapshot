package daemonset

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func New(name, namespace, version string) appsv1.DaemonSet {
	return appsv1.DaemonSet{
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
		Spec: appsv1.DaemonSetSpec{
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
