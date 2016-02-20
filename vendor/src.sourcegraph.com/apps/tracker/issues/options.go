package issues

// IssueListOptions are options for list operations.
type IssueListOptions struct {
	State StateFilter
}

// StateFilter is a filter by state.
type StateFilter State

const (
	// AllStates is a state filter that includes all issues.
	AllStates StateFilter = "all"
)
