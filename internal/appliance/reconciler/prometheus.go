package reconciler

import (
	"bytes"
	"context"
	"fmt"
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
	if err := r.reconcilePrometheusPVC(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling PVC")
	}

	if sg.Spec.Prometheus.Privileged {
		if err := r.reconcilePrometheusClusterRoleBinding(ctx, sg, owner); err != nil {
			return errors.Wrap(err, "reconciling ClusterRoleBinding")
		}
	} else {
		if err := r.reconcilePrometheusRoleBinding(ctx, sg, owner); err != nil {
			return errors.Wrap(err, "reconciling RoleBinding")
		}
	}
	return nil
}

func (r *Reconciler) reconcilePrometheusDeployment(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "prometheus"
	cfg := sg.Spec.Prometheus

	defaultImage := config.GetDefaultImage(sg, name)
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

	cfgMapName := name
	if cfg.ExistingConfigMap != "" {
		cfgMapName = cfg.ExistingConfigMap
	}
	podTemplate.Template.Spec.Volumes = []corev1.Volume{
		pod.NewVolumeFromPVC("data", name),
		pod.NewVolumeFromConfigMap("config", cfgMapName),
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
	cfg := sg.Spec.Prometheus
	if cfg.ExistingConfigMap != "" {
		return nil
	}

	tmpl, err := template.New("prometheus-config").Parse(string(config.PrometheusDefaultConfigTemplate))
	if err != nil {
		return errors.Wrap(err, "parsing default prometheus config template")
	}
	var defaultConfig bytes.Buffer
	if err := tmpl.Execute(&defaultConfig, sg); err != nil {
		return errors.Wrap(err, "rendering default prometheus config template")
	}

	name := "prometheus"
	cm := configmap.NewConfigMap(name, sg.Namespace)
	cm.Data = map[string]string{
		"prometheus.yml":  defaultConfig.String(),
		"extra_rules.yml": "",
	}

	return reconcileObject(ctx, r, cfg, &cm, &corev1.ConfigMap{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusRole(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "prometheus"
	cfg := sg.Spec.Prometheus

	resources := []string{
		"endpoints",
		"pods",
		"services",
	}
	if cfg.Privileged {
		resources = append(
			resources,
			"namespaces",
			"nodes",
			"nodes/metrics",
			"nodes/proxy",
		)
	}
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: resources,
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
			Verbs:     []string{"get"},
		},
	}
	if cfg.Privileged {
		rules = append(rules, rbacv1.PolicyRule{
			NonResourceURLs: []string{"/metrics"},
			Verbs:           []string{"get"},
		})

		// Make resource name sg-specific since this is a non-namespaced
		// (cluster-scoped) object
		name := fmt.Sprintf("%s-%s", sg.Namespace, "prometheus")
		role := role.NewClusterRole(name, sg.Namespace)
		role.Rules = rules
		return reconcileObject(ctx, r, cfg, &role, &rbacv1.ClusterRole{}, sg, owner)
	}

	role := role.NewRole(name, sg.Namespace)
	role.Rules = rules
	return reconcileObject(ctx, r, cfg, &role, &rbacv1.Role{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusRoleBinding(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "prometheus"
	binding := rolebinding.NewRoleBinding(name, sg.Namespace)
	binding.RoleRef = rbacv1.RoleRef{
		Kind: "Role",
		Name: name,
	}
	binding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "prometheus",
			Namespace: sg.Namespace,
		},
	}
	return reconcileObject(ctx, r, sg.Spec.Prometheus, &binding, &rbacv1.RoleBinding{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusClusterRoleBinding(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	// Make resource name sg-specific since this is a non-namespaced
	// (cluster-scoped) object
	name := fmt.Sprintf("%s-%s", sg.Namespace, "prometheus")
	binding := rolebinding.NewClusterRoleBinding(name, sg.Namespace)
	binding.RoleRef = rbacv1.RoleRef{
		Kind: "ClusterRole",
		Name: name,
	}
	binding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "prometheus",
			Namespace: sg.Namespace,
		},
	}
	return reconcileObject(ctx, r, sg.Spec.Prometheus, &binding, &rbacv1.ClusterRoleBinding{}, sg, owner)
}

func (r *Reconciler) reconcilePrometheusPVC(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "prometheus"
	cfg := sg.Spec.Prometheus
	pvc, err := pvc.NewPersistentVolumeClaim(name, sg.Namespace, cfg)
	if err != nil {
		return err
	}
	return reconcileObject(ctx, r, cfg, &pvc, &corev1.PersistentVolumeClaim{}, sg, owner)
}
