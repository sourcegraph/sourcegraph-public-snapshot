package appliance

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/internal/appliance/hash"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/deployment"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/maps"
)

func (r *Reconciler) reconcileBlobstore(ctx context.Context, sg *Sourcegraph) error {
	if err := r.reconcileBlobstorePersistentVolumeClaims(ctx, sg); err != nil {
		return err
	}

	if err := r.reconcileBlobstoreServices(ctx, sg); err != nil {
		return err
	}

	if err := r.reconcileBlobstoreDeployments(ctx, sg); err != nil {
		return err
	}

	return nil
}

func buildBlobstorePersistentVolumeClaim(sg *Sourcegraph) (corev1.PersistentVolumeClaim, error) {
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

	p, err := pvc.NewPersistentVolumeClaim("blobstore", sg.Namespace,
		pvc.WithResources(corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(storage),
			},
		}),
	)
	if err != nil {
		return corev1.PersistentVolumeClaim{}, err
	}

	// set StorageClass name if a custom storage class is being sgeated.
	if sg.Spec.StorageClass.Create {
		_ = pvc.WithStorageClassName(storageClassName)(&p)
	}

	return p, nil
}

func (r *Reconciler) reconcileBlobstorePersistentVolumeClaims(ctx context.Context, sg *Sourcegraph) error {
	p, err := buildBlobstorePersistentVolumeClaim(sg)
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

func buildBlobstoreService(sg *Sourcegraph) (corev1.Service, error) {
	name := "blobstore"

	s, err := service.NewService(name, sg.Namespace,
		service.WithPorts(corev1.ServicePort{
			Name:       name,
			Port:       9000,
			TargetPort: intstr.FromString(name),
		}),
		service.WithSelector(map[string]string{
			"app": name,
		}),
	)

	if err != nil {
		return corev1.Service{}, err
	}

	return s, nil
}

func (r *Reconciler) reconcileBlobstoreServices(ctx context.Context, sg *Sourcegraph) error {
	s, err := buildBlobstoreService(sg)
	if err != nil {
		return err
	}

	s.Labels = hash.SetTemplateHashLabel(s.Labels, s.Spec)

	var existing corev1.Service
	if r.IsObjectFound(ctx, s.Name, sg.Namespace, &existing) {
		if sg.Spec.Blobstore.Disabled {
			// blobstore service exists, but has been disabled. Delete the service.
			//
			// Using a precondition to make sure the version of the resource that is deleted
			// is the version we intend, and not a resource that was already resgeated.
			err = r.Delete(ctx, &existing, client.Preconditions{
				UID:             &existing.UID,
				ResourceVersion: &existing.ResourceVersion,
			})

			if err != nil && !apierrors.IsNotFound(err) {
				return err
			}

			return nil
		}
		// Object exists update if needed
		if hash.GetTemplateHashLabel(existing.Labels) == hash.GetTemplateHashLabel(s.Labels) {
			// no updates needed
			return nil
		}

		// need to update
		existing.Labels = maps.Merge(existing.Labels, s.Labels)
		existing.Annotations = maps.Merge(existing.Annotations, s.Annotations)
		existing.Spec = s.Spec

		return r.Update(ctx, &existing)
	}

	if sg.Spec.Blobstore.Disabled {
		return nil
	}

	// TODO set owner ref

	return r.Create(ctx, &s)
}

func buildBlobstoreDeployment(sg *Sourcegraph) (appsv1.Deployment, error) {
	name := "blobstore"

	containerImage := ""

	containerPorts := corev1.ContainerPort{
		Name:          name,
		ContainerPort: 9000,
	}

	containerResources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("500M"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("500M"),
		},
	}
	if sg.Spec.Blobstore.Resources != nil {
		limCPU := sg.Spec.Blobstore.Resources.Limits.Cpu()
		limMem := sg.Spec.Blobstore.Resources.Limits.Memory()
		reqCPU := sg.Spec.Blobstore.Resources.Limits.Cpu()
		reqMem := sg.Spec.Blobstore.Resources.Limits.Memory()

		containerResources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    pointers.DerefZero(limCPU),
				corev1.ResourceMemory: pointers.DerefZero(limMem),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    pointers.DerefZero(reqCPU),
				corev1.ResourceMemory: pointers.DerefZero(reqMem),
			},
		}
	}

	if sg.Spec.LocalDevMode {
		containerResources = corev1.ResourceRequirements{}
	}

	containerVolumeMounts := []corev1.VolumeMount{
		{
			Name:      "blobstore-data",
			MountPath: "/data",
		},
		{
			Name:      "blobstore",
			MountPath: "/blobstore",
		},
	}

	defaultContainer, err := container.NewContainer(name,
		container.WithPorts(containerPorts),
		container.WithImage(containerImage),
		container.WithResources(containerResources),
		container.WithVolumeMounts(containerVolumeMounts),
	)

	if err != nil {
		return appsv1.Deployment{}, err
	}

	podVolumes := []corev1.Volume{
		{
			Name: "blobstore-data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "blobstore",
				},
			},
		},
		{
			Name: "blobstore",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	podTemplate, err := pod.NewPodTemplate(name, sg.Namespace,
		pod.WithContainers(defaultContainer),
		pod.WithVolumes(podVolumes),
	)
	if err != nil {
		return appsv1.Deployment{}, err
	}

	defaultDeployment, err := deployment.NewDeployment(
		name,
		sg.Namespace,
		deployment.WithPodTemplateSpec(podTemplate.Template),
	)

	if err != nil {
		return appsv1.Deployment{}, err
	}

	return defaultDeployment, nil
}

func (r *Reconciler) reconcileBlobstoreDeployments(ctx context.Context, sg *Sourcegraph) error {
	d, err := buildBlobstoreDeployment(sg)
	if err != nil {
		return err
	}

	d.Labels = hash.SetTemplateHashLabel(d.Labels, d.Spec)

	var existing appsv1.Deployment
	if r.IsObjectFound(ctx, d.Name, sg.Namespace, &existing) {
		if sg.Spec.Blobstore.Disabled {
			// blobstore deployment exists, but has been disabled. Delete the deployment.
			//
			// Using a precondition to make sure the version of the resource that is deleted
			// is the version we intend, and not a resource that was already recreated.
			err = r.Delete(ctx, &existing, client.Preconditions{
				UID:             &existing.UID,
				ResourceVersion: &existing.ResourceVersion,
			})

			if err != nil && !apierrors.IsNotFound(err) {
				return err
			}

			return nil
		}
		// Object exists update if needed
		if hash.GetTemplateHashLabel(existing.Labels) == hash.GetTemplateHashLabel(d.Labels) {
			// no updates needed
			return nil
		}

		// need to update
		existing.Labels = maps.Merge(existing.Labels, d.Labels)
		existing.Annotations = maps.Merge(existing.Annotations, d.Annotations)
		existing.Spec = d.Spec

		return r.Update(ctx, &existing)
	}

	if sg.Spec.Blobstore.Disabled {
		return nil
	}

	// TODO set owner ref

	return r.Create(ctx, &d)
}
