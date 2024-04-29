package appliance

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/deployment"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/internal/maps"
)

func (r *Reconciler) reconcileRepoUpdater(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	if err := r.reconcileRepoUpdaterDeployment(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileRepoUpdaterService(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileRepoUpdaterServiceAccount(ctx, sg, owner); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileRepoUpdaterService(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	svc := service.NewService("repo-updater", sg.Namespace)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", TargetPort: intstr.FromString("http"), Port: 3182},
	}
	svc.Spec.Selector = map[string]string{
		"app": "repo-updater",
	}

	return reconcileRepoUpdaterObject(ctx, r, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileRepoUpdaterDeployment(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.RepoUpdater
	name := "repo-updater"

	ctr := container.NewContainer(name)

	// TODO: https://github.com/sourcegraph/sourcegraph/issues/62076
	ctr.Image = "index.docker.io/sourcegraph/repo-updater:5.3.2@sha256:5a414aa030c7e0922700664a43b449ee5f3fafa68834abef93988c5992c747c6"

	ctr.Env = []corev1.EnvVar{
		container.NewEnvVarSecretKeyRef("REDIS_CACHE_ENDPOINT", "redis-cache", "endpoint"),
		container.NewEnvVarSecretKeyRef("REDIS_STORE_ENDPOINT", "redis-store", "endpoint"),

		// OTEL_AGENT_HOST must be defined before OTEL_EXPORTER_OTLP_ENDPOINT to substitute the node IP on which the DaemonSet pod instance runs in the latter variable
		container.NewEnvVarFieldRef("OTEL_AGENT_HOST", "status.hostIP"),
		{Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: "http://$(OTEL_AGENT_HOST):4317"},
	}

	ctr.Ports = []corev1.ContainerPort{
		{Name: "http", ContainerPort: 3182},
		{Name: "debug", ContainerPort: 6060},
	}

	ctr.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   intstr.FromString("debug"),
				Scheme: corev1.URISchemeHTTP,
			},
		},
		FailureThreshold: 3,
		PeriodSeconds:    1,
		TimeoutSeconds:   5,
	}
	ctr.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/ready",
				Port:   intstr.FromString("debug"),
				Scheme: corev1.URISchemeHTTP,
			},
		},
		FailureThreshold: 3,
		PeriodSeconds:    1,
		TimeoutSeconds:   5,
	}

	// default resources
	ctr.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("500Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}

	if cfg.Resources != nil {
		ctr.Resources = *cfg.Resources
	}

	podTemplate := pod.NewPodTemplate(name)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}

	dep := deployment.NewDeployment(name, sg.Namespace, sg.Spec.RequestedVersion)
	dep.Spec.Template = podTemplate.Template
	dep.Spec.Template.Spec.ServiceAccountName = name

	return reconcileRepoUpdaterObject(ctx, r, &dep, &appsv1.Deployment{}, sg, owner)
}

func (r *Reconciler) reconcileRepoUpdaterServiceAccount(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.RepoUpdater
	sa := serviceaccount.NewServiceAccount("repo-updater", sg.Namespace)
	sa.SetAnnotations(maps.Merge(sa.GetAnnotations(), cfg.ServiceAccountAnnotations))
	return reconcileRepoUpdaterObject(ctx, r, &sa, &corev1.ServiceAccount{}, sg, owner)
}

func reconcileRepoUpdaterObject[T client.Object](ctx context.Context, r *Reconciler, obj, objKind T, sg *Sourcegraph, owner client.Object) error {
	if sg.Spec.RepoUpdater.Disabled {
		return r.ensureObjectDeleted(ctx, obj)
	}

	// Any secrets (or other configmaps) referenced in this spec can be
	// added to this struct so that they are hashed, and cause an update to the
	// resource if changed.
	updateIfChanged := struct {
		RepoUpdaterSpec
		Version string
	}{
		RepoUpdaterSpec: sg.Spec.RepoUpdater,
		Version:         sg.Spec.RequestedVersion,
	}

	return createOrUpdateObject(ctx, r, updateIfChanged, owner, obj, objKind)
}
