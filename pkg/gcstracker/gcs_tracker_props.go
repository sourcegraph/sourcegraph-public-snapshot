package gcstracker

// TrackedObjects represents a transmission of data objects to be tracked and stored
type TrackedObjects struct {
	DeviceInfo *DeviceInfo      `json:"device_info,omitempty"`
	Header     *Header          `json:"header,omitempty"`
	Objects    []*TrackedObject `json:"objects,omitempty"`
	UserInfo   *UserInfo        `json:"user_info,omitempty"`
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
	BusinessUserID    string `json:"business_user_id"`
	Email             string `json:"email"`
	IsPrivateCodeUser bool   `json:"is_private_code_user"`
}

// TrackedObject represents a user data object to be tracked and stored
type TrackedObject struct {
	ObjectID string      `json:"object_id,omitempty"`
	Type     string      `json:"type,omitempty"`
	Ctx      interface{} `json:"ctx,omitempty"`
}

// RepoWithDetailsContext is an (ideally) non-code host-specific data structure
// for representing key information about a git repository
type RepoWithDetailsContext struct {
	URI         string          `json:"uri,omitempty"`
	Owner       string          `json:"owner,omitempty"`
	Name        string          `json:"name,omitempty"`
	IsFork      bool            `json:"is_fork,omitempty"`
	IsPrivate   bool            `json:"is_private,omitempty"`
	CreatedAt   int64           `json:"created_at,omitempty"`
	Languages   []*RepoLanguage `json:"languages,omitempty"`
	CommitTimes []int64         `json:"latest_commit_tstamps,omitempty"`
}

type RepoLanguage struct {
	Language string `json:"language,omitempty"`
	Count    int    `json:"count,omitempty"`
}

type OrgWithDetailsContext struct {
	OrgName string       `json:"name,omitempty"`
	Members []*OrgMember `json:"members,omitempty"`
}

type OrgMember struct {
	MemberUserID string `json:"user_id,omitempty"`
}

// UserDetailsContext is a data structure for representing key information
// about a Sourcegraph user that isn't otherwise available on the frontend
//
// These fields are designed to be linked with user events to generate a
// `users` table
type UserDetailsContext struct {
	// ZapAuthCompleted indicates whether the user in question has ever authorized
	// Zap on the command line
	ZapAuthCompleted bool `json:"zap_auth_completed,omitempty"`
}
