package graphqlbackend

type ReviewEvent struct {
	EventCommon
	Thread_ ThreadOrIssueOrChangeset
	State_  ReviewState
}

func (v ReviewEvent) Thread() ThreadOrIssueOrChangeset { return v.Thread_ }

func (v ReviewEvent) State() ReviewState { return v.State_ }

type RequestReviewEvent struct {
	EventCommon
	Thread_ ThreadOrIssueOrChangeset
}

func (v RequestReviewEvent) Thread() ThreadOrIssueOrChangeset { return v.Thread_ }
