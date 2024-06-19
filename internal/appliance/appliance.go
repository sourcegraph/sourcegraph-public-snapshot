package appliance

import (
	"context"

	"github.com/sourcegraph/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	pb "github.com/sourcegraph/sourcegraph/internal/appliance/v1"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Appliance struct {
	client      client.Client
	status      Status
	sourcegraph config.Sourcegraph
	logger      log.Logger

	// Embed the UnimplementedApplianceServiceServer structs to ensure forwards compatibility (if the service is
	// compiled against a newer version of the proto file, the server will still have default implementations of any new
	// RPCs).
	pb.UnimplementedApplianceServiceServer
}

// Status is a Stage that an Appliance can be in.
type Status string

const (
	StatusUnknown    Status = "unknown"
	StatusSetup      Status = "setup"
	StatusInstalling Status = "installing"
)

func (s Status) String() string {
	return string(s)
}

func NewAppliance(client client.Client, logger log.Logger) *Appliance {
	return &Appliance{
		client:      client,
		status:      StatusSetup,
		sourcegraph: config.Sourcegraph{},
		logger:      logger,
	}
}

func (a *Appliance) GetCurrentVersion(ctx context.Context) string {
	return a.sourcegraph.Status.CurrentVersion
}

func (a *Appliance) GetCurrentStatus(ctx context.Context) Status {
	return a.status
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

	if err := a.client.Create(ctx, configMap); err != nil {
		return nil, err
	}

	return configMap, nil
}

func (a *Appliance) GetConfigMap(ctx context.Context, name, namespace string) (*corev1.ConfigMap, error) {
	var applianceSpec corev1.ConfigMap
	err := a.client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &applianceSpec)
	if apierrors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &applianceSpec, nil
}

func (a *Appliance) shouldSetupRun(ctx context.Context) (bool, error) {
	cfgMap, err := a.GetConfigMap(ctx, "sourcegraph-appliance", "default") //TODO unhardcode and load namespace properly
	switch {
	case err != nil:
		return false, err
	case a.status == StatusInstalling:
		// configMap does not exist but is being created
		return false, nil
	case cfgMap == nil:
		// configMap does not exist
		return true, nil
	case cfgMap.Annotations[config.AnnotationKeyManaged] == "false":
		// appliance is not managed
		return false, nil
	default:
		return true, nil
	}
}
