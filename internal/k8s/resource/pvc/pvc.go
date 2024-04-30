package pvc

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewPersistentVolumeClaim creates a new k8s PVC with some default values set.
func NewPersistentVolumeClaim(name, namespace string, storage resource.Quantity, storageClassName string) corev1.PersistentVolumeClaim {
	pvc := NewPersistentVolumeClaimSpecOnly(storage, storageClassName)
	pvc.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			"deploy": "sourcegraph",
		},
	}
	return pvc
}

// Useful for statefulsets, that do not require metadata
func NewPersistentVolumeClaimSpecOnly(storage resource.Quantity, storageClassName string) corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storage,
				},
			},
			StorageClassName: &storageClassName,
		},
	}
}
