package role

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewRole(t *testing.T) {
	t.Parallel()

	type args struct {
		name      string
		namespace string
		options   []Option
	}

	tests := []struct {
		name string
		args args
		want rbacv1.Role
	}{
		{
			name: "default role",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
			},
			want: rbacv1.Role{
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
			want: rbacv1.Role{
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
			want: rbacv1.Role{
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
			name: "with rules",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithRules([]rbacv1.PolicyRule{
						{
							Verbs:     []string{"get", "list", "watch"},
							APIGroups: []string{"core"},
							Resources: []string{"pods"},
						},
					}),
				},
			},
			want: rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "watch"},
						APIGroups: []string{"core"},
						Resources: []string{"pods"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewRole(tt.args.name, tt.args.namespace, tt.args.options...)
			if err != nil {
				t.Errorf("NewRole() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewRole() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
