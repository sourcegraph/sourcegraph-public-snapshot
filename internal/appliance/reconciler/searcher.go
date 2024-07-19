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

func (r *Reconciler) reconcileSearcher(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileSearcherStatefulSet(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling StatefulSet")
	}
	if err := r.reconcileSearcherService(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Service")
	}
	if err := r.reconcileSearcherServiceAccount(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ServiceAccount")
	}
	return nil
}

func (r *Reconciler) reconcileSearcherStatefulSet(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "searcher"
	cfg := sg.Spec.Searcher

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
	ctr.Env = append(ctr.Env, container.EnvVarsRedis()...)

	storageSize, err := resource.ParseQuantity(cfg.GetPersistentVolumeConfig().StorageSize)
	if err != nil {
		return errors.Wrap(err, "parsing storage size")
	}

	// Cache size is 90% of available attached storage
	cacheSize := float64(storageSize.Value()) * 0.9
	cacheSizeMB := int(math.Floor(cacheSize / 1024 / 1024))
	ctr.Env = append(
		ctr.Env,
		corev1.EnvVar{Name: "SEARCHER_CACHE_SIZE_MB", Value: fmt.Sprintf("%d", cacheSizeMB)},
		container.NewEnvVarFieldRef("POD_NAME", "metadata.name"),
		corev1.EnvVar{Name: "CACHE_DIR", Value: "/mnt/cache/$(POD_NAME)"},
	)
	ctr.Env = append(ctr.Env, container.EnvVarsOtel()...)

	ctr.Ports = []corev1.ContainerPort{
		{Name: "http", ContainerPort: 3181},
		{Name: "debug", ContainerPort: 6060},
	}
	ctr.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   intstr.FromString("http"),
				Scheme: corev1.URISchemeHTTP,
			},
		},
		PeriodSeconds:  5,
		TimeoutSeconds: 5,
	}
	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "cache", MountPath: "/mnt/cache"},
		{Name: "tmpdir", MountPath: "/tmp"},
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
	podTemplate.Template.Spec.Volumes = []corev1.Volume{
		{Name: "cache"},
		pod.NewVolumeEmptyDir("tmpdir"),
	}
	podTemplate.Template.Spec.ServiceAccountName = name

	pvc, err := pvc.NewPersistentVolumeClaim("cache", sg.Namespace, cfg)
	if err != nil {
		return err
	}

	sset := statefulset.NewStatefulSet(name, sg.Namespace, sg.Spec.RequestedVersion)
	sset.Spec.Template = podTemplate.Template
	sset.Spec.Replicas = &cfg.Replicas
	sset.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{pvc}

	ifChanged := struct {
		config.SearcherSpec
		RedisConnSpecs
	}{
		SearcherSpec:   cfg,
		RedisConnSpecs: redisConnSpecs,
	}
	return reconcileObject(ctx, r, ifChanged, &sset, &appsv1.StatefulSet{}, sg, owner)
}

func (r *Reconciler) reconcileSearcherService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "searcher"
	cfg := sg.Spec.Searcher
	svc := service.NewService(name, sg.Namespace, cfg)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", TargetPort: intstr.FromString("http"), Port: 3181},
		{Name: "debug", TargetPort: intstr.FromString("debug"), Port: 6060},
	}
	svc.Spec.Selector = map[string]string{
		"app": name,
	}
	return reconcileObject(ctx, r, cfg, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileSearcherServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.Searcher
	sa := serviceaccount.NewServiceAccount("searcher", sg.Namespace, cfg)
	return reconcileObject(ctx, r, cfg, &sa, &corev1.ServiceAccount{}, sg, owner)
}
