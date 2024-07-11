package reconciler

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/deployment"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
)

func (r *Reconciler) reconcileBlobstore(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
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

func buildBlobstorePersistentVolumeClaim(sg *config.Sourcegraph) (corev1.PersistentVolumeClaim, error) {
	return pvc.NewPersistentVolumeClaim("blobstore", sg.Namespace, sg.Spec.Blobstore)
}

func (r *Reconciler) reconcileBlobstorePersistentVolumeClaims(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	p, err := buildBlobstorePersistentVolumeClaim(sg)
	if err != nil {
		return err
	}

	return reconcileObject(ctx, r, sg.Spec.Blobstore, &p, &corev1.PersistentVolumeClaim{}, sg, owner)
}

func buildBlobstoreService(sg *config.Sourcegraph) corev1.Service {
	name := "blobstore"

	s := service.NewService(name, sg.Namespace, sg.Spec.Blobstore)
	s.Spec.Ports = []corev1.ServicePort{
		{
			Name:       name,
			Port:       9000,
			TargetPort: intstr.FromString(name),
		},
	}
	s.Spec.Selector = map[string]string{
		"app": name,
	}

	return s
}

func (r *Reconciler) reconcileBlobstoreServices(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	s := buildBlobstoreService(sg)
	return reconcileObject(ctx, r, sg.Spec.Blobstore, &s, &corev1.Service{}, sg, owner)
}

func buildBlobstoreDeployment(sg *config.Sourcegraph) appsv1.Deployment {
	name := "blobstore"

	containerPorts := []corev1.ContainerPort{{
		Name:          name,
		ContainerPort: 9000,
	}}

	containerVolumeMounts := []corev1.VolumeMount{
		{
			Name:      "blobstore",
			MountPath: "/blobstore",
		},
		{
			Name:      "blobstore-data",
			MountPath: "/data",
		},
	}

	defaultImage := config.GetDefaultImage(sg, name)
	defaultContainer := container.NewContainer(name, sg.Spec.Blobstore, config.ContainerConfig{
		Image: defaultImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("500M"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("500M"),
			},
		},
	})

	defaultContainer.Ports = containerPorts
	defaultContainer.VolumeMounts = containerVolumeMounts

	podVolumes := []corev1.Volume{
		{
			Name: "blobstore",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "blobstore-data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "blobstore",
				},
			},
		},
	}

	podTemplate := pod.NewPodTemplate(name, sg.Spec.Blobstore)
	podTemplate.Template.Spec.Containers = []corev1.Container{defaultContainer}
	podTemplate.Template.Spec.Volumes = podVolumes

	defaultDeployment := deployment.NewDeployment(
		name,
		sg.Namespace,
		sg.Spec.RequestedVersion,
	)
	defaultDeployment.Spec.Template = podTemplate.Template

	return defaultDeployment
}

func (r *Reconciler) reconcileBlobstoreDeployments(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	d := buildBlobstoreDeployment(sg)
	return reconcileObject(ctx, r, sg.Spec.Blobstore, &d, &appsv1.Deployment{}, sg, owner)
}
