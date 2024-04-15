package service

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	type args struct {
		name      string
		namespace string
		options   []Option
	}

	tests := []struct {
		name string
		args args
		want corev1.Service
	}{
		{
			name: "default service",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app":    "foo",
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "foo",
					},
					Type: corev1.ServiceTypeClusterIP,
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
						"app":         "bar",
						"deploy":      "horsegraph",
						"environment": "testing",
					}),
				},
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app":         "foo",
						"deploy":      "sourcegraph",
						"environment": "testing",
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "foo",
					},
					Type: corev1.ServiceTypeClusterIP,
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
						"foo":  "bar",
						"type": "test",
					}),
				},
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app":    "foo",
						"deploy": "sourcegraph",
					},
					Annotations: map[string]string{
						"foo":  "bar",
						"type": "test",
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "foo",
					},
					Type: corev1.ServiceTypeClusterIP,
				},
			},
		},
		{
			name: "with ports",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithPorts([]corev1.ServicePort{
						{
							Name:     "http",
							Protocol: "TCP",
							Port:     8080,
							TargetPort: intstr.IntOrString{
								Type:   0,
								IntVal: 8080,
							},
							NodePort: 8080,
						},
						{
							Name:     "app",
							Protocol: "TCP",
							Port:     400,
							TargetPort: intstr.IntOrString{
								Type:   0,
								IntVal: 400,
							},
							NodePort: 400,
						},
					}...),
				},
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app":    "foo",
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "foo",
					},
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Name:     "app",
							Protocol: "TCP",
							Port:     400,
							TargetPort: intstr.IntOrString{
								Type:   0,
								IntVal: 400,
							},
							NodePort: 400,
						},
						{
							Name:     "http",
							Protocol: "TCP",
							Port:     8080,
							TargetPort: intstr.IntOrString{
								Type:   0,
								IntVal: 8080,
							},
							NodePort: 8080,
						},
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
					WithSelector(map[string]string{
						"app": "horsegraph",
					}),
				},
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app":    "foo",
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "horsegraph",
					},
					Type: corev1.ServiceTypeClusterIP,
				},
			},
		},
		{
			name: "with service type",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithServiceType(corev1.ServiceTypeLoadBalancer),
				},
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app":    "foo",
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "foo",
					},
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewService(tt.args.name, tt.args.namespace, tt.args.options...)
			if err != nil {
				t.Errorf("NewContainer() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewService() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
