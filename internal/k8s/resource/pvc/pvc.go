package pvc

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewPersistentVolumeClaim creates a new k8s PVC with some default values set.
func NewPersistentVolumeClaim(name, namespace string, cfg config.StandardComponent) (corev1.PersistentVolumeClaim, error) {
	// If a nil value is passed in, default to zero values. Callers will then
	// have to override these values, and golden tests can catch any issues.
	var storageCfg config.PersistentVolumeConfig
	if cfg != nil {
		storageCfg = cfg.GetPersistentVolumeConfig()
	}

	storage, err := resource.ParseQuantity(storageCfg.StorageSize)
	if err != nil {
		return corev1.PersistentVolumeClaim{}, errors.Wrap(err, "parsing storage size")
	}
	return corev1.PersistentVolumeClaim{
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
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storage,
				},
			},
			StorageClassName: storageCfg.StorageClassName,
		},
	}, nil
}
