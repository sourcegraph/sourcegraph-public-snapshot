package cloud

import (
	"fmt"
	"time"

	cloudapiv1 "github.com/sourcegraph/cloud-api/go/cloudapi/v1"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Instance struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	InstanceType string `json:"instanceType"`
	Environment  string `json:"environment"`
	Version      string `json:"version"`
	Hostname     string `json:"hostname"`
	AdminEmail   string `json:"adminEmail"`

	CreatedAt time.Time `json:"createdAt"`
	DeletedAt time.Time `json:"deletedAt"`

	Project string `json:"project"`
	Region  string `json:"region"`
	Status  string `json:"status"`
}

func (i *Instance) String() string {
	// TODO(burmudar): use formatting spacing
	return fmt.Sprintf(`ID           : %s
Name         : %s
InstanceType : %s
Environment  : %s
Version      : %s
Hostname     : %s
AdminEmail   : %s
CreatedAt    : %s
Project      : %s
Region       : %s
DeletetAt    : %s
Status       : %s
`, i.ID, i.Name, i.InstanceType, i.Environment, i.Version, i.Hostname, i.AdminEmail,
		i.CreatedAt, i.Project, i.Region, i.DeletedAt,
		i.Status)
}
func newInstance(src *cloudapiv1.Instance) *Instance {
	details := src.GetInstanceDetails()
	platform := src.GetPlatformDetails()
	return &Instance{
		ID:           src.GetId(),
		Name:         details.Name,
		InstanceType: EphemeralInstanceType,
		Version:      details.Version,
		Hostname:     pointers.DerefZero(details.Hostname),
		AdminEmail:   pointers.DerefZero(details.AdminEmail),
		CreatedAt:    platform.GetCreatedAt().AsTime(),
		DeletedAt:    platform.GetDeletedAt().AsTime(),
		Project:      platform.GetGcpProjectId(),
		Region:       platform.GetGcpRegion(),
		Status:       resolveStatus(src.GetInstanceState()),
	}
}

func resolveStatus(i *cloudapiv1.InstanceState) string {
	switch i.InstanceStatus {
	case cloudapiv1.InstanceStatus_INSTANCE_STATUS_UNSPECIFIED:
		return "unspecified"
	case cloudapiv1.InstanceStatus_INSTANCE_STATUS_OK:
		return "ok"
	case cloudapiv1.InstanceStatus_INSTANCE_STATUS_PROGRESSING:
		return "in progress"
	case cloudapiv1.InstanceStatus_INSTANCE_STATUS_FAILED:
		return "failed"
	default:
		return "unknown"
	}
}

func toInstances(items ...*cloudapiv1.Instance) []*Instance {
	converted := []*Instance{}
	for _, item := range items {
		converted = append(converted, newInstance(item))
	}
	return converted
}
