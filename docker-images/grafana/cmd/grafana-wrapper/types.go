package main

// exported from grafana@v6.7.4
// https://sourcegraph.com/github.com/grafana/grafana@v6.7.4/-/blob/pkg/services/provisioning/notifiers/types.go

type notificationsAsConfig struct {
	Notifications       []*notificationFromConfig   `json:"notifiers" yaml:"notifiers"`
	DeleteNotifications []*deleteNotificationConfig `json:"delete_notifiers" yaml:"delete_notifiers"`
}

type deleteNotificationConfig struct {
	Uid     string `json:"uid" yaml:"uid"`
	Name    string `json:"name" yaml:"name"`
	OrgId   int64  `json:"org_id" yaml:"org_id"`
	OrgName string `json:"org_name" yaml:"org_name"`
}

type notificationFromConfig struct {
	Uid                   string                 `json:"uid" yaml:"uid"`
	OrgId                 int64                  `json:"org_id" yaml:"org_id"`
	OrgName               string                 `json:"org_name" yaml:"org_name"`
	Name                  string                 `json:"name" yaml:"name"`
	Type                  string                 `json:"type" yaml:"type"`
	SendReminder          bool                   `json:"send_reminder" yaml:"send_reminder"`
	DisableResolveMessage bool                   `json:"disable_resolve_message" yaml:"disable_resolve_message"`
	Frequency             string                 `json:"frequency" yaml:"frequency"`
	IsDefault             bool                   `json:"is_default" yaml:"is_default"`
	Settings              map[string]interface{} `json:"settings" yaml:"settings"`
}
