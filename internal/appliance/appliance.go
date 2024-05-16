package appliance

import (
	"context"

	"github.com/Masterminds/semver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/appliance/reconciler"
	pb "github.com/sourcegraph/sourcegraph/internal/appliance/v1"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Appliance struct {
	client client.Client

	version     semver.Version
	status      Status
	sourcegraph reconciler.Sourcegraph

	// Embed the UnimplementedApplianceServiceServer structs to ensure forwards compatibility (if the service is
	// compiled against a newer version of the proto file, the server will still have default implementations of any new
	// RPCs).
	pb.UnimplementedApplianceServiceServer
}

type Status struct {
	stage Stage
}

// Stage is a Stage that an Appliance can be in.
type Stage string

const (
	StageUnknown         Stage = "unknown"
	StageIdle            Stage = "idle"
	StageInstall         Stage = "install"
	StageInstalling      Stage = "installing"
	StageUpgrading       Stage = "upgrading"
	StageWaitingForAdmin Stage = "waitingForAdmin"
	StageRefresh         Stage = "refresh"
)

func (s Stage) String() string {
	return string(s)
}

func NewAppliance(client client.Client) *Appliance {
	return &Appliance{
		client: client,
		status: Status{
			stage: StageUnknown,
		},
		sourcegraph: reconciler.Sourcegraph{},
	}
}

func (a *Appliance) GetCurrentVersion(ctx context.Context) semver.Version {
	return a.version
}

func (a *Appliance) GetCurrentStage(ctx context.Context) Stage {
	return a.status.stage
}

func (a *Appliance) CreateConfigMap(ctx context.Context, name, namespace string) (*corev1.ConfigMap, error) {
	spec, err := yaml.Marshal(a.sourcegraph)
	if err != nil {
		return nil, err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
			Annotations: map[string]string{
				// required annotation for our controller filter.
				config.AnnotationKeyManaged: "true",
			},
		},
		Immutable: pointers.Ptr(false),
		Data: map[string]string{
			"spec": string(spec),
		},
	}

	return configMap, nil
}
