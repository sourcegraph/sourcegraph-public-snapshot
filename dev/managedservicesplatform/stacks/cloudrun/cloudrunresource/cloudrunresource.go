package cloudrunresource

import "fmt"

// NewName is the stable representation of a Cloud Run resource. It must be
// unique across environments.
func NewName(serviceID, environmentID, gcpRegion string) string {
	return fmt.Sprintf("%s-%s-%s",
		serviceID, environmentID, gcpRegion)
}
