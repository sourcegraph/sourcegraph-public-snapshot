package statefulset

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestNewStatefulSet(t *testing.T) {
	t.Parallel()

	type args struct {
		name      string
		namespace string
		options   []Option
	}

	tests := []struct {
		name string
		args args
		want appsv1.StatefulSet
	}{
		{
			name: "default statefulset",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with labels",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithLabels(map[string]string{
						"deploy": "horsegraph",
						"app":    "bar",
					}),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app":    "bar",
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with annotations",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithAnnotations(map[string]string{
						"foo": "bar",
					}),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
					Annotations: map[string]string{
						"foo": "bar",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with replicas",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithReplicas(int32(5)),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](5),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with selector",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithSelector(metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "horsegraph",
						},
					}),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "horsegraph",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with pod template spec",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithPodTemplateSpec(func() corev1.PodTemplateSpec {
						ts, _ := pod.NewPodTemplate("foo", "sourcegraph")
						return ts.Template
					}()),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "sourcegraph",
							Labels: map[string]string{
								"app":    "foo",
								"deploy": "sourcegraph",
							},
							Annotations: map[string]string{
								"kubectl.kubernetes.io/default-container": "foo",
							},
						},
						Spec: corev1.PodSpec{
							SecurityContext: &corev1.PodSecurityContext{
								RunAsUser:           pointers.Ptr[int64](100),
								RunAsGroup:          pointers.Ptr[int64](101),
								FSGroup:             pointers.Ptr[int64](101),
								FSGroupChangePolicy: pointers.Ptr(corev1.FSGroupChangeOnRootMismatch),
							},
						},
					},
				},
			},
		},
		{
			name: "with volume claim template",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithVolumeClaimTemplate(corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name: "repos",
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							Resources: corev1.VolumeResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("200Gi"),
								},
							},
						},
					}),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "repos",
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									corev1.ReadWriteOnce,
								},
								Resources: corev1.VolumeResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("200Gi"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "with service name",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithServiceName("test"),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "test",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with pod management policy",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithPodManagementPolicy(appsv1.ParallelPodManagement),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.ParallelPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with update strategy",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithUpdateStrategy(appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.OnDeleteStatefulSetStrategyType,
					}),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.OnDeleteStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with revision history",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithRevisionHistoryLimit(int32(5)),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(10),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](5),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
		{
			name: "with minready seconds",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithMinReadySeconds(int32(12)),
				},
			},
			want: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: appsv1.StatefulSetSpec{
					MinReadySeconds:      int32(12),
					PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					ServiceName: "foo",
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewStatefulSet(tt.args.name, tt.args.namespace, tt.args.options...)
			if err != nil {
				t.Errorf("NewStatefulSets() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewStatefulSets() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
