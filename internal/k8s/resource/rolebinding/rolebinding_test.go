package rolebinding

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewRoleBinding(t *testing.T) {
	t.Parallel()

	type args struct {
		name      string
		namespace string
		options   []Option
	}

	tests := []struct {
		name string
		args args
		want rbacv1.RoleBinding
	}{
		{
			name: "sourcegraph",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
			},
			want: rbacv1.RoleBinding{
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
			want: rbacv1.RoleBinding{
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
			want: rbacv1.RoleBinding{
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
			name: "with role ref",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithRoleRef(rbacv1.RoleRef{
						APIGroup: "rbac.authorization.k8s.io",
						Kind:     "Role",
						Name:     "foorole",
					}),
				},
			},
			want: rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     "foorole",
				},
			},
		},
		{
			name: "with subjects",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithSubjects([]rbacv1.Subject{
						{
							Kind:     "User",
							Name:     "foouser",
							APIGroup: "rbac.authorization.k8s.io",
						},
					}),
				},
			},
			want: rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     "User",
						Name:     "foouser",
						APIGroup: "rbac.authorization.k8s.io",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewRoleBinding(tt.args.name, tt.args.namespace, tt.args.options...)
			if err != nil {
				t.Errorf("NewRoleBinding() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewRoleBinding() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
