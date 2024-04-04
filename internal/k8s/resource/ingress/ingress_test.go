package ingress

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestNewIngress(t *testing.T) {
	t.Parallel()

	type args struct {
		name      string
		namespace string
		options   []Option
	}

	tests := []struct {
		name string
		args args
		want networkingv1.Ingress
	}{
		{
			name: "default ingress",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
			},
			want: networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
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
						"foo": "bar",
					}),
				},
			},
			want: networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
						"foo":    "bar",
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
			want: networkingv1.Ingress{
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
			},
		},
		{
			name: "with ingress class name",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithIngressClassName("foo"),
				},
			},
			want: networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: networkingv1.IngressSpec{
					IngressClassName: pointers.Ptr("foo"),
				},
			},
		},
		{
			name: "with ingress TLS",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithIngressTLS([]networkingv1.IngressTLS{
						{
							Hosts:      []string{"foo"},
							SecretName: "bar",
						},
					}),
				},
			},
			want: networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: networkingv1.IngressSpec{
					TLS: []networkingv1.IngressTLS{
						{
							Hosts:      []string{"foo"},
							SecretName: "bar",
						},
					},
				},
			},
		},
		{
			name: "with default ingress backend",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithDefaultIngressBackend(networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: "foo",
							Port: networkingv1.ServiceBackendPort{
								Name:   "http",
								Number: 80,
							},
						},
					}),
				},
			},
			want: networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: networkingv1.IngressSpec{
					DefaultBackend: &networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: "foo",
							Port: networkingv1.ServiceBackendPort{
								Name:   "http",
								Number: 80,
							},
						},
					},
				},
			},
		},
		{
			name: "with ingress rules",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithIngressRules([]networkingv1.IngressRule{
						{
							Host: "foo",
						},
					}),
				},
			},
			want: networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							Host: "foo",
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
			got, err := NewIngress(tt.args.name, tt.args.namespace, tt.args.options...)
			if err != nil {
				t.Errorf("NewIngress() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewIngress() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
