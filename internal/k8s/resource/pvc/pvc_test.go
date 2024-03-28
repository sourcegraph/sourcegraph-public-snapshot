package pvc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestNewPersistentVolumeClaim(t *testing.T) {
	t.Parallel()

	type args struct {
		name      string
		namespace string
		options   []Option
	}

	tests := []struct {
		name string
		args args
		want corev1.PersistentVolumeClaim
	}{
		{
			name: "default pvc",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
			},
			want: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
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
						"app":    "foobar",
						"deploy": "horsegraph",
					}),
				},
			},
			want: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"app":    "foobar",
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
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
						"app": "horsegraph",
					}),
				},
			},
			want: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
					Annotations: map[string]string{
						"app": "horsegraph",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
				},
			},
		},
		{
			name: "with access mode",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithAccessMode([]corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteMany,
					}),
				},
			},
			want: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteMany,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
				},
			},
		},
		{
			name: "with resources",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithResources(corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("100Gi"),
						},
					}),
				},
			},
			want: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("100Gi"),
						},
					},
				},
			},
		},
		{
			name: "with storage class name",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithStorageClassName("sourcegraph"),
				},
			},
			want: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
					StorageClassName: pointers.Ptr("sourcegraph"),
				},
			},
		},
		{
			name: "with volume name",
			args: args{
				name:      "foo",
				namespace: "sourcegraph",
				options: []Option{
					WithVolumeName("testing"),
				},
			},
			want: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "sourcegraph",
					Labels: map[string]string{
						"deploy": "sourcegraph",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
					VolumeName: "testing",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewPersistentVolumeClaim(tt.args.name, tt.args.namespace, tt.args.options...)
			if err != nil {
				t.Errorf("NewPersistentVolumeClaim() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewPersistentVolumeClaim() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
