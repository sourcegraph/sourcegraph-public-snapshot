package hash

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHashObject(t *testing.T) {
	equal := []struct {
		name    string
		objectA any
		ObjectB any
	}{
		{
			name:    "hash nil objects",
			objectA: nil,
			ObjectB: HashObject(nil),
		},
		{
			name: "hash two pods with the same content",
			objectA: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "sourcegraph",
					Name:      "foo",
					Labels: map[string]string{
						"foo": "bar",
						"app": "horsegraph",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Env: []corev1.EnvVar{
								{
									Name:  "var1",
									Value: "value1",
								},
							},
						},
					},
				},
			},
			ObjectB: HashObject(corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "sourcegraph",
					Name:      "foo",
					Labels: map[string]string{
						"foo": "bar",
						"app": "horsegraph",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Env: []corev1.EnvVar{
								{
									Name:  "var1",
									Value: "value1",
								},
							},
						},
					},
				},
			}),
		},
	}

	for _, tt := range equal {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := HashObject(tt.objectA)

			if !reflect.DeepEqual(tt.ObjectB, got) {
				t.Errorf("HashObject failed. Want: %v, got %v", tt.ObjectB, got)
			}
		})
	}

	unequal := []struct {
		name    string
		objectA any
		ObjectB any
	}{
		{
			name: "hash two pods with the different content",
			objectA: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "sourcegraph",
					Name:      "foo",
					Labels: map[string]string{
						"foo": "bar",
						"app": "horsegraph",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Env: []corev1.EnvVar{
								{
									Name:  "var1",
									Value: "value1",
								},
							},
						},
					},
				},
			},
			ObjectB: HashObject(corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "sourcegraph",
					Name:      "foo",
					Labels: map[string]string{
						"foo": "bar",
						"app": "camelgraph",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Env: []corev1.EnvVar{
								{
									Name:  "var1",
									Value: "value1",
								},
							},
						},
					},
				},
			}),
		},
		{
			name: "hash an object and it's pointer",
			objectA: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "sourcegraph",
					Name:      "foo",
					Labels: map[string]string{
						"foo": "bar",
						"app": "horsegraph",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Env: []corev1.EnvVar{
								{
									Name:  "var1",
									Value: "value1",
								},
							},
						},
					},
				},
			},
			ObjectB: HashObject(corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "sourcegraph",
					Name:      "foo",
					Labels: map[string]string{
						"foo": "bar",
						"app": "horsegraph",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Env: []corev1.EnvVar{
								{
									Name:  "var1",
									Value: "value1",
								},
							},
						},
					},
				},
			}),
		},
	}

	for _, tt := range unequal {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := HashObject(tt.objectA)

			if reflect.DeepEqual(tt.ObjectB, got) {
				t.Errorf("HashObject failed. Want: %v, got %v", tt.ObjectB, got)
			}
		})
	}
}
