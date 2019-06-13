package graphqlbackend

type ReviewEvent struct {
	EventCommon
	Thread_ Thread
	State_  ReviewState
}

func (v ReviewEvent) Thread() Thread { return v.Thread_ }

func (v ReviewEvent) State() ReviewState { return v.State_ }

type RequestReviewEvent struct {
	EventCommon
	Thread_ Thread
}

func (v RequestReviewEvent) Thread() Thread { return v.Thread_ }
