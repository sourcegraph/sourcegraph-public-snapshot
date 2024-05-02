package statefulset

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewStatefulSet creates a new k8s StatefulSet with some default values set.
func NewStatefulSet(name, namespace, version string) appsv1.StatefulSet {
	return appsv1.StatefulSet{
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
}
