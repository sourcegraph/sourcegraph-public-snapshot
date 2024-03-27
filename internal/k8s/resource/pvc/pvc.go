package pvc

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/maps"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// NewPersistentVolumeClaim creates a new k8s PVC with default values.
//
// Default values include:
//
//   - Access mode of `ReadWriteOnce`.
//   - Storage request of 10Gi.
//
// Additional options can be passed to modify the default values.
func NewPersistentVolumeClaim(name, namespace string, options ...Option) (corev1.PersistentVolumeClaim, error) {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
	}

	for _, opt := range options {
		err := opt(&pvc)
		if err != nil {
			return corev1.PersistentVolumeClaim{}, err
		}
	}

	return pvc, nil
}

// Option sets an option for a PVC.
type Option func(pvc *corev1.PersistentVolumeClaim) error

// WithLabels sets the PVC labels without overriding existing labels.
func WithLabels(labels map[string]string) Option {
	return func(pvc *corev1.PersistentVolumeClaim) error {
		pvc.Labels = maps.MergePreservingExistingKeys(pvc.Labels, labels)
		return nil
	}
}

// WithAnnotations sets the PVC annotations without overriding existing annotations.
func WithAnnotations(annotations map[string]string) Option {
	return func(pvc *corev1.PersistentVolumeClaim) error {
		pvc.Annotations = maps.MergePreservingExistingKeys(pvc.Annotations, annotations)
		return nil
	}
}

// WithAccessMode sets the Access Mode for the PVC.
func WithAccessMode(accessModes []corev1.PersistentVolumeAccessMode) Option {
	return func(pvc *corev1.PersistentVolumeClaim) error {
		pvc.Spec.AccessModes = accessModes
		return nil
	}
}

// WithResources sets the given Resource Requirements for the PVC.
func WithResources(resources corev1.ResourceRequirements) Option {
	return func(pvc *corev1.PersistentVolumeClaim) error {
		pvc.Spec.Resources = resources
		return nil
	}
}

// WithStorageClassName sets the storage class name for the PVC.
func WithStorageClassName(storageClassName string) Option {
	return func(pvc *corev1.PersistentVolumeClaim) error {
		pvc.Spec.StorageClassName = pointers.Ptr(storageClassName)
		return nil
	}
}

// WithVolumeName sets the given volume name for the PVC.
func WithVolumeName(volumeName string) Option {
	return func(pvc *corev1.PersistentVolumeClaim) error {
		pvc.Spec.VolumeName = volumeName
		return nil
	}
}
