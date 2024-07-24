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
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/statefulset"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *Reconciler) reconcileIndexedSearcher(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileIndexedSearcherStatefulSet(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling StatefulSet")
	}
	if err := r.reconcileIndexedSearcherService(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling Service")
	}
	if err := r.reconcileIndexedSearchIndexerService(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling indexer Service")
	}
	if err := r.reconcileIndexedSearcherServiceAccount(ctx, sg, owner); err != nil {
		return errors.Wrap(err, "reconciling ServiceAccount")
	}
	return nil
}

func (r *Reconciler) reconcileIndexedSearcherStatefulSet(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "indexed-search"
	cfg := sg.Spec.IndexedSearch

	webServer := container.NewContainer("zoekt-webserver", cfg, config.ContainerConfig{
		Image: config.GetDefaultImage(sg, "indexed-searcher"),
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("2G"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4G"),
			},
		},
	})
	webServer.Ports = []corev1.ContainerPort{
		{Name: "http", ContainerPort: 6070},
	}
	webServer.ReadinessProbe = &corev1.Probe{
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
	webServer.VolumeMounts = []corev1.VolumeMount{{Name: "data", MountPath: "/data"}}
	webServer.Env = append(webServer.Env, container.EnvVarsOtel()...)
	webServer.Env = append(webServer.Env, corev1.EnvVar{Name: "OPENTELEMETRY_DISABLED", Value: "false"})

	indexServer := container.NewContainer("zoekt-indexserver", cfg, config.ContainerConfig{
		Image: config.GetDefaultImage(sg, "search-indexer"),
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("4G"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("8"),
				corev1.ResourceMemory: resource.MustParse("8G"),
			},
		},
	})
	indexServer.Ports = []corev1.ContainerPort{
		{Name: "index-http", ContainerPort: 6072},
	}
	indexServer.VolumeMounts = []corev1.VolumeMount{{Name: "data", MountPath: "/data"}}
	indexServer.Env = append(indexServer.Env, container.EnvVarsOtel()...)
	indexServer.Env = append(indexServer.Env, corev1.EnvVar{Name: "OPENTELEMETRY_DISABLED", Value: "false"})

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.Containers = []corev1.Container{webServer, indexServer}
	podTemplate.Template.Spec.Volumes = []corev1.Volume{pod.NewVolumeFromPVC("data", "data")}
	podTemplate.Template.Spec.ServiceAccountName = name

	pvc, err := pvc.NewPersistentVolumeClaim("data", sg.Namespace, cfg)
	if err != nil {
		return err
	}

	sset := statefulset.NewStatefulSet(name, sg.Namespace, sg.Spec.RequestedVersion)
	sset.Spec.Template = podTemplate.Template
	sset.Spec.Replicas = &cfg.Replicas
	sset.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{pvc}

	return reconcileObject(ctx, r, cfg, &sset, &appsv1.StatefulSet{}, sg, owner)
}

func (r *Reconciler) reconcileIndexedSearcherService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "indexed-search"
	cfg := sg.Spec.IndexedSearch
	svc := service.NewService(name, sg.Namespace, cfg)
	svc.Spec.Ports = []corev1.ServicePort{{Port: 6070}}
	svc.Spec.Selector = map[string]string{
		"app": name,
	}

	return reconcileObject(ctx, r, cfg, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileIndexedSearchIndexerService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "indexed-search-indexer"
	cfg := sg.Spec.IndexedSearch
	svc := service.NewService(name, sg.Namespace, nil)
	svc.SetAnnotations(map[string]string{
		"prometheus.io/port":            "6072",
		"sourcegraph.prometheus/scrape": "true",
	})
	svc.Spec.Ports = []corev1.ServicePort{{Port: 6072, TargetPort: intstr.FromInt32(6072)}}
	svc.Spec.Selector = map[string]string{
		"app": "indexed-search",
	}

	return reconcileObject(ctx, r, cfg, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileIndexedSearcherServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.IndexedSearch
	sa := serviceaccount.NewServiceAccount("indexed-search", sg.Namespace, cfg)
	return reconcileObject(ctx, r, cfg, &sa, &corev1.ServiceAccount{}, sg, owner)
}
