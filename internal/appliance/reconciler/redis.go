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
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/secret"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func (r *Reconciler) reconcileRedis(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileRedisInstance(ctx, sg, owner, "cache", sg.Spec.RedisCache); err != nil {
		return errors.Wrap(err, "reconciling redis-cache")
	}
	if err := r.reconcileRedisInstance(ctx, sg, owner, "store", sg.Spec.RedisStore); err != nil {
		return errors.Wrap(err, "reconciling redis-store")
	}
	return nil
}

func (r *Reconciler) reconcileRedisInstance(ctx context.Context, sg *config.Sourcegraph, owner client.Object, kind string, cfg config.RedisSpec) error {
	if err := r.reconcileRedisDeployment(ctx, sg, owner, kind, cfg); err != nil {
		return errors.Wrap(err, "reconciling Deployment")
	}
	if err := r.reconcileRedisPVC(ctx, sg, owner, kind, cfg); err != nil {
		return errors.Wrap(err, "reconciling PersistentVolumeClaim")
	}
	if err := r.reconcileRedisService(ctx, sg, owner, kind, cfg); err != nil {
		return errors.Wrap(err, "reconciling Service")
	}
	if err := r.reconcileRedisSecret(ctx, sg, owner, kind, cfg); err != nil {
		return errors.Wrap(err, "reconciling Secret")
	}
	return nil
}

func (r *Reconciler) reconcileRedisDeployment(ctx context.Context, sg *config.Sourcegraph, owner client.Object, kind string, cfg config.RedisSpec) error {
	name := "redis-" + kind

	defaultImage := config.GetDefaultImage(sg, name)
	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: defaultImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("7Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("7Gi"),
			},
		},
	})
	ctr.Ports = []corev1.ContainerPort{
		{Name: "redis", ContainerPort: 6379},
	}
	ctr.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString("redis"),
			},
		},
		FailureThreshold:    2,
		InitialDelaySeconds: 60,
		PeriodSeconds:       30,
		TimeoutSeconds:      5,
	}
	ctr.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/bin/sh", "-c",
					`
#!/bin/bash
if [ -f /etc/redis/redis.conf ]; then
  REDISCLI_AUTH=$(grep -h "requirepass" /etc/redis/redis.conf | cut -d ' ' -f 2)
fi
response=$(
  redis-cli ping
)
if [ "$response" != "PONG" ]; then
  echo "$response"
  exit 1
fi
`,
				},
			},
		},
		InitialDelaySeconds: 5,
	}
	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "redis-data", MountPath: "/redis-data"},
	}
	ctr.SecurityContext.RunAsUser = pointers.Ptr(int64(999))
	ctr.SecurityContext.RunAsGroup = pointers.Ptr(int64(1000))

	exporterImage := config.GetDefaultImage(sg, "redis_exporter")
	exporterCtr := container.NewContainer("redis-exporter", cfg, config.ContainerConfig{
		Image: exporterImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
		},
	})
	exporterCtr.Ports = []corev1.ContainerPort{
		{Name: "redisexp", ContainerPort: 9121},
	}
	exporterCtr.SecurityContext.RunAsUser = pointers.Ptr(int64(999))
	exporterCtr.SecurityContext.RunAsGroup = pointers.Ptr(int64(1000))

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr, exporterCtr}
	podTemplate.Template.Spec.Volumes = []corev1.Volume{
		pod.NewVolumeFromPVC("redis-data", name),
	}
	podTemplate.Template.Spec.SecurityContext.FSGroup = pointers.Ptr(int64(1000))

	dep := deployment.NewDeployment(name, sg.Namespace, sg.Spec.RequestedVersion)
	dep.Spec.Strategy = appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType}
	dep.Spec.Template = podTemplate.Template

	return reconcileObject(ctx, r, cfg, &dep, &appsv1.Deployment{}, sg, owner)
}

func (r *Reconciler) reconcileRedisService(ctx context.Context, sg *config.Sourcegraph, owner client.Object, kind string, cfg config.RedisSpec) error {
	name := "redis-" + kind
	svc := service.NewService(name, sg.Namespace, cfg)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "redis", Port: 6379, TargetPort: intstr.FromString("redis")},
	}
	svc.Spec.Selector = map[string]string{
		"app": name,
	}
	return reconcileObject(ctx, r, cfg, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileRedisPVC(ctx context.Context, sg *config.Sourcegraph, owner client.Object, kind string, cfg config.RedisSpec) error {
	name := "redis-" + kind
	pvc, err := pvc.NewPersistentVolumeClaim(name, sg.Namespace, cfg)
	if err != nil {
		return err
	}
	return reconcileObject(ctx, r, cfg, &pvc, &corev1.PersistentVolumeClaim{}, sg, owner)
}

func (r *Reconciler) reconcileRedisSecret(ctx context.Context, sg *config.Sourcegraph, owner client.Object, kind string, cfg config.RedisSpec) error {
	name := "redis-" + kind
	secret := secret.NewSecret(name, sg.Namespace, sg.Spec.RequestedVersion)
	secret.StringData = map[string]string{
		"endpoint": name + ":6379",
	}
	return reconcileObject(ctx, r, cfg, &secret, &corev1.Secret{}, sg, owner)
}
