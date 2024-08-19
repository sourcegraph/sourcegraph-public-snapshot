package constants

// ProjectStatus is the status of a project.
type ProjectStatus string

const (
	// ProjectStatusAny is any project status.
	ProjectStatusAny ProjectStatus = "status-any"
	// ProjectStatusOpen is any project that is open.
	ProjectStatusOpen ProjectStatus = "status-open"
	// ProjectStatusClosed is any project that is closed.
	ProjectStatusClosed ProjectStatus = "status-closed"
	// ProjectStatusActive is any project that is active.
	ProjectStatusActive ProjectStatus = "status-active"
	// ProjectStatusArchived is any project that is archived.
	ProjectStatusArchived ProjectStatus = "status-archived"
)
