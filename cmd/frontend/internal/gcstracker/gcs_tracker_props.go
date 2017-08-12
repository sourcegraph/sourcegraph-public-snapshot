package gcstracker

// TrackedObjects represents a transmission of data objects to be tracked and stored
type TrackedObjects struct {
	DeviceInfo *DeviceInfo      `json:"device_info,omitempty"`
	Header     *Header          `json:"header,omitempty"`
	Objects    []*TrackedObject `json:"objects,omitempty"`
	UserInfo   *UserInfo        `json:"user_info,omitempty"`
	BatchID    string           `json:"batch_id,omitempty"`
}

// Header represents environment-level properties
type Header struct {
	AppID        string `json:"app_id,omitempty"`
	Env          string `json:"env,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	ServerTstamp int64  `json:"server_tstamp"`
	// Event reflects the frontend event (i.e., event_label) that triggered the object tracking
	Event string `json:"event_label"`
}

// DeviceInfo represents platform- and device-level properties
type DeviceInfo struct {
	Platform         string `json:"platform,omitempty"`
	TrackerNamespace string `json:"tracker_namespace,omitempty"`
}

// UserInfo represents user-level properties
type UserInfo struct {
	BusinessUserID string `json:"business_user_id"`
	Email          string `json:"email"`
}

// TrackedObject represents a user data object to be tracked and stored
type TrackedObject struct {
	ObjectID string      `json:"object_id,omitempty"`
	Type     string      `json:"type,omitempty"`
	Ctx      interface{} `json:"ctx,omitempty"`
}

type OrgWithDetailsContext struct {
	OrgName string       `json:"name,omitempty"`
	Members []*OrgMember `json:"members,omitempty"`
}

type OrgMember struct {
	MemberUserID string `json:"user_id,omitempty"`
}

// GitHubInstallationEvent is metadata associated with a GitHub app
// installation event received by a webhook
type GitHubInstallationEvent struct {
	Action       string              `json:"action"`
	Installation *GitHubInstallation `json:"installation"`
	Sender       *GitHubAccount      `json:"sender"`
}

// GitHubRepositoriesEvent is metadata associated with a GitHub app
// installation repository selection event received by a webhook
type GitHubRepositoriesEvent struct {
	Action              string              `json:"action"`
	Installation        *GitHubInstallation `json:"installation"`
	RepositorySelection string              `json:"repository_selection"`
	Sender              *GitHubAccount      `json:"sender"`
}

type GitHubInstallation struct {
	ID      int            `json:"id"`
	Account *GitHubAccount `json:"account"`
}

// GitHubInstalledRepository contains details about a repository
// associated with a GitHub InstallationRepositoriesEvent
type GitHubInstalledRepository struct {
	Action   string `json:"action"` // "added" or "removed"
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

type GitHubAccount struct {
	Login     string `json:"login"`
	ID        string `json:"id"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
	Type      string `json:"type"`
}
