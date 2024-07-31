package config

// Status is a point in the Appliance lifecycle that an Appliance can be in.
type Status string

func (s Status) String() string {
	return string(s)
}

const (
	ConfigmapName = "sourcegraph-appliance"

	AnnotationKeyManaged             = "appliance.sourcegraph.com/managed"
	AnnotationConditions             = "appliance.sourcegraph.com/conditions"
	AnnotationKeyCurrentVersion      = "appliance.sourcegraph.com/currentVersion"
	AnnotationKeyConfigHash          = "appliance.sourcegraph.com/configHash"
	AnnotationKeyShouldTakeOwnership = "appliance.sourcegraph.com/adopted"

	// TODO set status on configmap to communicate it across reboots
	AnnotationKeyStatus = "appliance.sourcegraph.com/status"

	StatusUnknown         Status = "unknown"
	StatusInstall         Status = "install"
	StatusInstalling      Status = "installing"
	StatusUpgrading       Status = "upgrading"
	StatusWaitingForAdmin Status = "wait-for-admin"
	StatusRefresh         Status = "refresh"
	StatusMaintenance     Status = "maintenance"
)

func IsPostInstallStatus(status Status) bool {
	switch status {
	case StatusUnknown, StatusInstall, StatusInstalling, StatusWaitingForAdmin:
		return false
	}
	return true
}
