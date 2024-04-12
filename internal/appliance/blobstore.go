package appliance

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/appliance/hash"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/maps"
)

func (r *Reconciler) reconcileBlobstore(ctx context.Context, sg *Sourcegraph) error {
	if err := r.reconcileBlobstorePersistentVolumeClaims(ctx, sg); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) buildBlobstorePersistentVolumeClaim(sg *Sourcegraph) (corev1.PersistentVolumeClaim, error) {
	storage := sg.Spec.Blobstore.StorageSize
	if storage == "" {
		storage = "100Gi"
	}

	if _, err := resource.ParseQuantity(storage); err != nil {
		return corev1.PersistentVolumeClaim{}, errors.Errorf("invalid blobstore storage size: %s", storage)
	}

	storageClassName := sg.Spec.StorageClass.Name
	if storageClassName == "" {
		storageClassName = "sourcegraph"
	}

	persistentVolumeClaim, err := pvc.NewPersistentVolumeClaim("blobstore", sg.Namespace,
		pvc.WithResources(corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(storage),
			},
		}),
	)
	if err != nil {
		return corev1.PersistentVolumeClaim{}, err
	}

	// set StorageClass name if a custom storage class is being created.
	if sg.Spec.StorageClass.Create {
		_ = pvc.WithStorageClassName(storageClassName)(&persistentVolumeClaim)
	}

	return persistentVolumeClaim, nil
}

func (r *Reconciler) reconcileBlobstorePersistentVolumeClaims(ctx context.Context, sg *Sourcegraph) error {
	p, err := r.buildBlobstorePersistentVolumeClaim(sg)
	if err != nil {
		return err
	}

	p.Labels = hash.SetTemplateHashLabel(p.Labels, p.Spec)

	var existing corev1.PersistentVolumeClaim
	if r.IsObjectFound(ctx, p.Name, p.Namespace, &existing) {
		if sg.Spec.Blobstore.Disabled {
			return nil
		}

		// Object exists update if needed
		if hash.GetTemplateHashLabel(existing.Labels) == hash.GetTemplateHashLabel(p.Labels) {
			// no updates needed
			return nil
		}

		// need to update
		existing.Labels = maps.Merge(existing.Labels, p.Labels)
		existing.Annotations = maps.Merge(existing.Annotations, p.Annotations)
		existing.Spec = p.Spec

		return r.Update(ctx, &existing)
	}

	if sg.Spec.Blobstore.Disabled {
		return nil
	}

	// Note: we don't set a controller reference here as we want PVCs to persist if blobstore is deleted.
	// This helps to protect against accidental data deletions.

	return r.Create(ctx, &p)
}

func (r *Reconciler) buildBlobstoreService(sg *Sourcegraph) (corev1.Service, error) {
	name := "blobstore"

}
