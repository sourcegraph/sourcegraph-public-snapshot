package entities

// Cursor represents the pagination cursor on many responses.
type Cursor struct {
	Limit  uint64 `json:"limit"`
	After  uint64 `json:"after"`
	Before uint64 `json:"before"`
}