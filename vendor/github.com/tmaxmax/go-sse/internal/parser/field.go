package parser

// FieldName is the name of the field.
type FieldName string

// A Field represents an unprocessed field of a single event. The Name is the field's identifier, which is used to
// process the fields afterwards.
//
// As a special case, if a parser (FieldParser or Parser) returns a field without a name,
// it means that a whole event was parsed. In other words, all the fields before the one without a name
// and after another such field are part of the same event.
type Field struct {
	Name  FieldName
	Value string
}

// Valid field names.
const (
	FieldNameData  = FieldName("data")
	FieldNameEvent = FieldName("event")
	FieldNameRetry = FieldName("retry")
	FieldNameID    = FieldName("id")
	// FieldNameComment is a sentinel value that indicates
	// comment fields. It is not a valid field name that should
	// be written to a SSE stream.
	FieldNameComment = FieldName(":")

	maxFieldNameLength = 5
)

func getFieldName(b string) (FieldName, bool) {
	switch FieldName(b) { //nolint:exhaustive // Cannot have Comment here
	case FieldNameData:
		return FieldNameData, true
	case FieldNameEvent:
		return FieldNameEvent, true
	case FieldNameRetry:
		return FieldNameRetry, true
	case FieldNameID:
		return FieldNameID, true
	default:
		return "", false
	}
}
