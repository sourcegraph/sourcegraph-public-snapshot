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

func (r *Reconciler) reconcilePGSQL(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	if err := r.reconcilePGSQLStatefulSet(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcilePGSQLPersistentVolumeClaim(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcilePGSQLConfigMap(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcilePGSQLSecret(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcilePGSQLService(ctx, sg, owner); err != nil {
		return err
	}
	if err := r.reconcilePGSQLServiceAccount(ctx, sg, owner); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcilePGSQLStatefulSet(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.PGSQL
	name := "pgsql"

	ctrImage := config.GetDefaultImage(sg, "postgres-12-alpine")
	ctr := container.NewContainer(name, cfg, config.ContainerConfig{
		Image: ctrImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
		},
	})
	ctr.SecurityContext = &corev1.SecurityContext{
		RunAsUser:                pointers.Ptr[int64](999),
		RunAsGroup:               pointers.Ptr[int64](999),
		AllowPrivilegeEscalation: pointers.Ptr(false),
		ReadOnlyRootFilesystem:   pointers.Ptr(true),
	}

	databaseSecretName := "pgsql-auth"
	ctr.Env = append(ctr.Env, container.EnvVarsPostgres(databaseSecretName)...)
	ctr.Ports = []corev1.ContainerPort{{Name: name, ContainerPort: 5432}}
	ctr.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/liveness.sh"},
			},
		},
		InitialDelaySeconds: 15,
	}
	ctr.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/ready.sh"},
			},
		},
	}
	ctr.StartupProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/liveness.sh"},
			},
		},
		FailureThreshold: 360,
		PeriodSeconds:    10,
	}
	ctr.VolumeMounts = []corev1.VolumeMount{
		{Name: "disk", MountPath: "/data"},
		{Name: "pgsql-conf", MountPath: "/conf"},
		{Name: "dshm", MountPath: "/dev/shm"},
		{Name: "lockdir", MountPath: "/var/run/postgresql"},
	}

	initCtrImage := config.GetDefaultImage(sg, "alpine-3.14")
	initCtr := container.NewContainer("correct-data-dir-permissions", cfg, config.ContainerConfig{
		Image: initCtrImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50M"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50M"),
			},
		},
	})
	initCtr.SecurityContext = &corev1.SecurityContext{
		RunAsUser:                pointers.Ptr[int64](999),
		RunAsGroup:               pointers.Ptr[int64](999),
		AllowPrivilegeEscalation: pointers.Ptr(false),
		ReadOnlyRootFilesystem:   pointers.Ptr(true),
	}
	initCtr.VolumeMounts = []corev1.VolumeMount{{Name: "disk", MountPath: "/data"}}
	initCtr.Command = []string{"sh", "-c", "if [ -d /data/pgdata-12 ]; then chmod 750 /data/pgdata-12; fi"}

	pgExpCtrImage := config.GetDefaultImage(sg, "postgres_exporter")
	pgExpCtr := container.NewContainer("pgsql-exporter", cfg, config.ContainerConfig{
		Image: pgExpCtrImage,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50M"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50M"),
			},
		},
	})
	pgExpCtr.SecurityContext = &corev1.SecurityContext{
		RunAsUser:                pointers.Ptr[int64](999),
		RunAsGroup:               pointers.Ptr[int64](999),
		AllowPrivilegeEscalation: pointers.Ptr(false),
		ReadOnlyRootFilesystem:   pointers.Ptr(true),
	}
	pgExpCtr.Env = append(pgExpCtr.Env, container.EnvVarsPostgresExporter(databaseSecretName)...)
	pgExpCtr.Env = append(pgExpCtr.Env, corev1.EnvVar{
		Name: "PG_EXPORTER_EXTEND_QUERY_PATH", Value: "/config/queries.yaml",
	})

	podVolumes := []corev1.Volume{
		pod.NewVolumeEmptyDir("lockdir"),
		{Name: "dshm", VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium:    corev1.StorageMediumMemory,
				SizeLimit: pointers.Ptr(resource.MustParse("1Gi")),
			},
		}},
		{Name: "disk", VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: "pgsql",
			},
		}},
		{Name: "pgsql-conf", VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				DefaultMode: pointers.Ptr[int32](0777),
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "pgsql-conf",
				},
			},
		}},
	}

	podTemplate := pod.NewPodTemplate(name, cfg)
	podTemplate.Template.Spec.TerminationGracePeriodSeconds = pointers.Ptr[int64](120)
	podTemplate.Template.Spec.InitContainers = []corev1.Container{initCtr}
	podTemplate.Template.Spec.Containers = []corev1.Container{ctr, pgExpCtr}
	podTemplate.Template.Spec.ServiceAccountName = name
	podTemplate.Template.Spec.Volumes = podVolumes
	podTemplate.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup:             pointers.Ptr[int64](999),
		RunAsUser:           pointers.Ptr[int64](999),
		RunAsGroup:          pointers.Ptr[int64](999),
		FSGroupChangePolicy: pointers.Ptr(corev1.FSGroupChangeOnRootMismatch),
	}

	sset := statefulset.NewStatefulSet(name, sg.Namespace, sg.Spec.RequestedVersion)
	sset.Spec.Template = podTemplate.Template

	return reconcileObject(ctx, r, sg.Spec.PGSQL, &sset, &appsv1.StatefulSet{}, sg, owner)
}

func (r *Reconciler) reconcilePGSQLPersistentVolumeClaim(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.PGSQL
	p, err := pvc.NewPersistentVolumeClaim("pgsql", sg.Namespace, cfg)
	if err != nil {
		return err
	}
	return reconcileObject(ctx, r, sg.Spec.PGSQL, &p, &corev1.PersistentVolumeClaim{}, sg, owner)
}

func (r *Reconciler) reconcilePGSQLConfigMap(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cm := configmap.NewConfigMap("pgsql-conf", sg.Namespace)
	cm.Data = map[string]string{"postgresql.conf": string(config.PgsqlConfig)}

	return reconcileObject(ctx, r, sg.Spec.PGSQL, &cm, &corev1.ConfigMap{}, sg, owner)
}

func (r *Reconciler) reconcilePGSQLSecret(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	scrt := secret.NewSecret("pgsql-auth", sg.Namespace, sg.Spec.RequestedVersion)

	cn := sg.Spec.PGSQL.DatabaseConnection
	scrt.Data = map[string][]byte{
		"host":     []byte(cn.Host),
		"port":     []byte(cn.Port),
		"user":     []byte(cn.User),
		"password": []byte(cn.Password),
		"database": []byte(cn.Database),
	}

	return reconcileObject(ctx, r, sg.Spec.PGSQL, &scrt, &corev1.Secret{}, sg, owner)
}

func (r *Reconciler) reconcilePGSQLService(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	svc := service.NewService("pgsql", sg.Namespace, sg.Spec.PGSQL)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "pgsql", TargetPort: intstr.FromString("pgsql"), Port: 5432},
	}
	svc.Spec.Selector = map[string]string{"app": "pgsql"}

	return reconcileObject(ctx, r, sg.Spec.PGSQL, &svc, &corev1.Service{}, sg, owner)
}

func (r *Reconciler) reconcilePGSQLServiceAccount(ctx context.Context, sg *config.Sourcegraph, owner client.Object) error {
	cfg := sg.Spec.PGSQL
	sa := serviceaccount.NewServiceAccount("pgsql", sg.Namespace, cfg)
	return reconcileObject(ctx, r, sg.Spec.PGSQL, &sa, &corev1.ServiceAccount{}, sg, owner)
}
