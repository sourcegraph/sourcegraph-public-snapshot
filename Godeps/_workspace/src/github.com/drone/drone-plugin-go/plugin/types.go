package plugin

// Repo represents a version control repository.
type Repo struct {
	Kind        string `json:"scm"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Avatar      string `json:"avatar_url"`
	Link        string `json:"link_url"`
	Clone       string `json:"clone_url"`
	Branch      string `json:"default_branch"`
	Timeout     int64  `json:"timeout"`
	IsPrivate   bool   `json:"private"`
	IsTrusted   bool   `json:"trusted"`
	AllowPull   bool   `json:"allow_pr"`
	AllowPush   bool   `json:"allow_push"`
	AllowDeploy bool   `json:"allow_deploys"`
	AllowTag    bool   `json:"allow_tags"`
}

// System provides important information about the Drone
// server to the plugin.
type System struct {
	Version string   `json:"version"`
	Link    string   `json:"link_url"`
	Plugins []string `json:"plugins"`
	Globals []string `json:"globals"`
}

// Workspace defines the build's workspace inside the
// container. This helps the plugin locate the source
// code directory.
type Workspace struct {
	Root string `json:"root"`
	Path string `json:"path"`

	Netrc []*NetrcEntry `json:"netrc"`
	Keys  *Keypair      `json:"keys"`
}

// Keypair represents an RSA public and private key assigned to a
// repository. It may be used to clone private repositories, or as
// a deployment key.
type Keypair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

// NetrcEntry defines a "machine" entry in a .netrc file that should
// be injected into the build environment. It will be used to
// authorize access to https resources, such as git+https clones.
type NetrcEntry struct {
	Machine  string `json:"machine"`
	Login    string `json:"login"`
	Password string `json:"user"`
}

// Build represents the process of compiling and testing a changeset,
// typically triggered by the remote system (ie GitHub).
type Build struct {
	Number    int    `json:"number"`
	Event     string `json:"event"`
	Status    string `json:"status"`
	Enqueued  int64  `json:"enqueued_at"`
	Created   int64  `json:"created_at"`
	Started   int64  `json:"started_at"`
	Finished  int64  `json:"finished_at"`
	Deploy    string `json:"deploy_to"`
	Commit    string `json:"commit"`
	Branch    string `json:"branch"`
	Ref       string `json:"ref"`
	Refspec   string `json:"refspec"`
	Remote    string `json:"remote"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Author    string `json:"author"`
	Avatar    string `json:"author_avatar"`
	Email     string `json:"author_email"`
	Link      string `json:"link_url"`
}

// Job represents a single job that is being executed as part
// of a Build.
type Job struct {
	ID       int64  `json:"id"`
	Number   int    `json:"number"`
	Status   string `json:"status"`
	ExitCode int    `json:"exit_code"`
	Enqueued int64  `json:"enqueued_at"`
	Started  int64  `json:"started_at"`
	Finished int64  `json:"finished_at"`

	Environment map[string]string `json:"environment"`
}
