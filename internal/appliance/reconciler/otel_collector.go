package reconciler

import (
	"bytes"
	"context"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/configmap"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/deployment"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *Reconciler) reconcileOtelCollector(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileOtelCollectorDeployment(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Deployment")
	}
	if err := r.reconcileOtelCollectorService(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Service")
	}
	if err := r.reconcileOtelCollectorServiceAccount(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ServiceAccount")
	}
	if err := r.reconcileOtelCollectorConfigMap(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ServiceAccount")
	}
	return nil
}

func (r *Reconciler) reconcileOtelCollectorDeployment(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "otel-collector"
	cfg := sg.Spec.OtelCollector

	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: config.GetDefaultImage(sg, "opentelemetry-collector"),
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3"),
				corev1.ResourceMemory: resource.MustParse("3Gi"),
			},
		},
	})

	configFile := "/etc/otel-collector/configs/logging.yaml"
	if len(cfg.Exporters) > 0 {
		configFile = "/etc/otel-collector/config.yaml"
	}
	if !sg.Spec.Jaeger.IsDisabled() {
		configFile = "/etc/otel-collector/configs/jaeger.yaml"
		ctr.Env = append(
			ctr.Env,
			corev1.EnvVar{Name: "JAEGER_HOST", Value: "jaeger-collector"},
			corev1.EnvVar{Name: "JAEGER_OTLP_GRPC_PORT", Value: "4320"},
			corev1.EnvVar{Name: "JAEGER_OTLP_HTTP_PORT", Value: "4321"},
		)
	}

	ctr.Command = []string{
		"/bin/otelcol-sourcegraph",
		"--config=" + configFile,
	}
	ctr.Ports = []corev1.ContainerPort{
		{Name: "zpages", ContainerPort: 55679},
		{Name: "otlp-grpc", ContainerPort: 4317},
		{Name: "otlp-http", ContainerPort: 4318},
		{Name: "metrics", ContainerPort: 8888},
	}

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(13133)},
		},
	}
	ctr.LivenessProbe = probe
	ctr.ReadinessProbe = probe

	if len(cfg.Exporters) > 0 {
		ctr.VolumeMounts = []corev1.VolumeMount{
			{Name: "config", MountPath: "/etc/otel-collector"},
		}
	}
	if cfg.ExportersTLSSecretName != "" {
		ctr.VolumeMounts = append(ctr.VolumeMounts, corev1.VolumeMount{Name: "otel-collector-tls", MountPath: "/tls"})
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}
	podTemplate.Template.Spec.ServiceAccountName = name

	if len(cfg.Exporters) > 0 {
		vol := pod.NewVolumeFromConfigMap("config", name)
		vol.ConfigMap.Items = []corev1.KeyToPath{
			{Key: "config.yaml", Path: "config.yaml"},
		}
		podTemplate.Template.Spec.Volumes = []corev1.Volume{vol}
	}
	if cfg.ExportersTLSSecretName != "" {
		podTemplate.Template.Spec.Volumes = append(
			podTemplate.Template.Spec.Volumes,
			pod.NewVolumeFromSecret("otel-collector-tls", cfg.ExportersTLSSecretName),
		)
	}

	dep := deployment.NewDeployment(name, sg.Namespace, sg.Spec.RequestedVersion)
	dep.Spec.Template = podTemplate.Template

	return reconcileObject(ctx, r, cfg, &dep, &appsv1.Deployment{}, sg, owner)
}

func (r *Reconciler) reconcileOtelCollectorService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	svc := service.NewService("otel-collector", sg.Namespace, sg.Spec.OtelCollector)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "otlp-grpc", TargetPort: intstr.FromInt(4317), Port: 4317},
		{Name: "otlp-http", TargetPort: intstr.FromInt(4318), Port: 4318},
		{Name: "metrics", TargetPort: intstr.FromInt(8888), Port: 8888},
	}
	svc.Spec.Selector = map[string]string{
		"app": "otel-collector",
	}

	return reconcileObject(ctx, r, sg.Spec.OtelCollector, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileOtelCollectorServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.OtelCollector
	sa := serviceaccount.NewServiceAccount("otel-collector", sg.Namespace, cfg)
	return reconcileObject(ctx, r, sg.Spec.OtelCollector, &sa, &corev1.ServiceAccount{}, sg, owner)
}

func (r *Reconciler) reconcileOtelCollectorConfigMap(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "otel-collector"
	cfg := sg.Spec.OtelCollector

	cm := configmap.NewConfigMap(name, sg.Namespace)

	// We only deploy collector config when jaeger is disabled.
	if !sg.Spec.Jaeger.IsDisabled() {
		return ensureObjectDeleted(ctx, r, owner, &cm)
	}

	tmpl, err := template.New("otel-collector-config").
		Funcs(config.TemplateFuncMap).
		Parse(string(config.OtelCollectorConfigTemplate))
	if err != nil {
		return errors.Wrap(err, "parsing otel-collector config template")
	}
	var collectorCfg bytes.Buffer
	if err := tmpl.Execute(&collectorCfg, cfg); err != nil {
		return errors.Wrap(err, "rendering otel-collector config template")
	}
	cm.Data = map[string]string{
		"config.yaml": collectorCfg.String(),
	}

	return reconcileObject(ctx, r, cfg, &cm, &corev1.ConfigMap{}, sg, owner)
}
