package reconciler

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/configmap"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pvc"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/secret"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/service"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/statefulset"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func (r *Reconciler) reconcileCodeInsights(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcileCodeInsightsStatefulSet(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileCodeInsightsPersistentVolumeClaim(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileCodeInsightsConfigMap(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileCodeInsightsSecret(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileCodeInsightsService(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcileCodeInsightsServiceAccount(ctx, sg, owner); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileCodeInsightsStatefulSet(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.CodeInsights
	name := "codeinsights-db"

	ctrImage := config.GetDefaultImage(sg, name)

	ctr := container.NewContainer("codeinsights", cfg, config.ContainerConfig{
		Image: ctrImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		},
	})
	ctr.SecurityContext = &corev1.SecurityContext{
		RunAsUser:                pointers.Ptr[int64](70),
		RunAsGroup:               pointers.Ptr[int64](70),
		AllowPrivilegeEscalation: pointers.Ptr(false),
		ReadOnlyRootFilesystem:   pointers.Ptr(true),
	}

	databaseSecretName := "codeinsights-db-auth"
	ctr.Env = append(ctr.Env, container.EnvVarsPostgres(databaseSecretName)...)
	ctr.Env = append(
		ctr.Env,
		corev1.EnvVar{Name: "PGDATA", Value: "/var/lib/postgresql/data/pgdata"},
		corev1.EnvVar{Name: "POSTGRESQL_CONF_DIR", Value: "/conf"},
	)
	ctr.Ports = []corev1.ContainerPort{{Name: name, ContainerPort: 5432}}
	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "disk", MountPath: "/var/lib/postgresql/data/"},
		{Name: "codeinsights-conf", MountPath: "/conf"},
		{Name: "lockdir", MountPath: "/var/run/postgresql"},
	}

	initCtrImage := config.GetDefaultImage(sg, "alpine-3.14")
	initCtr := container.NewContainer("correct-data-dir-permissions", cfg, config.ContainerConfig{
		Image: initCtrImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
	})
	initCtr.SecurityContext = &corev1.SecurityContext{
		RunAsUser:                pointers.Ptr[int64](70),
		RunAsGroup:               pointers.Ptr[int64](70),
		AllowPrivilegeEscalation: pointers.Ptr(false),
		ReadOnlyRootFilesystem:   pointers.Ptr(true),
	}
	initCtr.VolumeMounts = []corev1.VolumeMount{{Name: "disk", MountPath: "/var/lib/postgresql/data"}}
	initCtr.Command = []string{"sh", "-c", "if [ -d /var/lib/postgresql/data/pgdata ]; then chmod 750 /var/lib/postgresql/data/pgdata; fi"}

	pgExpCtrImage := config.GetDefaultImage(sg, "postgres_exporter")
	pgExpCtr := container.NewContainer("pgsql-exporter", cfg, config.ContainerConfig{
		Image: pgExpCtrImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
	})
	pgExpCtr.Env = append(pgExpCtr.Env, container.EnvVarsPostgresExporter(databaseSecretName)...)
	pgExpCtr.Env = append(pgExpCtr.Env, corev1.EnvVar{
		Name: "PG_EXPORTER_EXTEND_QUERY_PATH", Value: "/config/code_insights_queries.yaml",
	})

	podVolumes := []corev1.Volume{
		pod.NewVolumeFromPVC("disk", name),
		pod.NewVolumeFromConfigMap("codeinsights-conf", "codeinsights-db-conf"),
		pod.NewVolumeEmptyDir("lockdir"),
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.TerminationGracePeriodSeconds = pointers.Ptr[int64](120)
	podTemplate.Template.Spec.InitContainers = []corev1.Container{initCtr}
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr, pgExpCtr}
	podTemplate.Template.Spec.ServiceAccountName = name
	podTemplate.Template.Spec.Volumes = podVolumes
	podTemplate.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup:             pointers.Ptr[int64](70),
		RunAsUser:           pointers.Ptr[int64](70),
		RunAsGroup:          pointers.Ptr[int64](70),
		FSGroupChangePolicy: pointers.Ptr(corev1.FSGroupChangeOnRootMismatch),
	}

	sset := statefulset.NewStatefulSet(name, sg.Namespace, sg.Spec.RequestedVersion)
	sset.Spec.Template = podTemplate.Template

	return reconcileObject(ctx, r, sg.Spec.CodeInsights, &sset, &appsv1.StatefulSet{}, sg, owner)
}

func (r *Reconciler) reconcileCodeInsightsPersistentVolumeClaim(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.CodeInsights
	p, err := pvc.NewPersistentVolumeClaim("codeinsights-db", sg.Namespace, cfg)
	if err != nil {
		return err
	}
	return reconcileObject(ctx, r, sg.Spec.CodeInsights, &p, &corev1.PersistentVolumeClaim{}, sg, owner)
}

func (r *Reconciler) reconcileCodeInsightsConfigMap(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cm := configmap.NewConfigMap("codeinsights-db-conf", sg.Namespace)
	cm.Data = map[string]string{"postgresql.conf": string(config.CodeInsightsConfig)}

	return reconcileObject(ctx, r, sg.Spec.CodeInsights, &cm, &corev1.ConfigMap{}, sg, owner)
}

func (r *Reconciler) reconcileCodeInsightsSecret(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	scrt := secret.NewSecret("codeinsights-db-auth", sg.Namespace, sg.Spec.RequestedVersion)

	cn := sg.Spec.CodeInsights.DatabaseConnection
	scrt.Data = map[string][]byte{
		"host":     []byte(cn.Host),
		"port":     []byte(cn.Port),
		"user":     []byte(cn.User),
		"password": []byte(cn.Password),
		"database": []byte(cn.Database),
	}

	return reconcileObject(ctx, r, sg.Spec.CodeInsights, &scrt, &corev1.Secret{}, sg, owner)
}

func (r *Reconciler) reconcileCodeInsightsService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	name := "codeinsights-db"
	svc := service.NewService(name, sg.Namespace, sg.Spec.CodeInsights)
	svc.Spec.Ports = []corev1.ServicePort{{Name: name, TargetPort: intstr.FromString(name), Port: 5432}}
	svc.Spec.Selector = map[string]string{"app": name}

	return reconcileObject(ctx, r, sg.Spec.CodeInsights, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcileCodeInsightsServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.CodeInsights
	sa := serviceaccount.NewServiceAccount("codeinsights-db", sg.Namespace, cfg)
	return reconcileObject(ctx, r, sg.Spec.CodeInsights, &sa, &corev1.ServiceAccount{}, sg, owner)
}
