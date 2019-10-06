package types

// CommentObject stores the object that the comment is associated with. Exactly 1 field is nonzero.
//
// TODO!(sqs): it is spaghetti-code that this is in a separate package, but necessary as-is to avoid
// import cycles
type CommentObject struct {
	ParentCommentID int64
	ThreadID        int64
	CampaignID      int64
}
