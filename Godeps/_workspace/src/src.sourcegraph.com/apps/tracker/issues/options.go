package issues

type IssueListOptions struct {
	State StateFilter
}

type StateFilter State

const (
	AllStates StateFilter = "all"
)
