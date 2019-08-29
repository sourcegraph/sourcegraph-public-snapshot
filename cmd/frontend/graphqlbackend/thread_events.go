package graphqlbackend

type CreateThreadEvent struct {
	EventCommon
	Thread_ Thread
}

func (v CreateThreadEvent) Thread() Thread { return v.Thread_ }

type MergeThreadEvent struct {
	EventCommon
	Thread_ Thread
}

func (v MergeThreadEvent) Thread() Thread { return v.Thread_ }

type CloseThreadEvent struct {
	EventCommon
	Thread_ Thread
}

func (v CloseThreadEvent) Thread() Thread { return v.Thread_ }

type ReopenThreadEvent struct {
	EventCommon
	Thread_ Thread
}

func (v ReopenThreadEvent) Thread() Thread { return v.Thread_ }
