package secret

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestNewSecret(t *testing.T) {
	t.Parallel()

	type args struct {
		name      string
		namespace string
		options   []Option
	}

	tests := []struct {
		name string
		args args
		want corev1.Secret
	}{
		{
			name: "default secret",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
			},
			want: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Immutable: pointers.Ptr(false),
				Type:      corev1.SecretTypeOpaque,
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
			want: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
						"foo":    "bar",
					},
				},
				Immutable: pointers.Ptr(false),
				Type:      corev1.SecretTypeOpaque,
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
			want: corev1.Secret{
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
				Immutable: pointers.Ptr(false),
				Type:      corev1.SecretTypeOpaque,
			},
		},
		{
			name: "immutable",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					Immutable(),
				},
			},
			want: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Immutable: pointers.Ptr(true),
				Type:      corev1.SecretTypeOpaque,
			},
		},
		{
			name: "with data",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithData(map[string][]byte{
						"username": []byte("myuser"),
						"password": []byte("mypass"),
					}),
				},
			},
			want: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Immutable: pointers.Ptr(false),
				Data: map[string][]byte{
					"username": []byte("myuser"),
					"password": []byte("mypass"),
				},
				Type: corev1.SecretTypeOpaque,
			},
		},
		{
			name: "with string data",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithStringData(map[string]string{
						"foo": "bar",
					}),
				},
			},
			want: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Immutable: pointers.Ptr(false),
				StringData: map[string]string{
					"foo": "bar",
				},
				Type: corev1.SecretTypeOpaque,
			},
		},
		{
			name: "of type",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					OfType(corev1.SecretTypeBasicAuth),
				},
			},
			want: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Immutable: pointers.Ptr(false),
				Type:      corev1.SecretTypeBasicAuth,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewSecret(tt.args.name, tt.args.namespace, tt.args.options...)
			if err != nil {
				t.Errorf("NewPodTemplate() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewSecret() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
