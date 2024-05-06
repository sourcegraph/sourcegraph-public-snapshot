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
	svc := service.NewService("repo-updater", sg.Namespace, sg.Spec.RepoUpdater)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", TargetPort: intstr.FromString("http"), Port: 3182},
	}
	svc.Spec.Selector = map[string]string{
		"app": "repo-updater",
	}

	return reconcileObject(ctx, r, sg.Spec.RepoUpdater, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileRepoUpdaterDeployment(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.RepoUpdater
	name := "repo-updater"

	ctr := container.NewContainer(name, cfg, corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("500Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	})

	// TODO: https://github.com/sourcegraph/sourcegraph/issues/62076
	ctr.Image = "index.docker.io/sourcegraph/repo-updater:5.3.2@sha256:5a414aa030c7e0922700664a43b449ee5f3fafa68834abef93988c5992c747c6"

	ctr.Env = append(ctr.Env, container.EnvVarsRedis()...)
	ctr.Env = append(ctr.Env, container.EnvVarsOtel()...)

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

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}

	dep := deployment.NewDeployment(name, sg.Namespace, sg.Spec.RequestedVersion)
	dep.Spec.Template = podTemplate.Template
	dep.Spec.Template.Spec.ServiceAccountName = name

	return reconcileObject(ctx, r, sg.Spec.RepoUpdater, &dep, &appsv1.Deployment{}, sg, owner)
}

func (r *Reconciler) reconcileRepoUpdaterServiceAccount(ctx context.Context, sg *Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.RepoUpdater
	sa := serviceaccount.NewServiceAccount("repo-updater", sg.Namespace, cfg)
	return reconcileObject(ctx, r, sg.Spec.RepoUpdater, &sa, &corev1.ServiceAccount{}, sg, owner)
}
