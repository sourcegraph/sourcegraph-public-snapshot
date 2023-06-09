package events

type ClientInfo struct {
	SiteID            string `json:"site_id"`
	LicenseKey        string `json:"license_key"`
	InitialAdminEmail string `json:"initial_admin_email"`
	DeployType        string `json:"deploy_type"`
	Version           string `json:"Version"`
}

type TelemetryGatewayProxyRequest struct {
	Client ClientInfo
	Events []TelemetryEvent
}

type TelemetryEvent struct {
	EventName       string  `json:"name"`
	URL             string  `json:"url"`
	AnonymousUserID string  `json:"anonymous_user_id"`
	FirstSourceURL  string  `json:"first_source_url"`
	LastSourceURL   string  `json:"last_source_url"`
	UserID          int     `json:"user_id"`
	Source          string  `json:"source"`
	Timestamp       string  `json:"timestamp"`
	FeatureFlags    string  `json:"feature_flags"`
	CohortID        *string `json:"cohort_id,omitempty"`
	Referrer        string  `json:"referrer,omitempty"`
	PublicArgument  string  `json:"public_argument"`
	DeviceID        *string `json:"device_id,omitempty"`
	InsertID        *string `json:"insert_id,omitempty"`
}
