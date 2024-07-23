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
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func (r *Reconciler) reconcileWorker(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileWorkerDeployment(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Deployment")
	}
	if err := r.reconcileWorkerService(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Service")
	}
	if err := r.reconcileWorkerExecutorsService(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling executors Service")
	}
	if err := r.reconcileWorkerServiceAccount(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ServiceAccount")
	}
	return nil
}

func (r *Reconciler) reconcileWorkerDeployment(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "worker"
	cfg := sg.Spec.Worker

	defaultImage := config.GetDefaultImage(sg, name)
	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: defaultImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("2G"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4G"),
			},
		},
	})

	ctr.Env = append(ctr.Env, container.EnvVarsRedis()...)
	ctr.Env = addPreciseCodeIntelBlobstoreVars(ctr.Env, sg)
	ctr.Env = append(
		ctr.Env,
		container.NewEnvVarFieldRef("POD_NAME", "metadata.name"),
	)
	ctr.Env = append(ctr.Env, container.EnvVarsOtel()...)

	ctr.Ports = []corev1.ContainerPort{
		{Name: "http", ContainerPort: 3189},
		{Name: "debug", ContainerPort: 6060},
		{Name: "prom", ContainerPort: 6996},
	}
	ctr.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz",
				Port: intstr.FromString("debug"),
			},
		},
		InitialDelaySeconds: 60,
		TimeoutSeconds:      5,
	}
	ctr.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/ready",
				Port: intstr.FromString("debug"),
			},
		},
		PeriodSeconds:  5,
		TimeoutSeconds: 5,
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}
	redisConnSpecs, err := r.getRedisSecrets(ctx, sg)
	if err != nil {
		return err
	}
	redisConnHash, err := configHash(redisConnSpecs)
	if err != nil {
		return err
	}
	podTemplate.Template.ObjectMeta.Annotations["checksum/redis"] = redisConnHash

	dep := deployment.NewDeployment(name, sg.Namespace, sg.Spec.RequestedVersion)
	dep.Spec.Replicas = pointers.Ptr(cfg.Replicas)
	dep.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
		MaxSurge:       pointers.Ptr(intstr.FromInt32(1)),
		MaxUnavailable: pointers.Ptr(intstr.FromInt32(1)),
	}
	dep.Spec.Template = podTemplate.Template

	ifChanged := struct {
		config.WorkerSpec
		RedisConnSpecs
	}{
		WorkerSpec:     cfg,
		RedisConnSpecs: redisConnSpecs,
	}
	return reconcileObject(ctx, r, ifChanged, &dep, &appsv1.Deployment{}, sg, owner)
}

func (r *Reconciler) reconcileWorkerService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "worker"
	cfg := sg.Spec.Worker

	svc := service.NewService(name, sg.Namespace, cfg)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 3189, TargetPort: intstr.FromString("http")},
		{Name: "debug", Port: 6060, TargetPort: intstr.FromString("debug")},
	}
	svc.Spec.Selector = map[string]string{
		"app": name,
	}

	return reconcileObject(ctx, r, cfg, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileWorkerExecutorsService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.Worker

	svc := service.NewService("worker-executors", sg.Namespace, nil)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "prom", Port: 6996, TargetPort: intstr.FromString("prom")},
	}
	svc.Spec.Selector = map[string]string{
		"app": "worker",
	}
	svc.SetAnnotations(map[string]string{
		"prometheus.io/port":            "6996",
		"sourcegraph.prometheus/scrape": "true",
	})

	return reconcileObject(ctx, r, cfg, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileWorkerServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.Worker
	sa := serviceaccount.NewServiceAccount("worker", sg.Namespace, cfg)
	return reconcileObject(ctx, r, cfg, &sa, &corev1.ServiceAccount{}, sg, owner)
}
