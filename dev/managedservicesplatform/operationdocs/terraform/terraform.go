package terraform

import (
	"encoding/json"
	"io"
	"os"
)

// Monitoring stack cdktf json
type Monitoring struct {
	ResourceType ResourceType `json:"resource"`
}

// ResourceType is a terraform resource type e.g. `google_monitoring_alert_policy`
type ResourceType struct {
	GoogleMonitoringAlertPolicy map[string]AlertPolicy `json:"google_monitoring_alert_policy"`
}

// AlertPolicy is the configuration for an alert policy
type AlertPolicy struct {
	DisplayName   string        `json:"display_name,omitempty"`
	Documentation Documentation `json:"documentation"`
	Severity      string        `json:"severity"`
}

// Documentation is the markdown formatted documentation for an alert
type Documentation struct {
	Content string `json:"content"`
}

// ParseMonitoringCDKTF parses the generated terraform json
func ParseMonitoringCDKTF(path string) (*Monitoring, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var monitoring Monitoring
	err = json.Unmarshal(bytes, &monitoring)
	if err != nil {
		return nil, err
	}
	return &monitoring, nil
}
