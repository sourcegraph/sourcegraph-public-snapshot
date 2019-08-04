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
