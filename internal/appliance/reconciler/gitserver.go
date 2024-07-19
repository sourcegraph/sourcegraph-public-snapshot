package reconciler

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/statefulset"
)

func (r *Reconciler) reconcileGitServer(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
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

func (r *Reconciler) reconcileGitServerStatefulSet(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.GitServer
	name := "gitserver"

	defaultImage := config.GetDefaultImage(sg, name)
	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: defaultImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
	})

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
	redisConnSpecs, err := r.getRedisSecrets(ctx, sg)
	if err != nil {
		return err
	}
	redisConnHash, err := configHash(redisConnSpecs)
	if err != nil {
		return err
	}
	podTemplate.Template.ObjectMeta.Annotations["checksum/redis"] = redisConnHash

	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}
	podTemplate.Template.Spec.ServiceAccountName = name
	podTemplate.Template.Spec.Volumes = podVolumes

	pvc, err := pvc.NewPersistentVolumeClaim("repos", sg.Namespace, sg.Spec.GitServer)
	if err != nil {
		return err
	}

	sset := statefulset.NewStatefulSet(name, sg.Namespace, sg.Spec.RequestedVersion)
	sset.Spec.Template = podTemplate.Template
	sset.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{pvc}

	ifChanged := struct {
		config.GitServerSpec
		RedisConnSpecs
	}{
		GitServerSpec:  cfg,
		RedisConnSpecs: redisConnSpecs,
	}
	return reconcileObject(ctx, r, ifChanged, &sset, &appsv1.StatefulSet{}, sg, owner)
}

func (r *Reconciler) reconcileGitServerService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	svc := service.NewService("gitserver", sg.Namespace, sg.Spec.GitServer)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "unused", TargetPort: intstr.FromInt32(10811), Port: 10811},
	}
	svc.Spec.Selector = map[string]string{
		"app":  "gitserver",
		"type": "gitserver",
	}

	return reconcileObject(ctx, r, sg.Spec.GitServer, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileGitServerServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.GitServer
	sa := serviceaccount.NewServiceAccount("gitserver", sg.Namespace, cfg)
	return reconcileObject(ctx, r, sg.Spec.GitServer, &sa, &corev1.ServiceAccount{}, sg, owner)
}
