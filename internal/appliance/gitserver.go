package appliance

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/statefulset"
)

func (r *Reconciler) reconcileGitServer(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	if err := r.reconcileGitServerStatefulSet(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileGitServerService(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileGitServerServiceAccount(ctx, sg, owner); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileGitServerStatefulSet(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.GitServer
	name := "gitserver"

	ctr := container.NewContainer(name, cfg, corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		},
	})

	storageSize, err := resource.ParseQuantity(cfg.StorageSize)
	if err != nil {
		return errors.Wrap(err, "parsing storage size")
	}

	// TODO: https://github.com/sourcegraph/sourcegraph/issues/62076
	ctr.Image = "index.docker.io/sourcegraph/gitserver:5.3.2@sha256:6c6042cf3e5f3f16de9b82e3d4ab1647f8bb924cd315245bd7a3162f5489e8c4"

	ctr.Env = append(ctr.Env, container.EnvVarsRedis()...)
	ctr.Env = append(ctr.Env, container.EnvVarsOtel()...)

	ctr.Ports = []corev1.ContainerPort{
		{Name: "rpc", ContainerPort: 3178},
	}

	ctr.Args = []string{"run"}

	ctr.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString("rpc"),
			},
		},
		TimeoutSeconds:      5,
		InitialDelaySeconds: 5,
	}

	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "tmpdir", MountPath: "/tmp"},
		{Name: "repos", MountPath: "/data/repos"},
	}

	podVolumes := []corev1.Volume{
		{Name: "repos"},
		pod.NewVolumeEmptyDir("tmpdir"),
	}

	if sshSecret := sg.Spec.GitServer.SSHSecret; sshSecret != "" {
		podVolumes = append(podVolumes, corev1.Volume{
			Name: "ssh",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  sshSecret,
					DefaultMode: pointers.Ptr[int32](0644),
				},
			},
		})
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}
	podTemplate.Template.Spec.ServiceAccountName = name
	podTemplate.Template.Spec.Volumes = podVolumes

	sset := statefulset.NewStatefulSet(name, sg.Namespace, sg.Spec.RequestedVersion)
	sset.Spec.Template = podTemplate.Template
	sset.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
		pvc.NewPersistentVolumeClaim("repos", sg.Namespace, storageSize, sg.Spec.StorageClass.Name),
	}

	return reconcileObject(ctx, r, sg.Spec.GitServer, &sset, &appsv1.StatefulSet{}, sg, owner)
}

func (r *Reconciler) reconcileGitServerService(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	svc := service.NewService("gitserver", sg.Namespace, sg.Spec.GitServer)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "unused", TargetPort: intstr.FromInt(10811), Port: 10811},
	}
	svc.Spec.Selector = map[string]string{
		"app":  "gitserver",
		"type": "gitserver",
	}

	return reconcileObject(ctx, r, sg.Spec.GitServer, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileGitServerServiceAccount(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.GitServer
	sa := serviceaccount.NewServiceAccount("gitserver", sg.Namespace, cfg)
	return reconcileObject(ctx, r, sg.Spec.GitServer, &sa, &corev1.ServiceAccount{}, sg, owner)
}
