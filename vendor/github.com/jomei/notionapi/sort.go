package notionapi

type SortOrder string

type TimestampType string

type SortObject struct {
	Property  string        `json:"property,omitempty"`
	Timestamp TimestampType `json:"timestamp,omitempty"`
	Direction SortOrder     `json:"direction,omitempty"`
}
