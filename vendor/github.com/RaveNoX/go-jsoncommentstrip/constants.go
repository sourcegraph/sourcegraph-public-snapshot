package jsoncommentstrip

const (
	commentSingle = "//"

	commentMultiStart = "/*"
	commentMultiEnd   = "*/"

	quoteMark = "\""
)

const (
	stateOther = iota
	stateQuotation
	stateEscape
	stateSComment
	stateMComment
)
