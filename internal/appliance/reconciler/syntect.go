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

func (r *Reconciler) reconcileSyntect(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileSyntectDeployment(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Deployment")
	}
	if err := r.reconcileSyntectService(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Service")
	}
	if err := r.reconcileSyntectServiceAccount(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ServiceAccount")
	}
	return nil
}

func (r *Reconciler) reconcileSyntectDeployment(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "syntect-server"
	cfg := sg.Spec.SyntectServer

	defaultImage := config.GetDefaultImage(sg, "syntax-highlighter")
	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: defaultImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("2G"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("6G"),
			},
		},
	})
	ctr.Ports = []corev1.ContainerPort{
		{Name: "http", ContainerPort: 9238},
	}
	ctr.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/health",
				Port: intstr.FromString("http"),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
	}
	ctr.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString("http"),
			},
		},
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}
	podTemplate.Template.Spec.ServiceAccountName = name

	dep := deployment.NewDeployment(name, sg.Namespace, sg.Spec.RequestedVersion)
	dep.Spec.Replicas = pointers.Ptr(cfg.Replicas)
	dep.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
		MaxSurge:       pointers.Ptr(intstr.FromInt32(1)),
		MaxUnavailable: pointers.Ptr(intstr.FromInt32(0)),
	}
	dep.Spec.Template = podTemplate.Template

	return reconcileObject(ctx, r, cfg, &dep, &appsv1.Deployment{}, sg, owner)
}

func (r *Reconciler) reconcileSyntectService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "syntect-server"
	cfg := sg.Spec.SyntectServer

	svc := service.NewService(name, sg.Namespace, cfg)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 9238, TargetPort: intstr.FromString("http")},
	}
	svc.Spec.Selector = map[string]string{
		"app": "syntect-server",
	}

	return reconcileObject(ctx, r, cfg, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileSyntectServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.SyntectServer
	sa := serviceaccount.NewServiceAccount("syntect-server", sg.Namespace, cfg)
	return reconcileObject(ctx, r, cfg, &sa, &corev1.ServiceAccount{}, sg, owner)
}
