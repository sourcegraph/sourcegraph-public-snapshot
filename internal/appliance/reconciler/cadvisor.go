package reconciler

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/daemonset"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func (r *Reconciler) reconcileCadvisor(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileCadvisorDaemonset(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Daemonset")
	}
	if err := r.reconcileCadvisorServiceAccount(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ServiceAccount")
	}
	return nil
}

func (r *Reconciler) reconcileCadvisorDaemonset(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "cadvisor"
	cfg := sg.Spec.Cadvisor

	defaultImage := config.GetDefaultImage(sg, name)
	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: defaultImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("150m"),
				corev1.ResourceMemory: resource.MustParse("200Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("300m"),
				corev1.ResourceMemory: resource.MustParse("2000Mi"),
			},
		},
	})
	ctr.Args = []string{
		"--store_container_labels=false",
		"--whitelisted_container_labels=io.kubernetes.container.name,io.kubernetes.pod.name,io.kubernetes.pod.namespace,io.kubernetes.pod.uid",
	}
	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "rootfs", MountPath: "/rootfs", ReadOnly: true},
		{Name: "var-run", MountPath: "/var/run", ReadOnly: true},
		{Name: "sys", MountPath: "/sys", ReadOnly: true},
		{Name: "docker", MountPath: "/var/lib/docker", ReadOnly: true},
		{Name: "disk", MountPath: "/dev/disk", ReadOnly: true},
		{Name: "kmsg", MountPath: "/dev/kmsg", ReadOnly: true},
	}
	ctr.Ports = []corev1.ContainerPort{
		{Name: "http", ContainerPort: 48080},
	}
	ctr.SecurityContext = &corev1.SecurityContext{
		Privileged: pointers.Ptr(true),
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}
	podTemplate.Template.Spec.ServiceAccountName = name
	podTemplate.Template.Spec.AutomountServiceAccountToken = pointers.Ptr(false)
	podTemplate.Template.Spec.Volumes = []corev1.Volume{
		pod.NewVolumeHostPath("rootfs", "/"),
		pod.NewVolumeHostPath("var-run", "/var/run"),
		pod.NewVolumeHostPath("sys", "/sys"),
		pod.NewVolumeHostPath("docker", "/var/lib/docker"),
		pod.NewVolumeHostPath("disk", "/dev/disk"),
		pod.NewVolumeHostPath("kmsg", "/dev/kmsg"),
	}
	podTemplate.Template.Spec.SecurityContext = nil

	// Usually we set the prometheus scrape annotations on a Service (and scrape
	// its endpoints rather than the load balancer), but it doesn't usually make
	// sense to deploy services alongside daemonsets. We set the scrape
	// annotation directly on the pod template here instead.
	// Even though this uses the PrometheusPort standard config feature, we
	// shouldn't move this code into pod.go, because otherwise every pod
	// template will have such an annotation.
	if promPort := cfg.GetPrometheusPort(); promPort != nil {
		annotations := map[string]string{
			"prometheus.io/port":            fmt.Sprintf("%d", *promPort),
			"sourcegraph.prometheus/scrape": "true",
		}
		podTemplate.Template.Annotations = annotations
	}

	ds := daemonset.New(name, sg.Namespace, sg.Spec.RequestedVersion)
	ds.Spec.Template = podTemplate.Template

	return reconcileObject(ctx, r, cfg, &ds, &appsv1.DaemonSet{}, sg, owner)
}

func (r *Reconciler) reconcileCadvisorServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.Cadvisor
	sa := serviceaccount.NewServiceAccount("cadvisor", sg.Namespace, cfg)
	return reconcileObject(ctx, r, cfg, &sa, &corev1.ServiceAccount{}, sg, owner)
}
