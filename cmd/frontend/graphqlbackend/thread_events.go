package graphqlbackend

type CreateThreadEvent struct {
	EventCommon
	Thread_ ThreadOrIssueOrChangeset
}

func (v CreateThreadEvent) Thread() ThreadOrIssueOrChangeset { return v.Thread_ }

type MergeChangesetEvent struct {
	EventCommon
	Changeset_ Changeset
}

func (v MergeChangesetEvent) Changeset() Changeset { return v.Changeset_ }

type CloseThreadEvent struct {
	EventCommon
	Thread_ ThreadOrIssueOrChangeset
}

func (v CloseThreadEvent) Thread() ThreadOrIssueOrChangeset { return v.Thread_ }

type ReopenThreadEvent struct {
	EventCommon
	Thread_ ThreadOrIssueOrChangeset
}

func (v ReopenThreadEvent) Thread() ThreadOrIssueOrChangeset { return v.Thread_ }

type CommentOnThreadEvent struct {
	EventCommon
	Thread_ ThreadOrIssueOrChangeset
	// TODO!(sqs): add comment
}

func (v CommentOnThreadEvent) Thread() ThreadOrIssueOrChangeset { return v.Thread_ }
