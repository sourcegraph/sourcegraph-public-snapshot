package reconciler

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/daemonset"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/role"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/rolebinding"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func (r *Reconciler) reconcileNodeExporter(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileNodeExporterRole(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileNodeExporterRoleBinding(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileNodeExporterService(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileNodeExporterDaemonSet(ctx, sg, owner); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileNodeExporterDaemonSet(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "node-exporter"
	cfg := sg.Spec.NodeExporter

	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: config.GetDefaultImage(sg, name),
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
		},
	})
	ctr.Args = []string{
		"--web.listen-address=:9100",
		"--path.sysfs=/host/sys",
		"--path.rootfs=/host/root",
		"--path.procfs=/host/proc",
		"--no-collector.wifi",
		"--no-collector.hwmon",
		"--collector.filesystem.ignored-mount-points=^/(dev|proc|sys|var/lib/docker/.+|var/lib/kubelet/pods/.+)($|/)",
		"--collector.netclass.ignored-devices=^(veth.*)$",
		"--collector.netdev.device-exclude=^(veth.*)$",
	}
	ctr.SecurityContext = &corev1.SecurityContext{
		RunAsUser:                pointers.Ptr[int64](65534),
		RunAsGroup:               pointers.Ptr[int64](65534),
		AllowPrivilegeEscalation: pointers.Ptr(false),
		ReadOnlyRootFilesystem:   pointers.Ptr(true),
	}
	ctr.Ports = []corev1.ContainerPort{
		{Name: "metrics", ContainerPort: 9100, Protocol: corev1.ProtocolTCP},
	}
	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "rootfs", MountPath: "/host/root", MountPropagation: pointers.Ptr(corev1.MountPropagationHostToContainer), ReadOnly: true},
		{Name: "sys", MountPath: "/host/sys", MountPropagation: pointers.Ptr(corev1.MountPropagationHostToContainer), ReadOnly: true},
		{Name: "proc", MountPath: "/host/proc", MountPropagation: pointers.Ptr(corev1.MountPropagationHostToContainer), ReadOnly: true},
	}

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port:   intstr.FromString("metrics"),
				Scheme: corev1.URISchemeHTTP},
		},
		InitialDelaySeconds: 0,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
		FailureThreshold:    3,
	}
	ctr.ReadinessProbe = probe
	ctr.LivenessProbe = probe

	template := pod.NewPodTemplate(name, cfg)
	template.Template.Spec.Containers = []corev1.Container{ctr}
	template.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser:    pointers.Ptr[int64](65534),
		RunAsGroup:   pointers.Ptr[int64](65534),
		FSGroup:      pointers.Ptr[int64](65534),
		RunAsNonRoot: pointers.Ptr(true),
	}

	template.Template.Spec.Volumes = []corev1.Volume{
		pod.NewVolumeHostPath("rootfs", "/"),
		pod.NewVolumeHostPath("sys", "/sys"),
		pod.NewVolumeHostPath("proc", "/proc"),
	}

	ds := daemonset.New(name, sg.Namespace, sg.Spec.RequestedVersion)
	ds.Spec.Template = template.Template

	return reconcileObject(ctx, r, sg.Spec.NodeExporter, &ds, &appsv1.DaemonSet{}, sg, owner)
}

func (r *Reconciler) reconcileNodeExporterService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	svc := service.NewService("node-exporter", sg.Namespace, sg.Spec.NodeExporter)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "metrics", TargetPort: intstr.FromString("metrics"), Port: 9100},
	}
	svc.Spec.Selector = map[string]string{"app": "node-exporter"}

	return reconcileObject(ctx, r, sg.Spec.NodeExporter, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileNodeExporterRole(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "node-exporter"
	cfg := sg.Spec.NodeExporter

	role := role.NewRole(name, sg.Namespace)

	readVerbs := []string{"use"}
	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups:     []string{"policy"},
			Resources:     []string{"podsecuritypolicies"},
			Verbs:         readVerbs,
			ResourceNames: []string{name},
		},
	}

	return reconcileObject(ctx, r, cfg, &role, &rbacv1.Role{}, sg, owner)
}

func (r *Reconciler) reconcileNodeExporterRoleBinding(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "node-exporter"
	binding := rolebinding.NewRoleBinding(name, sg.Namespace)
	binding.RoleRef = rbacv1.RoleRef{
		Kind: "ClusterRole",
		Name: name,
	}
	binding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      name,
			Namespace: sg.Namespace,
		},
	}
	return reconcileObject(ctx, r, sg.Spec.NodeExporter, &binding, &rbacv1.RoleBinding{}, sg, owner)
}
