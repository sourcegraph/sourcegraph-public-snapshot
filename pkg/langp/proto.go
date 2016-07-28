package langp

// Error is returned in the event of any request error, in addition to the HTTP
// status 400 Bad Request.
type Error struct {
	// ErrorMsg, if any, specifies that there was an error serving the request.
	ErrorMsg string `json:"Error"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.ErrorMsg
}

// Position represents a single specific position within a file located in a
// repository at a given revision.
type Position struct {
	// Repo is the repository URI in which the file is located.
	Repo string

	// Commit is the Git commit ID (not branch) of the repository.
	Commit string

	// File is the file which the user is viewing, relative to the repository root.
	File string

	// Line is the line number in the file (zero based), e.g. where a user's cursor
	// is located within the file.
	Line int

	// Character is the character offset on a line in the file (zero based), e.g.
	// where a user's cursor is located within the file.
	Character int
}

// LocalRefs represents references to a specific definition.
type LocalRefs struct {
	// Refs is a list of references to a definition defined within the requested
	// repository.
	Refs []Position
}

// HoverContent represents a subset of the content for when a user “hovers”
// over a definition.
//
// For example, one HoverContent object may represent the comments of a
// function, while the another HoverContent object may represent the function
// signature. In the future we may abuse this field to carry more data, and
// thus we use “type” instead of “language” like in LSP. In practice at this
// point, it always maps to a language (Go, Java, etc).
type HoverContent struct {
	// Type is the type of content (e.g. "Go").
	Type string

	// Value is the value of the content (e.g. "func NewRequest() *Request").
	Value string
}

// Hover represents a message for when a user "hovers" over a definition. It is
// a human-readable description of a definition.
type Hover struct {
	Contents []HoverContent
}
