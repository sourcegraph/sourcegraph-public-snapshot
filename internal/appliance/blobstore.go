package appliance

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/deployment"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) reconcileBlobstore(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	if err := r.reconcileBlobstorePersistentVolumeClaims(ctx, sg, owner); err != nil {
		return err
	}

	if err := r.reconcileBlobstoreServices(ctx, sg, owner); err != nil {
		return err
	}

	if err := r.reconcileBlobstoreDeployments(ctx, sg, owner); err != nil {
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

func (r *Reconciler) reconcileBlobstorePersistentVolumeClaims(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	p, err := buildBlobstorePersistentVolumeClaim(sg)
	if err != nil {
		return err
	}

	return reconcileBlobStoreObject(ctx, r, &p, &corev1.PersistentVolumeClaim{}, sg, owner)
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

func (r *Reconciler) reconcileBlobstoreServices(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	s, err := buildBlobstoreService(sg)
	if err != nil {
		return err
	}
	return reconcileBlobStoreObject(ctx, r, &s, &corev1.Service{}, sg, owner)
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
		sg.Spec.RequestedVersion,
		deployment.WithPodTemplateSpec(podTemplate.Template),
	)

	if err != nil {
		return appsv1.Deployment{}, err
	}

	return defaultDeployment, nil
}

func (r *Reconciler) reconcileBlobstoreDeployments(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	d, err := buildBlobstoreDeployment(sg)
	if err != nil {
		return err
	}
	return reconcileBlobStoreObject(ctx, r, &d, &appsv1.Deployment{}, sg, owner)
}

func reconcileBlobStoreObject[T client.Object](ctx context.Context, r *Reconciler, obj, objKind T, sg *Sourcegraph, owner client.Object) error {
	if sg.Spec.Blobstore.Disabled {
		return r.ensureObjectDeleted(ctx, obj)
	}

	// Any secrets (or other configmaps) referenced in BlobStoreSpec can be
	// added to this struct so that they are hashed, and cause an update to the
	// Deployment if changed.
	updateIfChanged := struct {
		BlobstoreSpec
		Version string
	}{
		BlobstoreSpec: sg.Spec.Blobstore,
		Version:       sg.Spec.RequestedVersion,
	}

	return createOrUpdateObject(ctx, r, updateIfChanged, owner, obj, objKind)
}
