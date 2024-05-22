package reconciler

import (
	"bytes"
	"context"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/configmap"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/deployment"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/role"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/rolebinding"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *Reconciler) reconcilePrometheus(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcilePrometheusDeployment(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Deployment")
	}
	if err := r.reconcilePrometheusService(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Service")
	}
	if err := r.reconcilePrometheusServiceAccount(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ServiceAccount")
	}
	if err := r.reconcilePrometheusConfigMap(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ConfigMap")
	}
	if err := r.reconcilePrometheusRole(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Role")
	}
	if err := r.reconcilePrometheusRoleBinding(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling RoleBinding")
	}
	if err := r.reconcilePrometheusPVC(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling PVC")
	}
	return nil
}

func (r *Reconciler) reconcilePrometheusDeployment(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "prometheus"
	cfg := sg.Spec.Prometheus

	defaultImage, err := config.GetDefaultImage(sg, name)
	if err != nil {
		return err
	}
	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: defaultImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("6G"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("6G"),
			},
		},
	})
	ctr.Ports = []corev1.ContainerPort{
		{Name: "http", ContainerPort: 9090},
	}
	ctr.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/-/ready",
				Port: intstr.FromString("http"),
			},
		},
		TimeoutSeconds:   3,
		FailureThreshold: 120,
		PeriodSeconds:    5,
	}
	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "data", MountPath: "/prometheus"},
		{Name: "config", MountPath: "/sg_prometheus_add_ons"},
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}
	podTemplate.Template.Spec.Volumes = []corev1.Volume{
		pod.NewVolumeFromPVC("data", name),
		pod.NewVolumeFromConfigMap("config", name),
	}
	podTemplate.Template.Spec.ServiceAccountName = name

	dep := deployment.NewDeployment(name, sg.Namespace, sg.Spec.RequestedVersion)
	dep.Spec.Strategy = appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	}
	dep.Spec.Template = podTemplate.Template

	return reconcileObject(ctx, r, cfg, &dep, &appsv1.Deployment{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "prometheus"
	cfg := sg.Spec.Prometheus

	svc := service.NewService(name, sg.Namespace, cfg)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 30090, TargetPort: intstr.FromString("http")},
	}
	svc.Spec.Selector = map[string]string{
		"app": "syntect-server",
	}

	return reconcileObject(ctx, r, cfg, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.Prometheus
	sa := serviceaccount.NewServiceAccount("prometheus", sg.Namespace, cfg)
	return reconcileObject(ctx, r, cfg, &sa, &corev1.ServiceAccount{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusConfigMap(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	tmpl, err := template.New("prometheus-config").Parse(string(config.PrometheusDefaultConfigTemplate))
	if err != nil {
		return errors.Wrap(err, "parsing default prometheus config template")
	}
	var defaultConfig bytes.Buffer
	if err := tmpl.Execute(&defaultConfig, sg); err != nil {
		return errors.Wrap(err, "rendering default prometheus config template")
	}

	name := "prometheus"
	cfg := sg.Spec.Prometheus
	cm := configmap.NewConfigMap(name, sg.Namespace)
	cm.Data = map[string]string{
		"prometheus.yml":  defaultConfig.String(),
		"extra_rules.yml": "",
	}

	return reconcileObject(ctx, r, cfg, &cm, &corev1.ConfigMap{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusRole(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	// TODO privileged resources:
	// "namespaces",
	// "nodes",
	// "nodes/metrics",
	// "nodes/proxy",
	// nonresourceURLs

	role := role.NewRole("prometheus", sg.Namespace)
	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{
				"endpoints",
				"pods",
				"services",
			},
			Verbs: []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"configmap"},
			Verbs:     []string{"get"},
		},
	}
	return reconcileObject(ctx, r, sg.Spec.Prometheus, &role, &rbacv1.Role{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusRoleBinding(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "prometheus"
	binding := rolebinding.NewRoleBinding(name, sg.Namespace)
	binding.RoleRef = rbacv1.RoleRef{
		Kind: "Role",
		Name: name,
	}
	return reconcileObject(ctx, r, sg.Spec.Prometheus, &binding, &rbacv1.RoleBinding{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusPVC(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "prometheus"
	cfg := sg.Spec.Prometheus
	storageSize, err := resource.ParseQuantity(cfg.StorageSize)
	if err != nil {
		return errors.Wrap(err, "parsing storage size")
	}
	pvc := pvc.NewPersistentVolumeClaim(name, sg.Namespace, storageSize, sg.Spec.StorageClass.Name)
	return reconcileObject(ctx, r, cfg, &pvc, &corev1.PersistentVolumeClaim{}, sg, owner)
}
