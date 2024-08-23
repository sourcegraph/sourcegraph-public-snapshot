package constants

// ManiphestTaskStatus is the status of a differential task.
type ManiphestTaskStatus string

const (
	// ManiphestTaskStatusAny is any status.
	ManiphestTaskStatusAny ManiphestTaskStatus = "status-any"
	// ManiphestTaskStatusOpen is any task that is open.
	ManiphestTaskStatusOpen ManiphestTaskStatus = "status-open"
	// ManiphestTaskStatusClosed is any task that is closed.
	ManiphestTaskStatusClosed ManiphestTaskStatus = "status-closed"
	// ManiphestTaskStatusResolved is any task that is resolved.
	ManiphestTaskStatusResolved ManiphestTaskStatus = "status-resolved"
	// ManiphestTaskStatusWontFix is any task that is wontfix.
	ManiphestTaskStatusWontFix ManiphestTaskStatus = "status-wontfix"
	// ManiphestTaskStatusInvalid is any task that is invalid.
	ManiphestTaskStatusInvalid ManiphestTaskStatus = "status-invalid"
	// ManiphestTaskStatusSpite is any task that is spite.
	ManiphestTaskStatusSpite ManiphestTaskStatus = "status-spite"
	// ManiphestTaskStatusDuplicate is any task that is duplicated.
	ManiphestTaskStatusDuplicate ManiphestTaskStatus = "status-duplicate"
)

// ManiphestQueryOrder is the order in which query results cna be ordered.
type ManiphestQueryOrder string

const (
	// ManiphestQueryOrderPriority orders results by priority.
	ManiphestQueryOrderPriority ManiphestQueryOrder = "order-priority"
	// ManiphestQueryOrderModified orders results by date modified.
	ManiphestQueryOrderModified ManiphestQueryOrder = "order-modified"
	// ManiphestQueryOrderCreated orders results by date created.
	ManiphestQueryOrderCreated ManiphestQueryOrder = "order-created"
)
