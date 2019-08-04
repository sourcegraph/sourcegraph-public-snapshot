package graphqlbackend

type ReviewEvent struct {
	EventCommon
	Thread_ ThreadOrIssueOrChangeset
	State_  ReviewState
}

func (v ReviewEvent) Thread() ThreadOrIssueOrChangeset { return v.Thread_ }

func (v ReviewEvent) State() ReviewState { return v.State_ }

type ReviewRequestedEvent struct {
	EventCommon
	Thread_ ThreadOrIssueOrChangeset
}

func (v ReviewRequestedEvent) Thread() ThreadOrIssueOrChangeset { return v.Thread_ }
