package deployment

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestNewDeployment(t *testing.T) {
	t.Parallel()

	type args struct {
		name      string
		namespace string
		version   string
		options   []Option
	}

	tests := []struct {
		name string
		args args
		want appsv1.Deployment
	}{
		{
			name: "default deployment",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"deploy":                      "sourcegraph",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(10),
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{},
				},
			},
		},
		{
			name: "with labels",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
				options: []Option{
					WithLabels(map[string]string{
						"deploy": "horsegraph",
						"app":    "horsegraph",
					}),
				},
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"app":                         "horsegraph",
						"deploy":                      "sourcegraph",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(10),
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{},
				},
			},
		},
		{
			name: "with annotations",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
				options: []Option{
					WithAnnotations(map[string]string{
						"app": "horsegraph",
						"foo": "bar",
					}),
				},
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"deploy":                      "sourcegraph",
					},
					Annotations: map[string]string{
						"app": "horsegraph",
						"foo": "bar",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(10),
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{},
				},
			},
		},
		{
			name: "with minreadyseconds",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
				options: []Option{
					WithMinReadySeconds(int32(20)),
				},
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"deploy":                      "sourcegraph",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(20),
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{},
				},
			},
		},
		{
			name: "with replicas",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
				options: []Option{
					WithReplicas(int32(10)),
				},
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"deploy":                      "sourcegraph",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(10),
					Replicas:             pointers.Ptr[int32](10),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{},
				},
			},
		},
		{
			name: "with revision history limit",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
				options: []Option{
					WithRevisionHistoryLimit(int32(100)),
				},
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"deploy":                      "sourcegraph",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(10),
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](100),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{},
				},
			},
		},
		{
			name: "with selector",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
				options: []Option{
					WithSelector(metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "bar",
						},
					}),
				},
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"deploy":                      "sourcegraph",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(10),
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "bar",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{},
				},
			},
		},
		{
			name: "with deployment strategy",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
				options: []Option{
					WithDeploymentStrategy(appsv1.DeploymentStrategy{
						Type: appsv1.RollingUpdateDeploymentStrategyType,
					}),
				},
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"deploy":                      "sourcegraph",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(10),
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RollingUpdateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{},
				},
			},
		},
		{
			name: "with pod template spec",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				version:   "1.2.3",
				options: []Option{
					WithPodTemplateSpec(func() corev1.PodTemplateSpec {
						ts, _ := pod.NewPodTemplate("foo", "sourcegraph")
						return ts.Template
					}()),
				},
			},
			want: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app.kubernetes.io/component": "foo",
						"app.kubernetes.io/name":      "sourcegraph",
						"app.kubernetes.io/version":   "1.2.3",
						"deploy":                      "sourcegraph",
					},
				},
				Spec: appsv1.DeploymentSpec{
					MinReadySeconds:      int32(10),
					Replicas:             pointers.Ptr[int32](1),
					RevisionHistoryLimit: pointers.Ptr[int32](10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "sourcegraph",
							Annotations: map[string]string{
								"kubectl.kubernetes.io/default-container": "foo",
							},
							Labels: map[string]string{
								"app":    "foo",
								"deploy": "sourcegraph",
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewDeployment(tt.args.name, tt.args.namespace, tt.args.version, tt.args.options...)
			if err != nil {
				t.Errorf("NewDeployment() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewDeployment() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
