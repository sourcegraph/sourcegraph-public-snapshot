package reconciler

import (
	"context"
	"fmt"
	"math"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/statefulset"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *Reconciler) reconcileSymbols(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileSymbolsStatefulSet(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileSymbolsService(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileSymbolsServiceAccount(ctx, sg, owner); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileSymbolsStatefulSet(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "symbols"
	cfg := sg.Spec.Symbols

	defaultImage := config.GetDefaultImage(sg, name)
	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: defaultImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("500M"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("2G"),
			},
		},
	})

	storageSize, err := resource.ParseQuantity(cfg.GetPersistentVolumeConfig().StorageSize)
	if err != nil {
		return errors.Wrap(err, "parsing storage size")
	}

	// Cache size is 90% of available attached storage
	cacheSize := float64(storageSize.Value()) * 0.9
	cacheSizeMB := int(math.Floor(cacheSize / 1024 / 1024))

	ctr.Env = append(ctr.Env, container.EnvVarsRedis()...)
	ctr.Env = append(
		ctr.Env,
		corev1.EnvVar{Name: "SYMBOLS_CACHE_SIZE_MB", Value: fmt.Sprintf("%d", cacheSizeMB)},

		container.NewEnvVarFieldRef("POD_NAME", "metadata.name"),
		corev1.EnvVar{Name: "SYMBOLS_CACHE_DIR", Value: "/mnt/cache/$(POD_NAME)"},

		corev1.EnvVar{Name: "TMPDIR", Value: "/mnt/tmp"},
	)
	ctr.Env = append(ctr.Env, container.EnvVarsOtel()...)

	ctr.Ports = []corev1.ContainerPort{
		{Name: "http", ContainerPort: 3184},
		{Name: "debug", ContainerPort: 6060},
	}
	ctr.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   intstr.FromString("http"),
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 60,
		TimeoutSeconds:      5,
	}
	ctr.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   intstr.FromString("http"),
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 60,
		PeriodSeconds:       5,
	}
	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "cache", MountPath: "/mnt/cache"},
		{Name: "tmp", MountPath: "/mnt/tmp"},
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	redisConnSpecs, err := r.getRedisSecrets(ctx, sg)
	if err != nil {
		return err
	}
	redisConnHash, err := configHash(redisConnSpecs)
	if err != nil {
		return err
	}
	podTemplate.Template.ObjectMeta.Annotations["checksum/redis"] = redisConnHash

	podTemplate.Template.Spec.Containers = []corev1.Container{ctr}
	podTemplate.Template.Spec.ServiceAccountName = name
	podTemplate.Template.Spec.Volumes = []corev1.Volume{
		{Name: "cache"},
		pod.NewVolumeEmptyDir("tmp"),
	}

	pvc, err := pvc.NewPersistentVolumeClaim("cache", sg.Namespace, cfg)
	if err != nil {
		return err
	}

	sset := statefulset.NewStatefulSet(name, sg.Namespace, sg.Spec.RequestedVersion)
	sset.Spec.Template = podTemplate.Template
	sset.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{pvc}

	ifChanged := struct {
		config.SymbolsSpec
		RedisConnSpecs
	}{
		SymbolsSpec:    cfg,
		RedisConnSpecs: redisConnSpecs,
	}
	return reconcileObject(ctx, r, ifChanged, &sset, &appsv1.StatefulSet{}, sg, owner)
}

func (r *Reconciler) reconcileSymbolsService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	svc := service.NewService("symbols", sg.Namespace, sg.Spec.Symbols)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", TargetPort: intstr.FromString("http"), Port: 3184},
		{Name: "debug", TargetPort: intstr.FromString("debug"), Port: 6060},
	}
	svc.Spec.Selector = map[string]string{
		"app": "symbols",
	}
	return reconcileObject(ctx, r, sg.Spec.Symbols, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileSymbolsServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.Symbols
	sa := serviceaccount.NewServiceAccount("symbols", sg.Namespace, cfg)
	return reconcileObject(ctx, r, sg.Spec.Symbols, &sa, &corev1.ServiceAccount{}, sg, owner)
}
