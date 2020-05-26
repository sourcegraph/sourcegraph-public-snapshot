package query

var empty = struct{}{}

var allFields = map[string]struct{}{
	FieldCase:               empty,
	FieldRepo:               empty,
	"r":                     empty,
	FieldRepoGroup:          empty,
	"g":                     empty,
	FieldFile:               empty,
	"f":                     empty,
	FieldFork:               empty,
	FieldArchived:           empty,
	FieldLang:               empty,
	"l":                     empty,
	"language":              empty,
	FieldType:               empty,
	FieldPatternType:        empty,
	FieldContent:            empty,
	FieldRepoHasFile:        empty,
	FieldRepoHasCommitAfter: empty,
	FieldBefore:             empty,
	"until":                 empty,
	FieldAfter:              empty,
	"since":                 empty,
	FieldAuthor:             empty,
	FieldCommitter:          empty,
	FieldMessage:            empty,
	"m":                     empty,
	"msg":                   empty,
	FieldIndex:              empty,
	FieldCount:              empty,
	FieldStable:             empty,
	FieldMax:                empty,
	FieldTimeout:            empty,
	FieldReplace:            empty,
	FieldCombyRule:          empty,
}
