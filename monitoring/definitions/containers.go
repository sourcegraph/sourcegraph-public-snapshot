package definitions

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Containers() *monitoring.Container {
	var (
		containerName = fmt.Sprintf("(%s)", strings.Join(images.DeploySourcegraphDockerImages, "|"))
	)

	containersNoAlertTransformer := func(observable shared.Observable) shared.Observable {
		return observable.WithNoAlerts(`Alerts are enabled in service-specific dashboard.`)
	}

	return &monitoring.Container{
		Name: "containers",
		Title: "Global Containers Resource Usage",
		Description: "Container usage and provisioning indicators of all services.",
		NoSourcegraphDebugServer: true,
		Groups: []monitoring.Group{
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerDevOps, &shared.ContainerMonitoringGroupOptions{
				ContainerMissing: containersNoAlertTransformer,
				CPUUsage: containersNoAlertTransformer,
				MemoryUsage: containersNoAlertTransformer,
				IOUsage: containersNoAlertTransformer,
			}),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerDevOps, &shared.ContainerProvisioningIndicatorsGroupOptions{
				LongTermCPUUsage: containersNoAlertTransformer,
				LongTermMemoryUsage: containersNoAlertTransformer,
				ShortTermCPUUsage: containersNoAlertTransformer,
				ShortTermMemoryUsage: containersNoAlertTransformer,
			}),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerDevOps, &shared.KubernetesMonitoringOptions{
				PodsAvailable: containersNoAlertTransformer,
			}),
		},
	}
}
